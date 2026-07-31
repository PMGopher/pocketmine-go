package utils

// InternetException is a port of pocketmine\utils\InternetException (a plain \RuntimeException subclass).
type InternetException struct {
	Message string
	Cause   error
}

func NewInternetException(message string) *InternetException {
	return &InternetException{Message: message}
}

func (e *InternetException) Error() string { return e.Message }
func (e *InternetException) Unwrap() error { return e.Cause }
