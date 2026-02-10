package check

import (
	"errors"
	"fmt"
)

var ErrNotEmpty = errors.New("not empty")

type ErrDirectoryNotEmpty struct {
	Path string
}

// Error implements error.
func (e *ErrDirectoryNotEmpty) Error() string {
	return fmt.Sprintf("%v is not empty", e.Path)
}

func (_ *ErrDirectoryNotEmpty) Is(target error) bool {
	return target == ErrNotEmpty
}
