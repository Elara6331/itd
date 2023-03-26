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

	"go.arsenm.dev/drpc/muxserver"
	"go.arsenm.dev/infinitime"
	"go.arsenm.dev/infinitime/blefs"
	"go.arsenm.dev/itd/internal/rpc"
	"go.arsenm.dev/logger/log"
	"storj.io/drpc/drpcmux"
)

var (
	ErrDFUInvalidFile    = errors.New("provided file is invalid for given upgrade type")
	ErrDFUNotEnoughFiles = errors.New("not enough files provided for given upgrade type")
	ErrDFUInvalidUpgType = errors.New("invalid upgrade type")
)

func startSocket(ctx context.Context, wg WaitGroup, dev *infinitime.Device) error {
	// Make socket directory if non-existant
	err := os.MkdirAll(filepath.Dir(k.String("socket.path")), 0o755)
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
		log.Warn("Error getting BLE filesystem").Err(err).Send()
	}

	mux := drpcmux.New()

	err = rpc.DRPCRegisterITD(mux, &ITD{dev})
	if err != nil {
		return err
	}

	err = rpc.DRPCRegisterFS(mux, &FS{dev, fs})
	if err != nil {
		return err
	}

	log.Info("Starting control socket").Str("path", k.String("socket.path")).Send()

	wg.Add(1)
	go func() {
		defer wg.Done("socket")
		muxserver.New(mux).Serve(ctx, ln)
	}()

	return nil
}

type ITD struct {
	dev *infinitime.Device
}

func (i *ITD) HeartRate(_ context.Context, _ *rpc.Empty) (*rpc.IntResponse, error) {
	hr, err := i.dev.HeartRate()
	return &rpc.IntResponse{Value: uint32(hr)}, err
}

