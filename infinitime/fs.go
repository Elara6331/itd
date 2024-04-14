package infinitime

import (
	"errors"
	"io"
	"io/fs"
	"math"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"go.elara.ws/itd/internal/fsproto"
	"tinygo.org/x/bluetooth"
)

// FS represents a remote BLE filesystem
type FS struct {
	mtx sync.Mutex
	dev *Device
}

// Stat gets information about a file at the given path.
//
// WARNING: Since there's no stat command in the BLE FS protocol,
// this function does a ReadDir and then finds the requested file
// in the results, which makes it pretty slow.
func (ifs *FS) Stat(p string) (fs.FileInfo, error) {
	dir := path.Dir(p)
	entries, err := ifs.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.Name() == path.Base(p) {
			return entry.Info()
		}
	}

	return nil, fsproto.ErrFileNotExists
}

// Remove removes a file or empty directory at the given path.
//
// For a function that removes directories recursively, see [FS.RemoveAll]
func (ifs *FS) Remove(path string) error {
	ifs.mtx.Lock()
	defer ifs.mtx.Unlock()

	char, err := ifs.dev.getChar(fsTransferChar)
	if err != nil {
		return err
	}

	return ifs.requestThenAwaitResponse(
		char,
		fsproto.DeleteFileOpcode,
		fsproto.DeleteFileRequest{
			PathLen: uint16(len(path)),
			Path:    path,
		},
		func(buf []byte) (bool, error) {
			var mdr fsproto.DeleteFileResponse
			return true, fsproto.ReadResponse(buf, fsproto.DeleteFileResp, &mdr)
		},
	)
}

// Rename moves a file or directory from an old path to a new path.
func (ifs *FS) Rename(old, new string) error {
	ifs.mtx.Lock()
	defer ifs.mtx.Unlock()

	char, err := ifs.dev.getChar(fsTransferChar)
	if err != nil {
		return err
	}

	return ifs.requestThenAwaitResponse(
		char,
		fsproto.MoveFileOpcode,
		fsproto.MoveFileRequest{
			OldPathLen: uint16(len(old)),
			OldPath:    old,
			NewPathLen: uint16(len(new)),
			NewPath:    new,
		},
		func(buf []byte) (bool, error) {
			var mfr fsproto.MoveFileResponse
			return true, fsproto.ReadResponse(buf, fsproto.MoveFileResp, &mfr)
		},
	)
}

// Mkdir creates a new directory at the specified path.
//
// For a function that creates necessary parents as well, see [FS.MkdirAll]
func (ifs *FS) Mkdir(path string) error {
	ifs.mtx.Lock()
	defer ifs.mtx.Unlock()

	char, err := ifs.dev.getChar(fsTransferChar)
	if err != nil {
		return err
	}

	return ifs.requestThenAwaitResponse(
		char,
		fsproto.MakeDirectoryOpcode,
		fsproto.MkdirRequest{
			PathLen: uint16(len(path)),
			Path:    path,
		},
		func(buf []byte) (bool, error) {
			var mdr fsproto.MkdirResponse
			return true, fsproto.ReadResponse(buf, fsproto.MakeDirectoryResp, &mdr)
		},
	)
}

// ReadDir reads the directory at the specified path and returns a list of directory entries.
func (ifs *FS) ReadDir(path string) ([]fs.DirEntry, error) {
	ifs.mtx.Lock()
	defer ifs.mtx.Unlock()

	char, err := ifs.dev.getChar(fsTransferChar)
	if err != nil {
		return nil, err
	}

	var out []fs.DirEntry
	return out, ifs.requestThenAwaitResponse(
		char,
		fsproto.ListDirectoryOpcode,
		fsproto.ListDirRequest{
			PathLen: uint16(len(path)),
			Path:    path,
		},
		func(buf []byte) (bool, error) {
			var ldr fsproto.ListDirResponse
			err := fsproto.ReadResponse(buf, fsproto.ListDirectoryResp, &ldr)
			if err != nil {
				return true, err
			}

			if ldr.EntryNum == ldr.TotalEntries {
				return true, nil
			}

			out = append(out, DirEntry{
				flags:   ldr.Flags,
				modtime: ldr.ModTime,
				size:    ldr.FileSize,
				path:    string(ldr.Path),
			})

			return false, nil
		},
	)
}

