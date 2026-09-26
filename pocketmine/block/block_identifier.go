package block

import (
	"fmt"
	"reflect"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

// Tile is a type alias (not just an interface with matching methods) for tile.Tile - the real
// pocketmine\block\tile\Tile interface, now that pocketmine/block/tile is substantially built out.
// A plain alias (rather than re-declaring the method set here) means any tile.Tile value - e.g.
// what format.Chunk's own tile storage holds - is usable as a block.Tile with no wrapping or
// assertion anywhere. block/tile doesn't import block (it deliberately declares its own minimal
// local World/Item interfaces instead, exactly so this alias could exist with no cycle risk - see
// tile's own package doc comment).
type Tile = tile.Tile

// BlockIdentifier is a port of pocketmine\block\BlockIdentifier.
//
// PHP takes an optional `class-string<Tile>` and validates it via reflection
// (Utils::testValidInstance) at construction time. Go has no runtime "is this a valid
// constructible Tile subclass" check to make — the natural equivalent is a constructor function
// already typed to return something satisfying Tile, so an invalid type is a compile error
// instead of a runtime one.
type BlockIdentifier struct {
	blockTypeID int
	newTile     TileFactory
	tileType    reflect.Type
}

// TileFactory creates the tile of a block (PHP's tile class-string plus `new $tileClass($world, $pos)`).
type TileFactory func(world tile.World, pos math.Vector3) Tile

func NewBlockIdentifier(blockTypeID int, newTile TileFactory) (*BlockIdentifier, error) {
	if blockTypeID < 0 {
		return nil, fmt.Errorf("block type ID may not be negative")
	}
	return &BlockIdentifier{blockTypeID: blockTypeID, newTile: newTile}, nil
}

func (b *BlockIdentifier) GetBlockTypeID() int { return b.blockTypeID }

// HasTile reports whether this block type has a tile (getTileClass() !== null).
func (b *BlockIdentifier) HasTile() bool { return b.newTile != nil }

// NewTileInstance constructs this block's tile at pos. ok is false if this block type has no tile.
func (b *BlockIdentifier) NewTileInstance(world tile.World, pos math.Vector3) (t Tile, ok bool) {
	if b.newTile == nil {
		return nil, false
	}
	return b.newTile(world, pos), true
}

// IsTileType is `$tile instanceof $this->getTileClass()`: whether t is this block type's tile
// class (same concrete type as the tiles NewTileInstance builds).
func (b *BlockIdentifier) IsTileType(t Tile) bool {
	if b.newTile == nil || t == nil {
		return false
	}
	if b.tileType == nil {
		b.tileType = reflect.TypeOf(b.newTile(nil, math.Vector3{}))
	}
	return reflect.TypeOf(t) == b.tileType
}