func (i *ITD) WatchHeartRate(_ *rpc.Empty, s rpc.DRPCITD_WatchHeartRateStream) error {
	heartRateCh, err := i.dev.WatchHeartRate(s.Context())
	if err != nil {
		return err
	}

	for heartRate := range heartRateCh {
		err = s.Send(&rpc.IntResponse{Value: uint32(heartRate)})
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *ITD) BatteryLevel(_ context.Context, _ *rpc.Empty) (*rpc.IntResponse, error) {
	bl, err := i.dev.BatteryLevel()
	return &rpc.IntResponse{Value: uint32(bl)}, err
}

func (i *ITD) WatchBatteryLevel(_ *rpc.Empty, s rpc.DRPCITD_WatchBatteryLevelStream) error {
	battLevelCh, err := i.dev.WatchBatteryLevel(s.Context())
	if err != nil {
		return err
	}

	for battLevel := range battLevelCh {
		err = s.Send(&rpc.IntResponse{Value: uint32(battLevel)})
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *ITD) Motion(_ context.Context, _ *rpc.Empty) (*rpc.MotionResponse, error) {
	motionVals, err := i.dev.Motion()
	return &rpc.MotionResponse{
		X: int32(motionVals.X),
		Y: int32(motionVals.Y),
		Z: int32(motionVals.Z),
	}, err
}

func (i *ITD) WatchMotion(_ *rpc.Empty, s rpc.DRPCITD_WatchMotionStream) error {
	motionValsCh, err := i.dev.WatchMotion(s.Context())
	if err != nil {
		return err
	}

	for motionVals := range motionValsCh {
		err = s.Send(&rpc.MotionResponse{
			X: int32(motionVals.X),
			Y: int32(motionVals.Y),
			Z: int32(motionVals.Z),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *ITD) StepCount(_ context.Context, _ *rpc.Empty) (*rpc.IntResponse, error) {
	sc, err := i.dev.StepCount()
	return &rpc.IntResponse{Value: sc}, err
}

func (i *ITD) WatchStepCount(_ *rpc.Empty, s rpc.DRPCITD_WatchStepCountStream) error {
	stepCountCh, err := i.dev.WatchStepCount(s.Context())
	if err != nil {
		return err
	}

	for stepCount := range stepCountCh {
		err = s.Send(&rpc.IntResponse{Value: stepCount})
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *ITD) Version(_ context.Context, _ *rpc.Empty) (*rpc.StringResponse, error) {
	v, err := i.dev.Version()
	return &rpc.StringResponse{Value: v}, err
}

func (i *ITD) Address(_ context.Context, _ *rpc.Empty) (*rpc.StringResponse, error) {
	return &rpc.StringResponse{Value: i.dev.Address()}, nil
}

func (i *ITD) Notify(_ context.Context, data *rpc.NotifyRequest) (*rpc.Empty, error) {
	return &rpc.Empty{}, i.dev.Notify(data.Title, data.Body)
}

func (i *ITD) SetTime(_ context.Context, data *rpc.SetTimeRequest) (*rpc.Empty, error) {
	return &rpc.Empty{}, i.dev.SetTime(time.Unix(0, data.UnixNano))
}

func (i *ITD) WeatherUpdate(context.Context, *rpc.Empty) (*rpc.Empty, error) {
	sendWeatherCh <- struct{}{}
	return &rpc.Empty{}, nil
}

func (i *ITD) FirmwareUpgrade(data *rpc.FirmwareUpgradeRequest, s rpc.DRPCITD_FirmwareUpgradeStream) error {
	i.dev.DFU.Reset()

	switch data.Type {
	case rpc.FirmwareUpgradeRequest_Archive:
		// If less than one file, return error
		if len(data.Files) < 1 {
			return ErrDFUNotEnoughFiles
		}
		// If file is not zip archive, return error
		if filepath.Ext(data.Files[0]) != ".zip" {
			return ErrDFUInvalidFile
		}
		// Load DFU archive
		err := i.dev.DFU.LoadArchive(data.Files[0])
		if err != nil {
			return err
		}
	case rpc.FirmwareUpgradeRequest_Files:
		// If less than two files, return error
		if len(data.Files) < 2 {
			return ErrDFUNotEnoughFiles
		}
		// If first file is not init packet, return error
		if filepath.Ext(data.Files[0]) != ".dat" {
			return ErrDFUInvalidFile
		}
		// If second file is not firmware image, return error
		if filepath.Ext(data.Files[1]) != ".bin" {
			return ErrDFUInvalidFile
		}
		// Load individual DFU files
		err := i.dev.DFU.LoadFiles(data.Files[0], data.Files[1])
		if err != nil {
			return err
		}
	default:
		return ErrDFUInvalidUpgType
	}

	go func() {
		for event := range i.dev.DFU.Progress() {
			_ = s.Send(&rpc.DFUProgress{
				Sent:     int64(event.Sent),
				Recieved: int64(event.Received),
				Total:    event.Total,
			})
		}

		firmwareUpdating = false
	}()

	// Set firmwareUpdating
	firmwareUpdating = true

	// Start DFU
	err := i.dev.DFU.Start()
	if err != nil {
		firmwareUpdating = false
		return err
	}

	return nil
}

type FS struct {
	dev *infinitime.Device
	fs  *blefs.FS
}

func (fs *FS) RemoveAll(_ context.Context, req *rpc.PathsRequest) (*rpc.Empty, error) {
	fs.updateFS()
	for _, path := range req.Paths {
		err := fs.fs.RemoveAll(path)
		if err != nil {
			return &rpc.Empty{}, err
		}
	}
	return &rpc.Empty{}, nil
}

func (fs *FS) Remove(_ context.Context, req *rpc.PathsRequest) (*rpc.Empty, error) {
	fs.updateFS()
	for _, path := range req.Paths {
		err := fs.fs.Remove(path)
		if err != nil {
			return &rpc.Empty{}, err
		}
	}
	return &rpc.Empty{}, nil
}

func (fs *FS) Rename(_ context.Context, req *rpc.RenameRequest) (*rpc.Empty, error) {
	fs.updateFS()
	return &rpc.Empty{}, fs.fs.Rename(req.From, req.To)
}

func (fs *FS) MkdirAll(_ context.Context, req *rpc.PathsRequest) (*rpc.Empty, error) {
	fs.updateFS()
	for _, path := range req.Paths {
		err := fs.fs.MkdirAll(path)
		if err != nil {
			return &rpc.Empty{}, err
		}
	}
	return &rpc.Empty{}, nil
}

func (fs *FS) Mkdir(_ context.Context, req *rpc.PathsRequest) (*rpc.Empty, error) {
	fs.updateFS()
	for _, path := range req.Paths {
		err := fs.fs.Mkdir(path)
		if err != nil {
			return &rpc.Empty{}, err
		}
	}
	return &rpc.Empty{}, nil
}

func (fs *FS) ReadDir(_ context.Context, req *rpc.PathRequest) (*rpc.DirResponse, error) {
	fs.updateFS()

	entries, err := fs.fs.ReadDir(req.Path)
	if err != nil {
		return nil, err
	}
	var fileInfo []*rpc.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		fileInfo = append(fileInfo, &rpc.FileInfo{
			Name:  info.Name(),
			Size:  info.Size(),
			IsDir: info.IsDir(),
		})
	}

	return &rpc.DirResponse{Entries: fileInfo}, nil
}

func (fs *FS) Upload(req *rpc.TransferRequest, s rpc.DRPCFS_UploadStream) error {
	fs.updateFS()

	localFile, err := os.Open(req.Source)
	if err != nil {
		return err
	}

	localInfo, err := localFile.Stat()
	if err != nil {
		return err
	}

	remoteFile, err := fs.fs.Create(req.Destination, uint32(localInfo.Size()))
	if err != nil {
		return err
	}

	go func() {
		// For every progress event
		for sent := range remoteFile.Progress() {
			_ = s.Send(&rpc.TransferProgress{
				Total: remoteFile.Size(),
				Sent:  sent,
			})
		}
	}()

	io.Copy(remoteFile, localFile)
	localFile.Close()
	remoteFile.Close()

	return nil
}

func (fs *FS) Download(req *rpc.TransferRequest, s rpc.DRPCFS_DownloadStream) error {
	fs.updateFS()

	localFile, err := os.Create(req.Destination)
	if err != nil {
		return err
	}

	remoteFile, err := fs.fs.Open(req.Source)
	if err != nil {
		return err
	}

	defer localFile.Close()
	defer remoteFile.Close()

	go func() {
		// For every progress event
		for sent := range remoteFile.Progress() {
			_ = s.Send(&rpc.TransferProgress{
				Total: remoteFile.Size(),
				Sent:  sent,
			})
		}
	}()

	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FS) LoadResources(req *rpc.PathRequest, s rpc.DRPCFS_LoadResourcesStream) error {
	resFl, err := os.Open(req.Path)
	if err != nil {
		return err
	}

	progCh, err := infinitime.LoadResources(resFl, fs.fs)
	if err != nil {
		return err
	}

	for evt := range progCh {
		err = s.Send(&rpc.ResourceLoadProgress{
			Name:      evt.Name,
			Total:     evt.Total,
			Sent:      evt.Sent,
			Operation: rpc.ResourceLoadProgress_Operation(evt.Operation),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (fs *FS) updateFS() {
	if fs.fs == nil || updateFS {
		// Get new FS
		newFS, err := fs.dev.FS()
		if err != nil {
			log.Warn("Error updating BLE filesystem").Err(err).Send()
		} else {
			// Set FS pointer to new FS
			fs.fs = newFS
			// Reset updateFS
			updateFS = false
		}
	}
}