// RemoveAll removes the file at the specified path and any children it contains,
// similar to the rm -r command.
func (ifs *FS) RemoveAll(p string) error {
	if p == "" {
		return nil
	}

	if path.Clean(p) == "/" {
		return fsproto.ErrNoRemoveRoot
	}

	fi, err := ifs.Stat(p)
	if err != nil {
		return nil
	}

	if fi.IsDir() {
		return ifs.removeWithChildren(p)
	} else {
		err = ifs.Remove(p)

		var code int8
		if err, ok := err.(fsproto.Error); ok {
			code = err.Code
		}

		if err != nil && code != -2 {
			return err
		}
	}

	return nil
}

// removeWithChildren removes the directory at the given path and its children recursively.
func (ifs *FS) removeWithChildren(p string) error {
	list, err := ifs.ReadDir(p)
	if err != nil {
		return err
	}

	for _, entry := range list {
		name := entry.Name()

		if name == "." || name == ".." {
			continue
		}
		entryPath := path.Join(p, name)

		if entry.IsDir() {
			err = ifs.removeWithChildren(entryPath)
		} else {
			err = ifs.Remove(entryPath)
		}

		var code int8
		if err, ok := err.(fsproto.Error); ok {
			code = err.Code
		}

		if err != nil && code != -2 {
			return err
		}
	}

	return ifs.Remove(p)
}

// MkdirAll creates a directory and any necessary parents in the file system,
// similar to the mkdir -p command.
func (ifs *FS) MkdirAll(path string) error {
	if path == "" || path == "/" {
		return nil
	}

	splitPath := strings.Split(path, "/")
	for i := 1; i < len(splitPath); i++ {
		curPath := strings.Join(splitPath[0:i+1], "/")

		err := ifs.Mkdir(curPath)

		var code int8
		if err, ok := err.(fsproto.Error); ok {
			code = err.Code
		}

		if err != nil && code != -17 {
			return err
		}
	}

	return nil
}

var _ fs.File = (*File)(nil)

// File represents a remote file on a BLE filesystem.
//
// If ProgressFunc is set, it will be called whenever a read or write happens
// with the amount of bytes transferred and the total size of the file.
type File struct {
	fs           *FS
	path         string
	offset       uint32
	size         uint32
	readOnly     bool
	closed       bool
	ProgressFunc func(transferred, total uint32)
}

// Open opens an existing file at the specified path.
// It returns a handle for the file and an error, if any.
func (ifs *FS) Open(path string) (*File, error) {
	return &File{
		fs:       ifs,
		path:     path,
		offset:   0,
		readOnly: true,
	}, nil
}

// Create creates a new file with the specified path and size.
// It returns a handle for the created file and an error, if any.
func (ifs *FS) Create(path string, size uint32) (*File, error) {
	return &File{
		fs:     ifs,
		path:   path,
		offset: 0,
		size:   size,
	}, nil
}

