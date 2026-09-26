package bedrockitem

// ItemTypeSerializeError is a port of pocketmine\data\bedrock\item\ItemTypeSerializeException.
type ItemTypeSerializeError struct{ Message string }

func (e *ItemTypeSerializeError) Error() string { return e.Message }

// ItemTypeDeserializeError is a port of pocketmine\data\bedrock\item\ItemTypeDeserializeException.
type ItemTypeDeserializeError struct{ Message string }

func (e *ItemTypeDeserializeError) Error() string { return e.Message }

// UnsupportedItemTypeError is a port of pocketmine\data\bedrock\item\UnsupportedItemTypeException.
type UnsupportedItemTypeError struct{ Message string }

func (e *UnsupportedItemTypeError) Error() string { return e.Message }

// recoverItemError turns a panic with one of this package's errors into *err (the meta
// deserializers throw like PHP).
func recoverItemError(err *error) {
	if p := recover(); p != nil {
		switch e := p.(type) {
		case *ItemTypeDeserializeError:
			*err = e
		case *UnsupportedItemTypeError:
			*err = e
		case *ItemTypeSerializeError:
			*err = e
		default:
			panic(p)
		}
	}
}
