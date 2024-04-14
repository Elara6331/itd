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
	"archive/zip"
	"context"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"go.elara.ws/drpc/muxserver"
	"go.elara.ws/itd/infinitime"
	"go.elara.ws/itd/internal/rpc"
	"go.elara.ws/logger/log"
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

	fs := dev.FS()
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
	errCh := make(chan error)

	err := i.dev.WatchHeartRate(s.Context(), func(rate uint8, err error) {
		if err != nil {
			errCh <- err
			return
		}

		err = s.Send(&rpc.IntResponse{Value: uint32(rate)})
		if err != nil {
			errCh <- err
		}
	})
	if err != nil {
		return err
	}

	select {
	case <-errCh:
		return err
	case <-s.Context().Done():
		return nil
	}
}

func (i *ITD) BatteryLevel(_ context.Context, _ *rpc.Empty) (*rpc.IntResponse, error) {
	bl, err := i.dev.BatteryLevel()
	return &rpc.IntResponse{Value: uint32(bl)}, err
}

func (i *ITD) WatchBatteryLevel(_ *rpc.Empty, s rpc.DRPCITD_WatchBatteryLevelStream) error {
	errCh := make(chan error)

	err := i.dev.WatchBatteryLevel(s.Context(), func(level uint8, err error) {
		if err != nil {
			errCh <- err
			return
		}

		err = s.Send(&rpc.IntResponse{Value: uint32(level)})
		if err != nil {
			errCh <- err
		}
	})
	if err != nil {
		return err
	}

	select {
	case <-errCh:
		return err
	case <-s.Context().Done():
		return nil
	}
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
	errCh := make(chan error)

	err := i.dev.WatchMotion(s.Context(), func(motion infinitime.MotionValues, err error) {
		if err != nil {
			errCh <- err
			return
		}

		err = s.Send(&rpc.MotionResponse{
			X: int32(motion.X),
			Y: int32(motion.Y),
			Z: int32(motion.Z),
		})
		if err != nil {
			errCh <- err
		}
	})
	if err != nil {
		return err
	}

	select {
	case <-errCh:
		return err
	case <-s.Context().Done():
		return nil
	}
}

func (i *ITD) StepCount(_ context.Context, _ *rpc.Empty) (*rpc.IntResponse, error) {
	sc, err := i.dev.StepCount()
	return &rpc.IntResponse{Value: sc}, err
}

