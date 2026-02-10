package chartmuseum

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ChartmuseumError 是 chartmuseum 返回的错误。
type ChartmuseumError struct {
	// A human-readable description of the status of this operation.
	Message string `json:"message,omitempty"`

	// Suggested HTTP return code for this status, 0 if not set.
	Code int `json:"code,omitempty"`
}

var (
	ErrUnauthorized = ChartmuseumError{Code: http.StatusUnauthorized}
)

// Error implements the Error interface.
func (e *ChartmuseumError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

func (e *ChartmuseumError) Is(target error) bool {
	if err := new(ChartmuseumError); errors.As(target, &err) {
		return e.Code == err.Code
	}
	return false
}

type response struct {
	Error string `json:"error,omitempty"`
}

// ErrorFromStatusCodeAndBody generates an ChartmuseumError from status code and
// body.
func ErrorFromStatusCodeAndBody(st int, body []byte) error {
	resp := new(response)
	if err := json.Unmarshal(body, resp); err != nil {
		return err
	}
	return &ChartmuseumError{Code: st, Message: resp.Error}
}
