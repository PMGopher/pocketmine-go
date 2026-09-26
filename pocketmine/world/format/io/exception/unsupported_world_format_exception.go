package exception

import "fmt"

// UnsupportedWorldFormatError is a port of UnsupportedWorldFormatException (a WorldException): the
// world is valid, but in a format or version this server can't load.
type UnsupportedWorldFormatError struct{ Message string }

func (e *UnsupportedWorldFormatError) Error() string { return e.Message }

// NewUnsupportedWorldFormatError builds an UnsupportedWorldFormatError from a format string.
func NewUnsupportedWorldFormatError(format string, args ...any) *UnsupportedWorldFormatError {
	return &UnsupportedWorldFormatError{Message: fmt.Sprintf(format, args...)}
}
