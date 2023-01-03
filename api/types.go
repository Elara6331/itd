package api

import (
	"fmt"
	"strconv"
)

type UpgradeType uint8

const (
	UpgradeTypeArchive UpgradeType = iota
	UpgradeTypeFiles
)

type FSData struct {
	Files []string
	Data  string
}

type FwUpgradeData struct {
	Type  UpgradeType
	Files []string
}

type NotifyData struct {
	Title string
	Body  string
}

type FSTransferProgress struct {
	Total uint32
	Sent  uint32
	Err   error
}

type FileInfo struct {
	Name  string
	Size  int64
	IsDir bool
}

func (fi FileInfo) String() string {
	var isDirChar rune
	if fi.IsDir {
		isDirChar = 'd'
	} else {
		isDirChar = '-'
	}

	// Get human-readable value for file size
	val, unit := bytesHuman(fi.Size)
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
		fi.Name,
	)
}

// bytesHuman returns a human-readable string for
// the amount of bytes inputted.
func bytesHuman(b int64) (float64, string) {
	const unit = 1000
	// Set possible units prefixes (PineTime flash is 4MB)
	units := [2]rune{'k', 'M'}
	// If amount of bytes is less than smallest unit
	if b < unit {
		// Return unchanged with unit "B"
		return float64(b), "B"
	}

	div, exp := int64(unit), 0
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
