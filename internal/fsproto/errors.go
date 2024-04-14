package fsproto

import (
	"errors"
	"fmt"
)

var (
	ErrFileNotExists = errors.New("file does not exist")
	ErrFileReadOnly  = errors.New("file is read only")
	ErrFileWriteOnly = errors.New("file is write only")
	ErrInvalidOffset = errors.New("offset out of range")
	ErrNoRemoveRoot  = errors.New("refusing to remove root directory")
	ErrFileClosed    = errors.New("cannot perform operation on a closed file")
)

// Error represents an error returned by BLE FS
type Error struct {
	Code int8
}

// Error returns the string associated with the error code
func (err Error) Error() string {
	switch err.Code {
	case 0x02:
		return "filesystem error"
	case 0x05:
		return "read-only filesystem"
	case 0x03:
		return "no such file"
	case 0x04:
		return "protocol error"
	case -5:
		return "input/output error"
	case -84:
		return "filesystem is corrupted"
	case -2:
		return "no such directory entry"
	case -17:
		return "entry already exists"
	case -20:
		return "entry is not a directory"
	case -39:
		return "directory is not empty"
	case -9:
		return "bad file number"
	case -27:
		return "file is too large"
	case -22:
		return "invalid parameter"
	case -28:
		return "no space left on device"
	case -12:
		return "no more memory available"
	case -61:
		return "no attr available"
	case -36:
		return "file name is too long"
	default:
		return fmt.Sprintf("unknown error (code %d)", err.Code)
	}
}
