package infinitime

import (
	"fmt"
	"io/fs"
	"strconv"
	"time"
)

// DirEntry represents an entry from a directory listing
type DirEntry struct {
	flags   uint32
	modtime uint64
	size    uint32
	path    string
}

// Name returns the name of the file described by the entry
func (de DirEntry) Name() string {
	return de.path
}

// IsDir reports whether the entry describes a directory.
func (de DirEntry) IsDir() bool {
	return de.flags&0b1 == 1
}

// Type returns the type bits for the entry.
func (de DirEntry) Type() fs.FileMode {
	if de.IsDir() {
		return fs.ModeDir
	} else {
		return 0
	}
}

// Info returns the FileInfo for the file or subdirectory described by the entry.
func (de DirEntry) Info() (fs.FileInfo, error) {
	return FileInfo{
		name:    de.path,
		size:    de.size,
		modtime: de.modtime,
		mode:    de.Type(),
		isDir:   de.IsDir(),
	}, nil
}

func (de DirEntry) String() string {
	var isDirChar rune
	if de.IsDir() {
		isDirChar = 'd'
	} else {
		isDirChar = '-'
	}

	// Get human-readable value for file size
	val, unit := bytesHuman(de.size)
	prec := 0
	// If value is less than 10, set precision to 1
	if val < 10 {
		prec = 1
	}
	// Convert float to string
	valStr := strconv.FormatFloat(val, 'f', prec, 64)

	// Return string formatted like so:
	// -  10 kB file
	// or:
	// d   0 B  .
	return fmt.Sprintf(
		"%c %3s %-2s %s",
		isDirChar,
		valStr,
		unit,
		de.path,
	)
}

func bytesHuman(b uint32) (float64, string) {
	const unit = 1000
	// Set possible unit prefixes (PineTime flash is 4MB)
	units := [2]rune{'k', 'M'}
	// If amount of bytes is less than smallest unit
	if b < unit {
		// Return unchanged with unit "B"
		return float64(b), "B"
	}

	div, exp := uint32(unit), 0
	// Get decimal values and unit prefix index
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	// Create string for full unit
	unitStr := string([]rune{units[exp], 'B'})

	// Return decimal with unit string
	return float64(b) / float64(div), unitStr
}

// FileInfo implements fs.FileInfo
type FileInfo struct {
	name    string
	size    uint32
	modtime uint64
	mode    fs.FileMode
	isDir   bool
}

// Name returns the base name of the file
func (fi FileInfo) Name() string {
	return fi.name
}

// Size returns the total size of the file
func (fi FileInfo) Size() int64 {
	return int64(fi.size)
}

// Mode returns the mode of the file
func (fi FileInfo) Mode() fs.FileMode {
	return fi.mode
}

// ModTime returns the modification time of the file
// As of now, this is unimplemented in InfiniTime, and
// will always return 0.
func (fi FileInfo) ModTime() time.Time {
	return time.Unix(0, int64(fi.modtime))
}

// IsDir returns whether the file is a directory
func (fi FileInfo) IsDir() bool {
	return fi.isDir
}

// Sys is unimplemented and returns nil
func (fi FileInfo) Sys() any {
	return nil
}
