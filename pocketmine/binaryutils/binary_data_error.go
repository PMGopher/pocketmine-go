package binaryutils

import "fmt"

// BinaryDataError is a port of pocketmine\utils\BinaryDataException (a plain \RuntimeException
// subclass), raised when a buffer doesn't have enough bytes left to satisfy a read.
type BinaryDataError struct {
	Message string
}

func newBinaryDataError(format string, args ...any) *BinaryDataError {
	return &BinaryDataError{Message: fmt.Sprintf(format, args...)}
}

func (e *BinaryDataError) Error() string { return e.Message }
