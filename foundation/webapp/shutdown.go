package webapp

import (
	"errors"
)

// shutdownError is a type used to help termination of the service.
type shutdownError struct {
	Message string
}

// NewShutdownError returns an error that causes the framework to signal
// to shutdown.
func NewShutdownError(message string) error {
	return &shutdownError{message}
}

func (se *shutdownError) Error() string {
	return se.Message
}

func IsShutdown(err error) bool {
	var se *shutdownError
	return errors.As(err, &se)
}
