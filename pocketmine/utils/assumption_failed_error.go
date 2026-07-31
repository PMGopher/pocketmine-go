package utils

// AssumptionFailedError is a port of pocketmine\utils\AssumptionFailedError.
//
// It should be used (panic'd with) in places where something is assumed to be true, but the
// type system does not provide a guarantee, so the server crashes properly if the assumption
// does not hold.
type AssumptionFailedError struct {
	Message string
}

func NewAssumptionFailedError(message string) *AssumptionFailedError {
	return &AssumptionFailedError{Message: message}
}

func (e *AssumptionFailedError) Error() string {
	return e.Message
}
