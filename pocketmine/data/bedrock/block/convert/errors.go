// Package blockconvert is a port of pocketmine\data\bedrock\block\convert: converting between
// blocks and their Bedrock block states (name + properties), for the network and for world saves.
package blockconvert

import (
	"fmt"

	"pocketmine-go/pocketmine/data/bedrock"
)

// BlockStateSerializeError is a port of pocketmine\data\bedrock\block\BlockStateSerializeException.
type BlockStateSerializeError struct{ Message string }

func (e *BlockStateSerializeError) Error() string { return e.Message }

// BlockStateDeserializeError is a port of pocketmine\data\bedrock\block\BlockStateDeserializeException
// (shared with the block data upgrader, which lives outside this package).
type BlockStateDeserializeError = bedrock.BlockStateDeserializeError

// UnsupportedBlockStateError is a port of
// pocketmine\data\bedrock\block\convert\UnsupportedBlockStateException: the state is valid, but
// no block is mapped to it.
type UnsupportedBlockStateError struct{ BlockStateDeserializeError }

// The property closures and helpers throw like PHP (panic with one of the errors above); the
// public entry points (Serialize, Deserialize) recover them into returned errors.

func serializeError(format string, args ...any) {
	panic(&BlockStateSerializeError{Message: fmt.Sprintf(format, args...)})
}

func deserializeError(format string, args ...any) *BlockStateDeserializeError {
	return &BlockStateDeserializeError{Message: fmt.Sprintf(format, args...)}
}

// recoverError turns a panic with one of this package's errors into *err.
func recoverError(err *error) {
	if p := recover(); p != nil {
		switch e := p.(type) {
		case *BlockStateSerializeError:
			*err = e
		case *BlockStateDeserializeError:
			*err = e
		case *UnsupportedBlockStateError:
			*err = e
		default:
			panic(p)
		}
	}
}