// Write writes data from the byte slice b to the file.
// It returns the number of bytes written and an error, if any.
func (fl *File) Write(b []byte) (int, error) {
	if fl.closed {
		return 0, fsproto.ErrFileClosed
	}

	if fl.readOnly {
		return 0, fsproto.ErrFileReadOnly
	}

	fl.fs.mtx.Lock()
	defer fl.fs.mtx.Unlock()

	char, err := fl.fs.dev.getChar(fsTransferChar)
	if err != nil {
		return 0, err
	}
	defer char.EnableNotifications(nil)

	var chunkLen uint32

	dataLen := uint32(len(b))
	transferred := uint32(0)
	mtu := uint32(fl.fs.mtu(char))

	// continueCh is used to prevent race conditions. When the
	// request loop starts, it reads from continueCh, blocking it
	// until it's "released" by the notification function after
	// the response is processed.
	continueCh := make(chan struct{}, 2)
	var notifErr error
	err = char.EnableNotifications(func(buf []byte) {
		var wfr fsproto.WriteFileResponse
		err = fsproto.ReadResponse(buf, fsproto.WriteFileResp, &wfr)
		if err != nil {
			notifErr = err
			char.EnableNotifications(nil)
			close(continueCh)
			return
		}

		transferred += chunkLen
		fl.offset += chunkLen

		if wfr.FreeSpace == 0 || transferred == dataLen {
			char.EnableNotifications(nil)
			close(continueCh)
			return
		}

		if fl.ProgressFunc != nil {
			fl.ProgressFunc(transferred, fl.size)
		}

		// Release the request loop
		continueCh <- struct{}{}
	})

	err = fsproto.WriteRequest(char, fsproto.WriteFileHeaderOpcode, fsproto.WriteFileHeaderRequest{
		PathLen:  uint16(len(fl.path)),
		Offset:   fl.offset,
		FileSize: fl.size,
		Path:     fl.path,
	})
	if err != nil {
		return int(transferred), err
	}

	for range continueCh {
		if notifErr != nil {
			return int(transferred), notifErr
		}

		amountLeft := dataLen - transferred
		chunkLen = mtu
		if amountLeft < mtu {
			chunkLen = amountLeft
		}

		err = fsproto.WriteRequest(char, fsproto.WriteFileOpcode, fsproto.WriteFileRequest{
			Status:   0x01,
			Offset:   fl.offset,
			ChunkLen: chunkLen,
			Data:     b[transferred : transferred+chunkLen],
		})
		if err != nil {
			return int(transferred), err
		}
	}

	return int(transferred), notifErr
}

// Read reads data from the file into the byte slice b.
// It returns the number of bytes read and an error, if any.
func (fl *File) Read(b []byte) (int, error) {
	if fl.closed {
		return 0, fsproto.ErrFileClosed
	}

	fl.fs.mtx.Lock()
	defer fl.fs.mtx.Unlock()

	char, err := fl.fs.dev.getChar(fsTransferChar)
	if err != nil {
		return 0, err
	}
	defer char.EnableNotifications(nil)

	transferred := uint32(0)
	maxLen := uint32(len(b))
	mtu := uint32(fl.fs.mtu(char))

	var (
		notifErr error
		done     bool
	)

	// continueCh is used to prevent race conditions. When the
	// request loop starts, it reads from continueCh, blocking it
	// until it's "released" by the notification function after
	// the response is processed.
	continueCh := make(chan struct{}, 2)
	err = char.EnableNotifications(func(buf []byte) {
		var rfr fsproto.ReadFileResponse
		err = fsproto.ReadResponse(buf, fsproto.ReadFileResp, &rfr)
		if err != nil {
			notifErr = err
			char.EnableNotifications(nil)
			close(continueCh)
			return
		}

		fl.size = rfr.FileSize

		if rfr.Offset == rfr.FileSize || rfr.ChunkLen == 0 {
			notifErr = io.EOF
			done = true
			char.EnableNotifications(nil)
			close(continueCh)
			return
		}

		n := copy(b[transferred:], rfr.Data[:rfr.ChunkLen])
		fl.offset += uint32(n)
		transferred += uint32(n)

		if fl.ProgressFunc != nil {
			fl.ProgressFunc(transferred, rfr.FileSize)
		}

		// Release the request loop
		continueCh <- struct{}{}
	})
	if err != nil {
		return 0, err
	}
	defer char.EnableNotifications(nil)

	amountLeft := maxLen - transferred
	chunkLen := mtu
	if amountLeft < mtu {
		chunkLen = amountLeft
	}

	err = fsproto.WriteRequest(char, fsproto.ReadFileHeaderOpcode, fsproto.ReadFileHeaderRequest{
		PathLen: uint16(len(fl.path)),
		Offset:  fl.offset,
		ReadLen: chunkLen,
		Path:    fl.path,
	})
	if err != nil {
		return 0, err
	}

	if notifErr != nil {
		return int(transferred), notifErr
	}

	for !done {
		// Wait for the notification function to release the loop
		<-continueCh

		if notifErr != nil {
			return int(transferred), notifErr
		}

		amountLeft = maxLen - transferred
		chunkLen = mtu
		if amountLeft < mtu {
			chunkLen = amountLeft
		}

		err = fsproto.WriteRequest(char, fsproto.ReadFileOpcode, fsproto.ReadFileRequest{
			Status:  0x01,
			Offset:  fl.offset,
			ReadLen: chunkLen,
		})
		if err != nil {
			return int(transferred), err
		}
	}

	return int(transferred), notifErr
}

