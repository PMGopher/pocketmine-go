// Package data is a port of the top level of the pocketmine\data namespace.
package data

// SavedDataLoadingError is a port of pocketmine\data\SavedDataLoadingException: saved data
// (entity/tile/item NBT, ...) couldn't be loaded because it's invalid.
type SavedDataLoadingError struct {
	Message string
	Cause   error
}

func (e *SavedDataLoadingError) Error() string { return e.Message }

func (e *SavedDataLoadingError) Unwrap() error { return e.Cause }

// NewSavedDataLoadingError builds a SavedDataLoadingError with no cause.
func NewSavedDataLoadingError(message string) *SavedDataLoadingError {
	return &SavedDataLoadingError{Message: message}
}
