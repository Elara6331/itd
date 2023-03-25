package fusefs

import (
	_ "unsafe"

	"github.com/hanwen/go-fuse/v2/fuse"
)

func Unmount(mountPoint string) error {
	return unmount(mountPoint, &fuse.MountOptions{DirectMount: false})
}

// Unfortunately, the FUSE library does not export its unmount function,
// so this is required until that changes
//
//go:linkname unmount github.com/hanwen/go-fuse/v2/fuse.unmount
func unmount(mountPoint string, opts *fuse.MountOptions) error
