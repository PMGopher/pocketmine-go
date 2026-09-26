package exception

import "fmt"

// CorruptedWorldError is a port of CorruptedWorldException (a WorldException): the world's files
// (level.dat, database) can't be read.
type CorruptedWorldError struct {
	Message string
	Cause   error
}

func (e *CorruptedWorldError) Error() string { return e.Message }
func (e *CorruptedWorldError) Unwrap() error { return e.Cause }

// NewCorruptedWorldError builds a CorruptedWorldError from a format string.
func NewCorruptedWorldError(format string, args ...any) *CorruptedWorldError {
	return &CorruptedWorldError{Message: fmt.Sprintf(format, args...)}
}
