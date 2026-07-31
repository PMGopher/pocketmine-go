package utils

// ServerException is a port of pocketmine\utils\ServerException (a plain \RuntimeException subclass).
type ServerException struct {
	Message string
	Cause   error
}

func NewServerException(message string) *ServerException {
	return &ServerException{Message: message}
}

func (e *ServerException) Error() string { return e.Message }
func (e *ServerException) Unwrap() error { return e.Cause }
