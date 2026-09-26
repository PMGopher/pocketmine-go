// Package tile is a port of pocketmine\block\tile.
//
// This package deliberately does NOT import pocketmine-go/pocketmine/block: block types will
// eventually need to import tile (to type-assert a World.GetTile result down to a concrete tile
// type, e.g. *tile.Note), and Go doesn't allow import cycles. So, exactly like block/utils and
// world/sound, tile declares its own minimal local World/Item interfaces instead of depending on
// block's - same reasoning as every other forward-compatible local interface in this port.
package tile

import (
	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	TagID = "id"
	TagX  = "x"
	TagY  = "y"
	TagZ  = "z"
)

// World is the minimal surface Tile needs from a world.
type World interface {
	RemoveTile(t Tile)
	// GetTileAt is the minimal surface Chest's pairing logic needs to look up a neighboring tile.
	GetTileAt(x, y, z int) (Tile, bool)
}

// Item is the minimal surface Tile.CopyDataFromItem/NameableComponent need from an item.
// GetNamedTag is needed by ShulkerBox.CopyDataFromItem's override, which (uniquely among
// container tiles) reads the item's whole NBT as save data rather than just its
// custom-block-data subset.
type Item interface {
	GetCustomBlockData() *nbt.CompoundTag
	GetNamedTag() *nbt.CompoundTag
	HasCustomName() bool
	GetCustomName() string
}

// Position is a minimal port of pocketmine\world\Position, local to this package for the same
// reason as World/Item above.
type Position struct {
	math.Vector3
	world World
}

func NewPosition(v math.Vector3, world World) Position { return Position{Vector3: v, world: world} }

func (p Position) IsValid() bool { return p.world != nil }

// GetWorld returns the position's world, and false if it has none (mirroring accessing a null
// world in PHP, which throws - same convention as block.Position.GetWorld).
func (p Position) GetWorld() (World, bool) {
	if p.world == nil {
		return nil, false
	}
	return p.world, true
}

// Tile is a port of pocketmine\block\tile\Tile. Concrete tile types embed TileBase and implement
// ReadSaveData/WriteSaveData/SaveID (SaveID is the save name TileFactory::getSaveId returns for
// the tile's class: the first name TileFactory registers for it).
type Tile interface {
	SaveNBT() *nbt.CompoundTag
	ReadSaveData(nbt *nbt.CompoundTag) error
	WriteSaveData(nbt *nbt.CompoundTag)
	SaveID() string
	GetPosition() Position
	IsClosed() bool
	Close()
	OnBlockDestroyed()
	CopyDataFromItem(item Item)
}

// TileBase is a port of pocketmine\block\tile\Tile's own state and default method bodies. Like
// Block in the block package, concrete tile types embed this and call Init(self) to wire up
// self-dispatch.
type TileBase struct {
	position Position
	closed   bool
	self     Tile
}

func NewTileBase(world World, pos math.Vector3) TileBase {
	return TileBase{position: NewPosition(pos, world)}
}

// Init wires up self-dispatch, the same self-referencing pattern as block.Block.Init.
func (t *TileBase) Init(self Tile) { t.self = self }

func (t *TileBase) rebind(self Tile) { t.self = self }

func (t *TileBase) GetPosition() Position { return t.position }

func (t *TileBase) IsClosed() bool { return t.closed }

// SaveNBT is a port of Tile::saveNBT. PHP looks the save ID up with
// TileFactory::getSaveId(get_class($this)); each tile type returns it from SaveID() instead.
func (t *TileBase) SaveNBT() *nbt.CompoundTag {
	n := nbt.NewCompoundTag()
	n.SetString(TagID, nbt.StringTag(t.self.SaveID()))
	n.SetInt(TagX, nbt.IntTag(t.position.FloorX()))
	n.SetInt(TagY, nbt.IntTag(t.position.FloorY()))
	n.SetInt(TagZ, nbt.IntTag(t.position.FloorZ()))
	n.SetLong(pocketmine.TagWorldDataVersion, nbt.LongTag(pocketmine.WorldDataVersion))
	t.self.WriteSaveData(n)
	return n
}

func (t *TileBase) GetCleanedNBT() *nbt.CompoundTag {
	n := nbt.NewCompoundTag()
	t.self.WriteSaveData(n)
	if n.Count() > 0 {
		return n
	}
	return nil
}

func (t *TileBase) CopyDataFromItem(item Item) {
	if blockNbt := item.GetCustomBlockData(); blockNbt != nil {
		// Best-effort, matching the PHP original's @internal contract (readSaveData errors were
		// only ever wrapped and rethrown as a RuntimeException here, not handled).
		_ = t.self.ReadSaveData(blockNbt)
	}
}

// OnBlockDestroyed is a port of Tile::onBlockDestroyed. onBlockDestroyedHook is optional - only
// tiles that need cleanup logic implement it (matching the PHP original's empty default body).
func (t *TileBase) OnBlockDestroyed() {
	if hook, ok := t.self.(interface{ OnBlockDestroyedHook() }); ok {
		hook.OnBlockDestroyedHook()
	}
	t.Close()
}

func (t *TileBase) Close() {
	if t.closed {
		return
	}
	// Tiles that override close() (containers remove their inventory's viewers first) implement
	// CloseHook.
	if hook, ok := t.self.(interface{ CloseHook() }); ok {
		hook.CloseHook()
	}
	t.closed = true
	if world, ok := t.position.GetWorld(); ok {
		world.RemoveTile(t.self)
	}
}
