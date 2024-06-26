package main

import (
	"context"
	"os"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"go.elara.ws/itd/infinitime"
	"go.elara.ws/itd/internal/fusefs"
	"go.elara.ws/logger/log"
)

func startFUSE(ctx context.Context, wg WaitGroup, dev *infinitime.Device) error {
	// This is where we'll mount the FS
	err := os.MkdirAll(k.String("fuse.mountpoint"), 0o755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Ignore the error because nothing might be mounted on the mountpoint
	_ = fusefs.Unmount(k.String("fuse.mountpoint"))

	root, err := fusefs.BuildRootNode(dev)
	if err != nil {
		log.Error("Building root node failed").
			Err(err).
			Send()
		return err
	}

	server, err := fs.Mount(k.String("fuse.mountpoint"), root, &fs.Options{
		MountOptions: fuse.MountOptions{
			// Set to true to see how the file system works.
			Debug:          false,
			SingleThreaded: true,
		},
	})
	if err != nil {
		log.Error("Mounting failed").
			Str("target", k.String("fuse.mountpoint")).
			Err(err).
			Send()
		return err
	}

	log.Info("Mounted on target").
		Str("target", k.String("fuse.mountpoint")).
		Send()

	fusefs.BuildProperties(dev)

	if err != nil {
		log.Warn("Error getting BLE filesystem").Err(err).Send()
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done("fuse")
		<-ctx.Done()
		server.Unmount()
	}()

	return nil
}
