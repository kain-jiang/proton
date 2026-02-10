package v1alpha1

import (
	"bytes"
	"fmt"
)

type Error struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Cause   string `json:"cause,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

func (e *Error) Error() string {
	var buff bytes.Buffer
	fmt.Fprintf(&buff, "%d: %s", e.Code, e.Message)
	if e.Cause != "" {
		fmt.Fprintf(&buff, ": %s", e.Cause)
	}
	return buff.String()
}

var _ error = (*Error)(nil)
