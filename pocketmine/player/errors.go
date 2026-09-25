package player

// PlayerDataLoadError is a port of pocketmine\player\PlayerDataLoadException.
type PlayerDataLoadError struct {
	Message string
	Cause   error
}

func (e *PlayerDataLoadError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}
func (e *PlayerDataLoadError) Unwrap() error { return e.Cause }

// PlayerDataSaveError is a port of pocketmine\player\PlayerDataSaveException.
type PlayerDataSaveError struct {
	Message string
	Cause   error
}

func (e *PlayerDataSaveError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}
func (e *PlayerDataSaveError) Unwrap() error { return e.Cause }
