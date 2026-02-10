package validation

import (
	"bytes"
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

type InvalidError struct {
	field.ErrorList
}

func (e *InvalidError) Error() string {
	if len(e.ErrorList) == 1 {
		return fmt.Sprintf("invalid: %v", e.ErrorList[0])
	}

	buff := new(bytes.Buffer)
	fmt.Fprintf(buff, "invalid:\n")
	for _, err := range e.ErrorList {
		fmt.Fprintf(buff, "  %v\n", err)
	}
	return buff.String()
}
