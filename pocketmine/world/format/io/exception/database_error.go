package exception

import "fmt"

// DatabaseError is an error reading the world's database itself (a damaged LevelDB table, an I/O
// error), as opposed to a CorruptedChunkError in one chunk's data. PHP's LevelDBException isn't
// caught by World::loadChunk, so it stops the server instead of the chunk being regenerated over
// the real one.
type DatabaseError struct {
	Err error
}

func NewDatabaseError(err error) *DatabaseError { return &DatabaseError{Err: err} }

func (e *DatabaseError) Error() string { return fmt.Sprintf("world database error: %v", e.Err) }

func (e *DatabaseError) Unwrap() error { return e.Err }