func (i *ITD) WatchStepCount(_ *rpc.Empty, s rpc.DRPCITD_WatchStepCountStream) error {
	errCh := make(chan error)

	err := i.dev.WatchStepCount(s.Context(), func(count uint32, err error) {
		if err != nil {
			errCh <- err
			return
		}

		err = s.Send(&rpc.IntResponse{Value: count})
		if err != nil {
			errCh <- err
		}
	})
	if err != nil {
		return err
	}

	select {
	case <-errCh:
		return err
	case <-s.Context().Done():
		return nil
	}
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

func (i *ITD) FirmwareUpgrade(data *rpc.FirmwareUpgradeRequest, s rpc.DRPCITD_FirmwareUpgradeStream) (err error) {
	var fwimg, initpkt *os.File

	switch data.Type {
	case rpc.FirmwareUpgradeRequest_Archive:
		fwimg, initpkt, err = extractDFU(data.Files[0])
		if err != nil {
			return err
		}
	case rpc.FirmwareUpgradeRequest_Files:
		if len(data.Files) < 2 {
			return ErrDFUNotEnoughFiles
		}

		if filepath.Ext(data.Files[0]) != ".dat" {
			return ErrDFUInvalidFile
		}

		if filepath.Ext(data.Files[1]) != ".bin" {
			return ErrDFUInvalidFile
		}

		initpkt, err = os.Open(data.Files[0])
		if err != nil {
			return err
		}

		fwimg, err = os.Open(data.Files[1])
		if err != nil {
			return err
		}
	default:
		return ErrDFUInvalidUpgType
	}

	defer os.Remove(fwimg.Name())
	defer os.Remove(initpkt.Name())
	defer fwimg.Close()
	defer initpkt.Close()

	firmwareUpdating = true
	defer func() { firmwareUpdating = false }()

	return i.dev.UpgradeFirmware(infinitime.DFUOptions{
		InitPacket:    initpkt,
		FirmwareImage: fwimg,
		ProgressFunc: func(sent, received, total uint32) {
			_ = s.Send(&rpc.DFUProgress{
				Sent:     int64(sent),
				Recieved: int64(received),
				Total:    int64(total),
			})
		},
	})
}

type FS struct {
	dev *infinitime.Device
	fs  *infinitime.FS
}

func (fs *FS) RemoveAll(_ context.Context, req *rpc.PathsRequest) (*rpc.Empty, error) {
	for _, path := range req.Paths {
		err := fs.fs.RemoveAll(path)
		if err != nil {
			return &rpc.Empty{}, err
		}
	}
	return &rpc.Empty{}, nil
}

func (fs *FS) Remove(_ context.Context, req *rpc.PathsRequest) (*rpc.Empty, error) {
	for _, path := range req.Paths {
		err := fs.fs.Remove(path)
		if err != nil {
			return &rpc.Empty{}, err
		}
	}
	return &rpc.Empty{}, nil
}

func (fs *FS) Rename(_ context.Context, req *rpc.RenameRequest) (*rpc.Empty, error) {
	return &rpc.Empty{}, fs.fs.Rename(req.From, req.To)
}

func (fs *FS) MkdirAll(_ context.Context, req *rpc.PathsRequest) (*rpc.Empty, error) {
	for _, path := range req.Paths {
		err := fs.fs.MkdirAll(path)
		if err != nil {
			return &rpc.Empty{}, err
		}
	}
	return &rpc.Empty{}, nil
}

func (fs *FS) Mkdir(_ context.Context, req *rpc.PathsRequest) (*rpc.Empty, error) {
	for _, path := range req.Paths {
		err := fs.fs.Mkdir(path)
		if err != nil {
			return &rpc.Empty{}, err
		}
	}
	return &rpc.Empty{}, nil
}

func (fs *FS) ReadDir(_ context.Context, req *rpc.PathRequest) (*rpc.DirResponse, error) {
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

	remoteFile.ProgressFunc = func(transferred, total uint32) {
		_ = s.Send(&rpc.TransferProgress{
			Total: total,
			Sent:  transferred,
		})
	}

	io.Copy(remoteFile, localFile)
	localFile.Close()
	remoteFile.Close()

	return nil
}

func (fs *FS) Download(req *rpc.TransferRequest, s rpc.DRPCFS_DownloadStream) error {
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

	remoteFile.ProgressFunc = func(transferred, total uint32) {
		_ = s.Send(&rpc.TransferProgress{
			Total: total,
			Sent:  transferred,
		})
	}

	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FS) LoadResources(req *rpc.PathRequest, s rpc.DRPCFS_LoadResourcesStream) error {
	return infinitime.LoadResources(req.Path, fs.fs, func(evt infinitime.ResourceLoadProgress) {
		_ = s.Send(&rpc.ResourceLoadProgress{
			Name:      evt.Name,
			Total:     int64(evt.Total),
			Sent:      int64(evt.Transferred),
			Operation: rpc.ResourceLoadProgress_Operation(evt.Operation),
		})
	})
}

func extractDFU(path string) (fwimg, initpkt *os.File, err error) {
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return nil, nil, err
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		if fwimg != nil && initpkt != nil {
			break
		}

		switch filepath.Ext(file.Name) {
		case ".bin":
			fwimg, err = os.CreateTemp(os.TempDir(), "itd_dfu_fwimg_*.bin")
			if err != nil {
				return nil, nil, err
			}

			zipFile, err := file.Open()
			if err != nil {
				return nil, nil, err
			}
			defer zipFile.Close()

			_, err = io.Copy(fwimg, zipFile)
			if err != nil {
				return nil, nil, err
			}

			err = zipFile.Close()
			if err != nil {
				return nil, nil, err
			}

			_, err = fwimg.Seek(0, io.SeekStart)
			if err != nil {
				return nil, nil, err
			}
		case ".dat":
			initpkt, err = os.CreateTemp(os.TempDir(), "itd_dfu_initpkt_*.dat")
			if err != nil {
				return nil, nil, err
			}

			zipFile, err := file.Open()
			if err != nil {
				return nil, nil, err
			}

			_, err = io.Copy(initpkt, zipFile)
			if err != nil {
				return nil, nil, err
			}

			err = zipFile.Close()
			if err != nil {
				return nil, nil, err
			}

			_, err = initpkt.Seek(0, io.SeekStart)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	if fwimg == nil || initpkt == nil {
		return nil, nil, errors.New("invalid dfu archive")
	}

	return fwimg, initpkt, nil
}
