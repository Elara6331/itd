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
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/infinitime/blefs"
	"go.arsenm.dev/itd/api"
	"go.arsenm.dev/lrpc/codec"
	"go.arsenm.dev/lrpc/server"
)

var (
	ErrDFUInvalidFile    = errors.New("provided file is invalid for given upgrade type")
	ErrDFUNotEnoughFiles = errors.New("not enough files provided for given upgrade type")
	ErrDFUInvalidUpgType = errors.New("invalid upgrade type")
)

func startSocket(ctx context.Context, dev *infinitime.Device) error {
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

	srv := server.New()

	itdAPI := &ITD{
		dev: dev,
	}
	err = srv.Register(itdAPI)
	if err != nil {
		return err
	}

	fsAPI := &FS{
		dev: dev,
		fs:  fs,
	}
	err = srv.Register(fsAPI)
	if err != nil {
		return err
	}

	go srv.Serve(ctx, ln, codec.Default)

	// Log socket start
	log.Info().Str("path", k.String("socket.path")).Msg("Started control socket")

	return nil
}

type ITD struct {
	dev *infinitime.Device
}

func (i *ITD) HeartRate(_ *server.Context) (uint8, error) {
	return i.dev.HeartRate()
}

func (i *ITD) WatchHeartRate(ctx *server.Context) error {
	ch, err := ctx.MakeChannel()
	if err != nil {
		return err
	}

	heartRateCh, err := i.dev.WatchHeartRate(ctx)
	if err != nil {
		return err
	}

	go func() {
		// For every heart rate value
		for heartRate := range heartRateCh {
			ch <- heartRate
		}
	}()

	return nil
}

func (i *ITD) BatteryLevel(_ *server.Context) (uint8, error) {
	return i.dev.BatteryLevel()
}

func (i *ITD) WatchBatteryLevel(ctx *server.Context) error {
	ch, err := ctx.MakeChannel()
	if err != nil {
		return err
	}

	battLevelCh, err := i.dev.WatchBatteryLevel(ctx)
	if err != nil {
		return err
	}

	go func() {
		// For every heart rate value
		for battLevel := range battLevelCh {
			ch <- battLevel
		}
	}()

	return nil
}

func (i *ITD) Motion(_ *server.Context) (infinitime.MotionValues, error) {
	return i.dev.Motion()
}

func (i *ITD) WatchMotion(ctx *server.Context) error {
	ch, err := ctx.MakeChannel()
	if err != nil {
		return err
	}

	motionValsCh, err := i.dev.WatchMotion(ctx)
	if err != nil {
		return err
	}

	go func() {
		// For every heart rate value
		for motionVals := range motionValsCh {
			ch <- motionVals
		}
	}()

	return nil
}

func (i *ITD) StepCount(_ *server.Context) (uint32, error) {
	return i.dev.StepCount()
}

func (i *ITD) WatchStepCount(ctx *server.Context) error {
	ch, err := ctx.MakeChannel()
	if err != nil {
		return err
	}

	stepCountCh, err := i.dev.WatchStepCount(ctx)
	if err != nil {
		return err
	}

	go func() {
		// For every heart rate value
		for stepCount := range stepCountCh {
			ch <- stepCount
		}
	}()

	return nil
}

func (i *ITD) Version(_ *server.Context) (string, error) {
	return i.dev.Version()
}

func (i *ITD) Address(_ *server.Context) string {
	return i.dev.Address()
}

func (i *ITD) Notify(_ *server.Context, data api.NotifyData) error {
	return i.dev.Notify(data.Title, data.Body)
}

func (i *ITD) SetTime(_ *server.Context, t *time.Time) error {
	return i.dev.SetTime(*t)
}

func (i *ITD) WeatherUpdate(_ *server.Context) {
	sendWeatherCh <- struct{}{}
}

func (i *ITD) FirmwareUpgrade(ctx *server.Context, reqData api.FwUpgradeData) error {
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

	ch, err := ctx.MakeChannel()
	if err != nil {
		return err
	}

	go func() {
		// For every progress event
		for event := range i.dev.DFU.Progress() {
			ch <- event
		}

		firmwareUpdating = false
		// Send zero object to signal completion
		close(ch)
	}()

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

type FS struct {
	dev *infinitime.Device
	fs  *blefs.FS
}

func (fs *FS) RemoveAll(_ *server.Context, paths []string) error {
	fs.updateFS()
	for _, path := range paths {
		err := fs.fs.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *FS) Remove(_ *server.Context, paths []string) error {
	fs.updateFS()
	for _, path := range paths {
		err := fs.fs.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *FS) Rename(_ *server.Context, paths [2]string) error {
	fs.updateFS()
	return fs.fs.Rename(paths[0], paths[1])
}

func (fs *FS) MkdirAll(_ *server.Context, paths []string) error {
	fs.updateFS()
	for _, path := range paths {
		err := fs.fs.MkdirAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *FS) Mkdir(_ *server.Context, paths []string) error {
	fs.updateFS()
	for _, path := range paths {
		err := fs.fs.Mkdir(path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (fs *FS) ReadDir(_ *server.Context, dir string) ([]api.FileInfo, error) {
	fs.updateFS()

	entries, err := fs.fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var fileInfo []api.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		fileInfo = append(fileInfo, api.FileInfo{
			Name:  info.Name(),
			Size:  info.Size(),
			IsDir: info.IsDir(),
		})
	}

	return fileInfo, nil
}

func (fs *FS) Upload(ctx *server.Context, paths [2]string) error {
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

	ch, err := ctx.MakeChannel()
	if err != nil {
		return err
	}
	go func() {
		// For every progress event
		for sent := range remoteFile.Progress() {
			ch <- api.FSTransferProgress{
				Total: remoteFile.Size(),
				Sent:  sent,
			}
		}

		// Send zero object to signal completion
		close(ch)
	}()

	go func() {
		io.Copy(remoteFile, localFile)
		localFile.Close()
		remoteFile.Close()
	}()

	return nil
}

func (fs *FS) Download(ctx *server.Context, paths [2]string) error {
	fs.updateFS()

	localFile, err := os.Create(paths[0])
	if err != nil {
		return err
	}

	remoteFile, err := fs.fs.Open(paths[1])
	if err != nil {
		return err
	}

	ch, err := ctx.MakeChannel()
	if err != nil {
		return err
	}
	go func() {
		// For every progress event
		for sent := range remoteFile.Progress() {
			ch <- api.FSTransferProgress{
				Total: remoteFile.Size(),
				Sent:  sent,
			}
		}

		// Send zero object to signal completion
		close(ch)
		localFile.Close()
		remoteFile.Close()
	}()

	go io.Copy(localFile, remoteFile)

	return nil
}

func (fs *FS) LoadResources(ctx *server.Context, path string) error {
	resFl, err := os.Open(path)
	if err != nil {
		return err
	}

	progCh, err := infinitime.LoadResources(resFl, fs.fs)
	if err != nil {
		return err
	}

	ch, err := ctx.MakeChannel()
	if err != nil {
		return err
	}

	go func() {
		for evt := range progCh {
			ch <- evt
		}
	}()

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
