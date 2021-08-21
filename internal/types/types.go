package types

type ReqDataFwUpgrade struct {
	Type  int
	Files []string
}

type Response struct {
	Value   interface{} `json:"value,omitempty"`
	Message string      `json:"msg,omitempty"`
	Error   bool        `json:"error"`
}

type Request struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

type ReqDataNotify struct {
	Title string
	Body  string
}