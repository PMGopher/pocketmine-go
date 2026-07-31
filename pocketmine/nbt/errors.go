package nbt

// NbtDataException, UnexpectedTagTypeException and NoSuchTagException are ports of the
// pocketmine/nbt exception hierarchy (all plain \RuntimeException subclasses in PHP, used purely
// for `instanceof`/error-message purposes).
type NbtDataException struct{ Message string }

func NewNbtDataException(message string) *NbtDataException { return &NbtDataException{message} }
func (e *NbtDataException) Error() string                  { return e.Message }

type UnexpectedTagTypeException struct{ Message string }

func NewUnexpectedTagTypeException(message string) *UnexpectedTagTypeException {
	return &UnexpectedTagTypeException{message}
}
func (e *UnexpectedTagTypeException) Error() string { return e.Message }

type NoSuchTagException struct{ Message string }

func NewNoSuchTagException(message string) *NoSuchTagException {
	return &NoSuchTagException{message}
}
func (e *NoSuchTagException) Error() string { return e.Message }

// InvalidTagValueException is a port of \InvalidArgumentException subclass InvalidTagValueException.
//
// It exists in PHP to enforce integer tag ranges and string/array size limits at construction
// time. Go's own int8/int16/int32/int64 types already enforce the byte/short/int/long ranges at
// compile time, so this is only needed for the checks Go's type system can't express (string
// length, tag name length).
type InvalidTagValueException struct{ Message string }

func NewInvalidTagValueException(message string) *InvalidTagValueException {
	return &InvalidTagValueException{message}
}
func (e *InvalidTagValueException) Error() string { return e.Message }
