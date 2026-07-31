package block

import (
	"fmt"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// World, Chunk, Entity, Player and Projectile are the minimal surfaces Block needs from the
// not-yet-ported world/entity/player packages. Declared locally so those future packages satisfy
// them automatically with no import needed here — the same forward-compatible-local-interface
// pattern used for Tile/Item/ToolTier above, and for permission.Plugin/command.Server elsewhere
// in this port.
type World interface {
	GetBlockAt(x, y, z int) Behavior
	SetBlock(pos Position, blk Behavior) error
	GetTile(pos Position) (Tile, bool)
	AddTile(tile Tile)
	GetOrLoadChunkAtPosition(pos Position) (Chunk, bool)
	AddSound(pos math.Vector3, s sound.Sound)
	ScheduleDelayedBlockUpdate(pos math.Vector3, delay int)
	GetFullLightAt(x, y, z int) int
	GetBlockLightAt(x, y, z int) int
	// UseBreakOn is the simplified (no item/player/particles) form of World::useBreakOn, the only
	// form used from within block package (Button.onNearbyBlockChange).
	UseBreakOn(pos math.Vector3) bool
}

type Chunk interface {
	SetBlockStateID(x, y, z int, stateID int)
}

type Entity interface {
	ResetFallDistance()
	GetPosition() math.Vector3
	SetOnGround(onGround bool)
}

// Living is a forward-compatible marker for pocketmine\entity\Living — same pattern as
// Flowable.go's Liquid marker. The future Living entity type just needs to satisfy this
// trivially; the block package only ever uses it for `instanceof Living` checks.
type Living interface {
	Entity
	IsLiving() bool
}

type Player interface {
	Entity
	GetHorizontalFacing() math.Facing
}

type Projectile interface {
	Entity
}

// BlockTransaction is the minimal surface Block.Place needs — a set of block changes to be
// applied together once placement succeeds, rather than writing to the world directly.
type BlockTransaction interface {
	AddBlock(pos Position, blk Behavior)
}

// Position is a port of pocketmine\world\Position: a Vector3 that optionally knows which World
// it's in.
type Position struct {
	math.Vector3
	world World
}

func NewPosition(x, y, z float64, world World) Position {
	return Position{Vector3: math.NewVector3(x, y, z), world: world}
}

func (p Position) IsValid() bool { return p.world != nil }

// GetWorld returns the position's world, erroring if it has none (mirroring accessing a null
// world reference in PHP, which is a programmer error at the call site).
func (p Position) GetWorld() (World, error) {
	if p.world == nil {
		return nil, fmt.Errorf("position does not have a valid world")
	}
	return p.world, nil
}

func (p Position) AsVector3() math.Vector3 { return p.Vector3 }

func (p Position) GetSide(side math.Facing, step int) Position {
	return Position{Vector3: p.Vector3.GetSide(side, step), world: p.world}
}
