/*
 *	itd uses bluetooth low energy to communicate with InfiniTime devices
 *	Copyright (C) 2021 Arsen Musayelyan
 *
 *	This program is free software: you can redistribute it and/or modify
 *	it under the terms of the GNU General Public License as published by
 *	the Free Software Foundation, either version 3 of the License, or
 *	(at your option) any later version.
 *
 *	This program is distributed in the hope that it will be useful,
 *	but WITHOUT ANY WARRANTY; without even the implied warranty of
 *	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *	GNU General Public License for more details.
 *
 *	You should have received a copy of the GNU General Public License
 *	along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/smallnest/rpcx/server"
	"github.com/smallnest/rpcx/share"
	"github.com/vmihailenco/msgpack/v5"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/infinitime/blefs"
	"go.arsenm.dev/itd/api"
)

// This type signifies an unneeded value.
// A struct{} is used as it takes no space in memory.
// This exists for readability purposes
type none = struct{}

var (
	ErrDFUInvalidFile    = errors.New("provided file is invalid for given upgrade type")
	ErrDFUNotEnoughFiles = errors.New("not enough files provided for given upgrade type")
	ErrDFUInvalidUpgType = errors.New("invalid upgrade type")
	ErrRPCXNoReturnURL   = errors.New("bidirectional requests over gateway require a returnURL field in the metadata")
)

type DoneMap map[string]chan struct{}

func (dm DoneMap) Exists(key string) bool {
	_, ok := dm[key]
	return ok
}

func (dm DoneMap) Done(key string) {
	ch := dm[key]
	ch <- struct{}{}
}

func (dm DoneMap) Create(key string) {
	dm[key] = make(chan struct{}, 1)
}

func (dm DoneMap) Remove(key string) {
	close(dm[key])
	delete(dm, key)
}

var done = DoneMap{}

func startSocket(dev *infinitime.Device) error {
	// Make socket directory if non-existant
	err := os.MkdirAll(filepath.Dir(k.String("socket.path")), 0755)
	if err != nil {
		return err
	}

	// Remove old socket if it exists
	err = os.RemoveAll(k.String("socket.path"))
	if err != nil {
		return err
	}

	// Listen on socket path
	ln, err := net.Listen("unix", k.String("socket.path"))
	if err != nil {
		return err
	}

	fs, err := dev.FS()
	if err != nil {
		log.Warn().Err(err).Msg("Error getting BLE filesystem")
	}

	srv := server.NewServer()

	itdAPI := &ITD{
		dev: dev,
		srv: srv,
	}
	err = srv.Register(itdAPI, "")
	if err != nil {
		return err
	}

	fsAPI := &FS{
		dev: dev,
		fs:  fs,
		srv: srv,
	}
	err = srv.Register(fsAPI, "")
	if err != nil {
		return err
	}

	go srv.ServeListener("tcp", ln)

	// Log socket start
	log.Info().Str("path", k.String("socket.path")).Msg("Started control socket")

	return nil
}

type ITD struct {
	dev *infinitime.Device
	srv *server.Server
}

func (i *ITD) HeartRate(_ context.Context, _ none, out *uint8) error {
	heartRate, err := i.dev.HeartRate()
	*out = heartRate
	return err
}

func (i *ITD) WatchHeartRate(ctx context.Context, _ none, out *string) error {
	// Get client message sender
	msgSender, ok := getMsgSender(ctx, i.srv)
	// If user is using gateway, the client connection will not be available
	if !ok {
		return ErrRPCXNoReturnURL
	}

	heartRateCh, cancel, err := i.dev.WatchHeartRate()
	if err != nil {
		return err
	}

	id := uuid.New().String()
	go func() {
		done.Create(id)
		// For every heart rate value
		for heartRate := range heartRateCh {
			select {
			case <-done[id]:
				// Stop notifications if done signal received
				cancel()
				done.Remove(id)
				return
			default:
				data, err := msgpack.Marshal(heartRate)
				if err != nil {
					log.Error().Err(err).Msg("Error encoding heart rate")
					continue
				}

				// Send response to connection if no done signal received
				msgSender.SendMessage(id, "HeartRateSample", nil, data)
			}
		}
	}()

	*out = id
	return nil
}

func (i *ITD) BatteryLevel(_ context.Context, _ none, out *uint8) error {
	battLevel, err := i.dev.BatteryLevel()
	*out = battLevel
	return err
}

func (i *ITD) WatchBatteryLevel(ctx context.Context, _ none, out *string) error {
	// Get client message sender
	msgSender, ok := getMsgSender(ctx, i.srv)
	// If user is using gateway, the client connection will not be available
	if !ok {
		return ErrRPCXNoReturnURL
	}

	battLevelCh, cancel, err := i.dev.WatchBatteryLevel()
	if err != nil {
		return err
	}

	id := uuid.New().String()
	go func() {
		done.Create(id)
		// For every heart rate value
		for battLevel := range battLevelCh {
			select {
			case <-done[id]:
				// Stop notifications if done signal received
				cancel()
				done.Remove(id)
				return
			default:
				data, err := msgpack.Marshal(battLevel)
				if err != nil {
					log.Error().Err(err).Msg("Error encoding battery level")
					continue
				}

				// Send response to connection if no done signal received
				msgSender.SendMessage(id, "BatteryLevelSample", nil, data)
			}
		}
	}()

	*out = id
	return nil
}

func (i *ITD) Motion(_ context.Context, _ none, out *infinitime.MotionValues) error {
	motionVals, err := i.dev.Motion()
	*out = motionVals
	return err
}

func (i *ITD) WatchMotion(ctx context.Context, _ none, out *string) error {
	// Get client message sender
	msgSender, ok := getMsgSender(ctx, i.srv)
	// If user is using gateway, the client connection will not be available
	if !ok {
		return ErrRPCXNoReturnURL
	}

	motionValsCh, cancel, err := i.dev.WatchMotion()
	if err != nil {
		return err
	}

	id := uuid.New().String()
	go func() {
		done.Create(id)
		// For every heart rate value
		for motionVals := range motionValsCh {
			select {
			case <-done[id]:
				// Stop notifications if done signal received
				cancel()
				done.Remove(id)
				return
			default:
				data, err := msgpack.Marshal(motionVals)
				if err != nil {
					log.Error().Err(err).Msg("Error encoding motion values")
					continue
				}

				// Send response to connection if no done signal received
				msgSender.SendMessage(id, "MotionSample", nil, data)
			}
		}
	}()

	*out = id
	return nil
}

func (i *ITD) StepCount(_ context.Context, _ none, out *uint32) error {
	stepCount, err := i.dev.StepCount()
	*out = stepCount
	return err
}

func (i *ITD) WatchStepCount(ctx context.Context, _ none, out *string) error {
	// Get client message sender
	msgSender, ok := getMsgSender(ctx, i.srv)
	// If user is using gateway, the client connection will not be available
	if !ok {
		return ErrRPCXNoReturnURL
	}

	stepCountCh, cancel, err := i.dev.WatchStepCount()
	if err != nil {
		return err
	}

	id := uuid.New().String()
	go func() {
		done.Create(id)
		// For every heart rate value
		for stepCount := range stepCountCh {
			select {
			case <-done[id]:
				// Stop notifications if done signal received
				cancel()
				done.Remove(id)
				return
			default:
				data, err := msgpack.Marshal(stepCount)
				if err != nil {
					log.Error().Err(err).Msg("Error encoding step count")
					continue
				}

				// Send response to connection if no done signal received
				msgSender.SendMessage(id, "StepCountSample", nil, data)
			}
		}
	}()

	*out = id
	return nil
}

func (i *ITD) Version(_ context.Context, _ none, out *string) error {
	version, err := i.dev.Version()
	*out = version
	return err
}

func (i *ITD) Address(_ context.Context, _ none, out *string) error {
	addr := i.dev.Address()
	*out = addr
	return nil
}

func (i *ITD) Notify(_ context.Context, data api.NotifyData, _ *none) error {
	return i.dev.Notify(data.Title, data.Body)
}

func (i *ITD) SetTime(_ context.Context, t time.Time, _ *none) error {
	return i.dev.SetTime(t)
}

func (i *ITD) WeatherUpdate(_ context.Context, _ none, _ *none) error {
	sendWeatherCh <- struct{}{}
	return nil
}

func (i *ITD) FirmwareUpgrade(ctx context.Context, reqData api.FwUpgradeData, out *string) error {
	i.dev.DFU.Reset()

	switch reqData.Type {
	case api.UpgradeTypeArchive:
		// If less than one file, return error
		if len(reqData.Files) < 1 {
			return ErrDFUNotEnoughFiles
		}
		// If file is not zip archive, return error
		if filepath.Ext(reqData.Files[0]) != ".zip" {
			return ErrDFUInvalidFile
		}
		// Load DFU archive
		err := i.dev.DFU.LoadArchive(reqData.Files[0])
		if err != nil {
			return err
		}
	case api.UpgradeTypeFiles:
		// If less than two files, return error
		if len(reqData.Files) < 2 {
			return ErrDFUNotEnoughFiles
		}
		// If first file is not init packet, return error
		if filepath.Ext(reqData.Files[0]) != ".dat" {
			return ErrDFUInvalidFile
		}
		// If second file is not firmware image, return error
		if filepath.Ext(reqData.Files[1]) != ".bin" {
			return ErrDFUInvalidFile
		}
		// Load individual DFU files
		err := i.dev.DFU.LoadFiles(reqData.Files[0], reqData.Files[1])
		if err != nil {
			return err
		}
	default:
		return ErrDFUInvalidUpgType
	}

	id := uuid.New().String()
	*out = id

	// Get client message sender
	msgSender, ok := getMsgSender(ctx, i.srv)
	// If user is using gateway, the client connection will not be available
	if ok {
		go func() {
			// For every progress event
			for event := range i.dev.DFU.Progress() {
				data, err := msgpack.Marshal(event)
				if err != nil {
					log.Error().Err(err).Msg("Error encoding DFU progress event")
					continue
				}

				msgSender.SendMessage(id, "DFUProgress", nil, data)
			}

			firmwareUpdating = false
			msgSender.SendMessage(id, "Done", nil, nil)
		}()
	}

	// Set firmwareUpdating
	firmwareUpdating = true

	go func() {
		// Start DFU
		err := i.dev.DFU.Start()
		if err != nil {
			log.Error().Err(err).Msg("Error while upgrading firmware")
			firmwareUpdating = false
			return
		}
	}()

	return nil
}

func (i *ITD) Done(_ context.Context, id string, _ *none) error {
	done.Done(id)
	return nil
}

type FS struct {
	dev *infinitime.Device
	fs  *blefs.FS
	srv *server.Server
}

func (fs *FS) Remove(_ context.Context, paths []string, _ *none) error {
	fs.updateFS()
	for _, path := range paths {
		err := fs.fs.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *FS) Rename(_ context.Context, paths [2]string, _ *none) error {
	fs.updateFS()
	return fs.fs.Rename(paths[0], paths[1])
}

func (fs *FS) Mkdir(_ context.Context, paths []string, _ *none) error {
	fs.updateFS()
	for _, path := range paths {
		err := fs.fs.Mkdir(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *FS) ReadDir(_ context.Context, dir string, out *[]api.FileInfo) error {
	fs.updateFS()

	entries, err := fs.fs.ReadDir(dir)
	if err != nil {
		return err
	}
	var fileInfo []api.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return err
		}
		fileInfo = append(fileInfo, api.FileInfo{
			Name:  info.Name(),
			Size:  info.Size(),
			IsDir: info.IsDir(),
		})
	}

	*out = fileInfo
	return nil
}

func (fs *FS) Upload(ctx context.Context, paths [2]string, out *string) error {
	fs.updateFS()

	localFile, err := os.Open(paths[1])
	if err != nil {
		return err
	}

	localInfo, err := localFile.Stat()
	if err != nil {
		return err
	}

	remoteFile, err := fs.fs.Create(paths[0], uint32(localInfo.Size()))
	if err != nil {
		return err
	}

	id := uuid.New().String()
	*out = id

	// Get client message sender
	msgSender, ok := getMsgSender(ctx, fs.srv)
	// If user is using gateway, the client connection will not be available
	if ok {
		go func() {
			// For every progress event
			for sent := range remoteFile.Progress() {
				data, err := msgpack.Marshal(api.FSTransferProgress{
					Total: remoteFile.Size(),
					Sent:  sent,
				})
				if err != nil {
					log.Error().Err(err).Msg("Error encoding filesystem transfer progress event")
					continue
				}

				msgSender.SendMessage(id, "FSProgress", nil, data)
			}

			msgSender.SendMessage(id, "Done", nil, nil)
		}()
	}

	go func() {
		io.Copy(remoteFile, localFile)
		localFile.Close()
		remoteFile.Close()
	}()

	return nil
}

func (fs *FS) Download(ctx context.Context, paths [2]string, out *string) error {
	fs.updateFS()

	localFile, err := os.Create(paths[0])
	if err != nil {
		return err
	}

	remoteFile, err := fs.fs.Open(paths[1])
	if err != nil {
		return err
	}

	id := uuid.New().String()
	*out = id

	// Get client message sender
	msgSender, ok := getMsgSender(ctx, fs.srv)
	// If user is using gateway, the client connection will not be available
	if ok {
		go func() {
			// For every progress event
			for rcvd := range remoteFile.Progress() {
				data, err := msgpack.Marshal(api.FSTransferProgress{
					Total: remoteFile.Size(),
					Sent:  rcvd,
				})
				if err != nil {
					log.Error().Err(err).Msg("Error encoding filesystem transfer progress event")
					continue
				}

				msgSender.SendMessage(id, "FSProgress", nil, data)
			}

			msgSender.SendMessage(id, "Done", nil, nil)
			localFile.Close()
			remoteFile.Close()
		}()
	}

	go io.Copy(localFile, remoteFile)

	return nil
}

func (fs *FS) updateFS() {
	if fs.fs == nil || updateFS {
		// Get new FS
		newFS, err := fs.dev.FS()
		if err != nil {
			log.Warn().Err(err).Msg("Error updating BLE filesystem")
		} else {
			// Set FS pointer to new FS
			fs.fs = newFS
			// Reset updateFS
			updateFS = false
		}
	}
}

// cleanPaths runs strings.TrimSpace and filepath.Clean
// on all inputs, and returns the updated slice
func cleanPaths(paths []string) []string {
	for index, path := range paths {
		newPath := strings.TrimSpace(path)
		paths[index] = filepath.Clean(newPath)
	}
	return paths
}

func getMsgSender(ctx context.Context, srv *server.Server) (MessageSender, bool) {
	// Get client message sender
	clientConn, ok := ctx.Value(server.RemoteConnContextKey).(net.Conn)
	// If the connection exists, use rpcMsgSender
	if ok {
		return &rpcMsgSender{srv, clientConn}, true
	} else {
		// Get metadata if it exists
		metadata, ok := ctx.Value(share.ReqMetaDataKey).(map[string]string)
		if !ok {
			return nil, false
		}
		// Get returnURL field from metadata if it exists
		returnURL, ok := metadata["returnURL"]
		if !ok {
			return nil, false
		}
		// Use httpMsgSender
		return &httpMsgSender{returnURL}, true
	}

}

// The MessageSender interface sends messages to the client
type MessageSender interface {
	SendMessage(servicePath, serviceMethod string, metadata map[string]string, data []byte) error
}

// rpcMsgSender sends messages using RPCX, for clients that support it
type rpcMsgSender struct {
	srv  *server.Server
	conn net.Conn
}

// SendMessage uses the server to send an RPCX message back to the client
func (r *rpcMsgSender) SendMessage(servicePath, serviceMethod string, metadata map[string]string, data []byte) error {
	return r.srv.SendMessage(r.conn, servicePath, serviceMethod, metadata, data)
}

// httpMsgSender sends messages to the given return URL, for clients that provide it
type httpMsgSender struct {
	url string
}

// SendMessage uses HTTP to send a message back to the client
func (h *httpMsgSender) SendMessage(servicePath, serviceMethod string, metadata map[string]string, data []byte) error {
	// Create new POST request with provided URL
	req, err := http.NewRequest(http.MethodPost, h.url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	// Set service path and method headers
	req.Header.Set("X-RPCX-ServicePath", servicePath)
	req.Header.Set("X-RPCX-ServiceMethod", serviceMethod)

	// Create new URL query values
	query := url.Values{}
	// Transfer values from metadata to query
	for k, v := range metadata {
		query.Set(k, v)
	}
	// Set metadata header by encoding query values
	req.Header.Set("X-RPCX-Meta", query.Encode())

	// Perform request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	// Close body
	return res.Body.Close()
}
