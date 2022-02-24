package types

import (
	"fmt"
	"strconv"
)

const (
	ReqTypeHeartRate = iota
	ReqTypeBattLevel
	ReqTypeFwVersion
	ReqTypeFwUpgrade
	ReqTypeBtAddress
	ReqTypeNotify
	ReqTypeSetTime
	ReqTypeWatchHeartRate
	ReqTypeWatchBattLevel
	ReqTypeMotion
	ReqTypeWatchMotion
	ReqTypeStepCount
	ReqTypeWatchStepCount
	ReqTypeCancel
	ReqTypeFS
	ReqTypeWeatherUpdate
)

const (
	UpgradeTypeArchive = iota
	UpgradeTypeFiles
)

const (
	FSTypeWrite = iota
	FSTypeRead
	FSTypeMove
	FSTypeDelete
	FSTypeList
	FSTypeMkdir
)

type ReqDataFS struct {
	Type  int      `json:"type"`
	Files []string `json:"files"`
	Data  string   `json:"data,omitempty"`
}

type ReqDataFwUpgrade struct {
	Type  int
	Files []string
}

type Response struct {
	Type    int         `json:"type"`
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"msg,omitempty"`
	ID      string      `json:"id,omitempty"`
	Error   bool        `json:"error"`
}

type Request struct {
	Type int         `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

type ReqDataNotify struct {
	Title string
	Body  string
}

type DFUProgress struct {
	Received int64 `mapstructure:"recvd"`
	Total    int64 `mapstructure:"total"`
	Sent     int64 `mapstructure:"sent"`
}

type FSTransferProgress struct {
	Type  int    `json:"type" mapstructure:"type"`
	Total uint32 `json:"total" mapstructure:"total"`
	Sent  uint32 `json:"sent" mapstructure:"sent"`
	Done  bool   `json:"done" mapstructure:"done"`
}

type MotionValues struct {
	X int16
	Y int16
	Z int16
}

type FileInfo struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	IsDir bool   `json:"isDir"`
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
