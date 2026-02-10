package rest

import (
	"encoding/json"
)

type Error struct {
	Code string `json:"code,omitempty"`

	Message string `json:"message,omitempty"`

	Cause string `json:"cause,omitempty"`

	Detail string `json:"detail,omitempty"`
}

func (e Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}
