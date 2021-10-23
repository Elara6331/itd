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
	ReqTypeCancelHeartRate
	ReqTypeWatchBattLevel
	ReqTypeCancelBattLevel
	ReqTypeMotion
	ReqTypeWatchMotion
	ReqTypeCancelMotion
	ReqTypeStepCount
	ReqTypeWatchStepCount
	ReqTypeCancelStepCount
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
	ResTypeCancelHeartRate
	ResTypeWatchBattLevel
	ResTypeCancelBattLevel
	ResTypeMotion
	ResTypeWatchMotion
	ResTypeCancelMotion
	ResTypeStepCount
	ResTypeWatchStepCount
	ResTypeCancelStepCount
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
