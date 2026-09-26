// Package exception is a port of pocketmine\world\format\io\exception.
package exception

import "fmt"

// CorruptedChunkError is a port of CorruptedChunkException (a ChunkException): a chunk's saved
// data can't be decoded.
type CorruptedChunkError struct {
	Message string
	Cause   error
}

func (e *CorruptedChunkError) Error() string { return e.Message }
func (e *CorruptedChunkError) Unwrap() error { return e.Cause }

// NewCorruptedChunkError builds a CorruptedChunkError from a format string.
func NewCorruptedChunkError(format string, args ...any) *CorruptedChunkError {
	return &CorruptedChunkError{Message: fmt.Sprintf(format, args...)}
}

// WrapCorruptedChunk builds a CorruptedChunkError from another error (PHP's
// `new CorruptedChunkException($e->getMessage(), 0, $e)`).
func WrapCorruptedChunk(err error) *CorruptedChunkError {
	return &CorruptedChunkError{Message: err.Error(), Cause: err}
}
