package util

import "encoding/json"

// HttpError represents an easy way to pass HttpErrors from Sling
type HttpError struct {
	StatusCode int    `json:"status_code"`
	Resource   string `json:"resource"`
	Message    string `json:"message"`
}

func (h HttpError) Error() string {
	jsonBytes, err := json.Marshal(h)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}
