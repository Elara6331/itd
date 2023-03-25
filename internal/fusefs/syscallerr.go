package fusefs

import (
	"syscall"

	"go.arsenm.dev/infinitime/blefs"
)

func syscallErr(err error) syscall.Errno {
	if err == nil {
		return 0
	}

	switch err := err.(type) {
	case blefs.FSError:
		switch err.Code {
		case 0x02: // filesystem error
			return syscall.EIO // TODO
		case 0x05: // read-only filesystem
			return syscall.EROFS
		case 0x03: // no such file
			return syscall.ENOENT
		case 0x04: // protocol error
			return syscall.EPROTO
		case -5: // input/output error
			return syscall.EIO
		case -84: // filesystem is corrupted
			return syscall.ENOTRECOVERABLE // TODO
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
			return syscall.ENODATA // TODO
		case -36: // file name is too long
			return syscall.ENAMETOOLONG
		}
	default:
		switch err {
		case blefs.ErrFileNotExists: // file does not exist
			return syscall.ENOENT
		case blefs.ErrFileReadOnly: // file is read only
			return syscall.EACCES
		case blefs.ErrFileWriteOnly: // file is write only
			return syscall.EACCES
		case blefs.ErrInvalidOffset: // invalid file offset
			return syscall.EFAULT // TODO
		case blefs.ErrOffsetChanged: // offset has already been changed
			return syscall.ESPIPE
		case blefs.ErrReadOpen: // only one file can be opened for reading at a time
			return syscall.ENFILE
		case blefs.ErrWriteOpen: // only one file can be opened for writing at a time
			return syscall.ENFILE
		case blefs.ErrNoRemoveRoot: // refusing to remove root directory
			return syscall.EPERM
		}
	}

	return syscall.EIO
}
