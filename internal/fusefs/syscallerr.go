package fusefs

import (
	"syscall"

	"go.elara.ws/itd/internal/fsproto"
)

func syscallErr(err error) syscall.Errno {
	if err == nil {
		return 0
	}

	switch err := err.(type) {
	case fsproto.Error:
		switch err.Code {
		case 0x02: // filesystem error
			return syscall.EIO
		case 0x05: // read-only filesystem
			return syscall.EROFS
		case 0x03: // no such file
			return syscall.ENOENT
		case 0x04: // protocol error
			return syscall.EPROTO
		case -5: // input/output error
			return syscall.EIO
		case -84: // filesystem is corrupted
			return syscall.ENOTRECOVERABLE
		case -2: // no such directory entry
			return syscall.ENOENT
		case -17: // entry already exists
			return syscall.EEXIST
		case -20: // entry is not a directory
			return syscall.ENOTDIR
		case -39: // directory is not empty
			return syscall.ENOTEMPTY
		case -9: // bad file number
			return syscall.EBADF
		case -27: // file is too large
			return syscall.EFBIG
		case -22: // invalid parameter
			return syscall.EINVAL
		case -28: // no space left on device
			return syscall.ENOSPC
		case -12: // no more memory available
			return syscall.ENOMEM
		case -61: // no attr available
			return syscall.ENODATA
		case -36: // file name is too long
			return syscall.ENAMETOOLONG
		}
	default:
		switch err {
		case fsproto.ErrFileNotExists: // file does not exist
			return syscall.ENOENT
		case fsproto.ErrFileReadOnly: // file is read only
			return syscall.EACCES
		case fsproto.ErrFileWriteOnly: // file is write only
			return syscall.EACCES
		case fsproto.ErrInvalidOffset: // invalid file offset
			return syscall.EINVAL
		case fsproto.ErrNoRemoveRoot: // refusing to remove root directory
			return syscall.EPERM
		case fsproto.ErrFileClosed: // cannot perform operation on closed file
			return syscall.EBADF
		default:
			return syscall.EINVAL
		}
	}

	return syscall.EIO
}
