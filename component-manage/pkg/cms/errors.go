package cms

// NotFoundError resource not found
type NotFoundError struct {
	msg string
}

func (e *NotFoundError) Error() string {
	return e.msg
}

// IsNotFoundError the Error type instance
func IsNotFoundError(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}
