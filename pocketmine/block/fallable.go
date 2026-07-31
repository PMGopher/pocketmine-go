package block

import "pocketmine-go/pocketmine/world/sound"

// FallingBlockEntity is a minimal forward-compatible surface for
// pocketmine\entity\object\FallingBlock — declared here since Fallable.OnHitGround is the only
// current consumer, matching the local-interface pattern used elsewhere for not-yet-ported types.
type FallingBlockEntity interface {
	GetFallDistance() float64
}

// Fallable is a port of pocketmine\block\utils\Fallable.
type Fallable interface {
	// TickFalling returns a replacement block for FallingBlock to become on its next tick (e.g.
	// turning into a fluid on contact with water), or ok=false to keep falling unchanged.
	TickFalling() (replacement Behavior, ok bool)
	OnHitGround(blockEntity FallingBlockEntity) bool
	GetFallDamagePerBlock() float64
	GetMaxFallDamage() float64
	// GetLandSound returns the sound played when FallingBlock hits the ground, or ok=false for none.
	GetLandSound() (s sound.Sound, ok bool)
}

// FallableComponent is a port of pocketmine\block\utils\FallableTrait's default method bodies.
//
// FallableTrait::onNearbyBlockChange (replace self with air and spawn a FallingBlock entity) isn't
// ported here: it needs both the block registry (VanillaBlocks.AIR()) and the FallingBlock entity
// type, neither of which exist yet. Concrete Fallable block types should call World.UseBreakOn (or
// leave OnNearbyBlockChange as Block's default) as a stand-in until entity/src and the block
// registry are ported — see individual block files for the exact gap noted at their call site.
type FallableComponent struct{}

func (FallableComponent) TickFalling() (Behavior, bool) { return nil, false }

func (FallableComponent) OnHitGround(blockEntity FallingBlockEntity) bool { return true }

func (FallableComponent) GetFallDamagePerBlock() float64 { return 0 }

func (FallableComponent) GetMaxFallDamage() float64 { return 0 }

func (FallableComponent) GetLandSound() (sound.Sound, bool) { return nil, false }
