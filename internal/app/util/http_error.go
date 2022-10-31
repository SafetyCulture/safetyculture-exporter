package util

import "encoding/json"

// HTTPError represents an easy way to pass HttpErrors from Sling
type HTTPError struct {
	StatusCode int    `json:"status_code"`
	Resource   string `json:"resource"`
	Message    string `json:"message"`
}

func (h HTTPError) Error() string {
	jsonBytes, err := json.Marshal(h)
	if err != nil {
		return ""
	}
	return string(jsonBytes)
}