// Stat returns information about the file,
func (fl *File) Stat() (fs.FileInfo, error) {
	return fl.fs.Stat(fl.path)
}

// Seek sets the offset for the next Read or Write on the file to the specified offset.
// The whence parameter specifies the seek reference point:
//
//	io.SeekStart: offset is relative to the start of the file.
//	io.SeekCurrent: offset is relative to the current offset.
//	io.SeekEnd: offset is relative to the end of the file.
//
// Seek returns the new offset and an error, if any.
func (fl *File) Seek(offset int64, whence int) (int64, error) {
	if fl.closed {
		return 0, fsproto.ErrFileClosed
	}

	if offset > math.MaxUint32 {
		return 0, fsproto.ErrInvalidOffset
	}
	u32Offset := uint32(offset)

	fl.fs.mtx.Lock()
	defer fl.fs.mtx.Unlock()

	if fl.size == 0 {
		return 0, errors.New("file size unknown")
	}

	var newOffset uint32
	switch whence {
	case io.SeekStart:
		newOffset = u32Offset
	case io.SeekCurrent:
		newOffset = fl.offset + u32Offset
	case io.SeekEnd:
		newOffset = fl.size + u32Offset
	}

	if newOffset > fl.size || newOffset < 0 {
		return 0, fsproto.ErrInvalidOffset
	}
	fl.offset = newOffset

	return int64(fl.offset), nil
}

// Close closes the file for future operations
func (fl *File) Close() error {
	fl.fs.mtx.Lock()
	defer fl.fs.mtx.Unlock()
	fl.closed = true
	return nil
}

// requestThenAwaitResponse executes a BLE FS request and then waits for one or more responses,
// until fn returns true or an error is encountered.
func (ifs *FS) requestThenAwaitResponse(char *bluetooth.DeviceCharacteristic, opcode fsproto.FSReqOpcode, req any, fn func(buf []byte) (bool, error)) error {
	var stopped atomic.Bool
	errCh := make(chan error, 1)
	char.EnableNotifications(func(buf []byte) {
		stop, err := fn(buf)
		if err != nil && !stopped.Load() {
			errCh <- err
			char.EnableNotifications(nil)
			return
		} else if !stopped.Load() {
			errCh <- nil
		}

		if stop && !stopped.Load() {
			stopped.Store(true)
			close(errCh)
			char.EnableNotifications(nil)
		}
	})
	defer char.EnableNotifications(nil)

	err := fsproto.WriteRequest(char, opcode, req)
	if err != nil {
		return err
	}

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func (ifs *FS) mtu(char *bluetooth.DeviceCharacteristic) uint16 {
	mtuVal, _ := char.GetMTU()
	if mtuVal == 0 {
		mtuVal = 256
	}
	return mtuVal - 20
}

var _ fs.FS = (*GoFS)(nil)
var _ fs.StatFS = (*GoFS)(nil)
var _ fs.ReadDirFS = (*GoFS)(nil)

// GoFS implements [io/fs.FS], [io/fs.StatFS], and [io/fs.ReadDirFS]
// for the InfiniTime filesystem
type GoFS struct {
	*FS
}

// Open opens an existing file at the specified path.
// It returns a handle for the file and an error, if any.
func (gfs GoFS) Open(path string) (fs.File, error) {
	return gfs.FS.Open(path)
}
