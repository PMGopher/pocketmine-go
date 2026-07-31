package utils

import "fmt"

// ConfigLoadException is a port of pocketmine\utils\ConfigLoadException.
type ConfigLoadException struct {
	Message string
	Cause   error
}

// WrapConfigLoadException mirrors ConfigLoadException::wrap().
func WrapConfigLoadException(fileName string, cause error) *ConfigLoadException {
	return &ConfigLoadException{
		Message: fmt.Sprintf("Failed to parse config %s: %s", fileName, cause.Error()),
		Cause:   cause,
	}
}

func (e *ConfigLoadException) Error() string { return e.Message }
func (e *ConfigLoadException) Unwrap() error { return e.Cause }
