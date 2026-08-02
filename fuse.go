package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"go.elara.ws/itd/infinitime"
	"go.elara.ws/itd/internal/fusefs"
)

func startFUSE(ctx context.Context, wg WaitGroup, dev *infinitime.Device) error {
	// This is where we'll mount the FS
	err := os.MkdirAll(cfg.Fuse.Mountpoint, 0o755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Ignore the error because nothing might be mounted on the mountpoint
	_ = fusefs.Unmount(cfg.Fuse.Mountpoint)

	root, err := fusefs.BuildRootNode(log, dev)
	if err != nil {
		log.Error("Building root node failed", slog.Any("error", err))
		return err
	}

	server, err := fs.Mount(cfg.Fuse.Mountpoint, root, &fs.Options{
		MountOptions: fuse.MountOptions{
			// Set to true to see how the file system works.
			Debug:          false,
			SingleThreaded: true,
		},
	})
	if err != nil {
		return err
	}

	log.Info("Mounted on target", slog.String("target", cfg.Fuse.Mountpoint))

	wg.Add(1)
	go func() {
		defer wg.Done("fuse")
		<-ctx.Done()
		server.Unmount()
	}()

	return nil
}
