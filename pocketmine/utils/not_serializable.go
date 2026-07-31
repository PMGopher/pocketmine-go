package utils

import "errors"

// NotSerializable is a port of pocketmine\utils\NotSerializable.
//
// PHP's trait blocks the single __serialize()/__unserialize() hook. Go has no equivalent
// universal hook, so this instead blocks the two most common serialization paths a struct
// could otherwise fall into: encoding/json and encoding/gob.
type NotSerializable struct{}

func (NotSerializable) MarshalJSON() ([]byte, error) {
	return nil, errors.New("serialization of this object is not allowed")
}

func (*NotSerializable) UnmarshalJSON([]byte) error {
	return errors.New("unserialization of this object is not allowed")
}

func (NotSerializable) GobEncode() ([]byte, error) {
	return nil, errors.New("serialization of this object is not allowed")
}

func (*NotSerializable) GobDecode([]byte) error {
	return errors.New("unserialization of this object is not allowed")
}
