package configuration

import "github.com/hashicorp/go-multierror"

func AppendError(err error, errs ...error) error {
	if allNil(errs) {
		return err
	}
	return multierror.Append(err, errs...)
}

func allNil(errs []error) bool {
	for _, err := range errs {
		if err != nil {
			return false
		}
	}
	return true
}
