package block

import "fmt"

// Tile is the minimal surface BlockIdentifier needs from a block entity ("tile" in Bedrock
// terms — a chest, furnace, sign, etc.). Declared as a marker here; the future block/tile
// package will flesh out its real methods once it's ported, the same forward-compatible-local-
// interface pattern used for permission.Plugin and command.Server elsewhere in this port.
type Tile interface {
	//marker — block/tile will add real methods once it's ported
}

// BlockIdentifier is a port of pocketmine\block\BlockIdentifier.
//
// PHP takes an optional `class-string<Tile>` and validates it via reflection
// (Utils::testValidInstance) at construction time. Go has no runtime "is this a valid
// constructible Tile subclass" check to make — the natural equivalent is a constructor function
// already typed to return something satisfying Tile, so an invalid type is a compile error
// instead of a runtime one.
type BlockIdentifier struct {
	blockTypeID int
	newTile     func() Tile
}

func NewBlockIdentifier(blockTypeID int, newTile func() Tile) (*BlockIdentifier, error) {
	if blockTypeID < 0 {
		return nil, fmt.Errorf("block type ID may not be negative")
	}
	return &BlockIdentifier{blockTypeID: blockTypeID, newTile: newTile}, nil
}

func (b *BlockIdentifier) GetBlockTypeID() int { return b.blockTypeID }

// NewTileInstance constructs this block's tile. ok is false if this block type has no tile.
func (b *BlockIdentifier) NewTileInstance() (tile Tile, ok bool) {
	if b.newTile == nil {
		return nil, false
	}
	return b.newTile(), true
}
