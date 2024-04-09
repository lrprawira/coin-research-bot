package common

type StatusData struct {
	Timestamp    string `json:"timestamp"`
	ErrorMessage string `json:"error_message"`
	ErrorCode    string `json:"error_code"`
}
