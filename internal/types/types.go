package types

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
)

const (
	ResTypeHeartRate = iota
	ResTypeBattLevel
	ResTypeFwVersion
	ResTypeDFUProgress
	ResTypeBtAddress
	ResTypeNotify
	ResTypeSetTime
	ResTypeWatchHeartRate
	ResTypeWatchBattLevel
	ResTypeMotion
	ResTypeWatchMotion
	ResTypeStepCount
	ResTypeWatchStepCount
	ResTypeCancel
)

const (
	UpgradeTypeArchive = iota
	UpgradeTypeFiles
)

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
}

type MotionValues struct {
	X int16
	Y int16
	Z int16
}
