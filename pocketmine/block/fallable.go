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
// The trait's onNearbyBlockChange is FallableOnNearbyBlockChange below (it needs the concrete
// block, which a component doesn't have).
type FallableComponent struct{}

// SpawnFallingBlockFunc creates and spawns a FallingBlock entity for blk (which has just been
// replaced with air). pocketmine/entity/object, which imports this package, installs it from its
// init() - the same dependency-inversion hook as NewItemBlockFunc.
var SpawnFallingBlockFunc func(blk Behavior)

// FallableOnNearbyBlockChange is a port of FallableTrait::onNearbyBlockChange: an unsupported
// fallable block turns into a FallingBlock entity.
func FallableOnNearbyBlockChange(blk Behavior) {
	pos := blk.GetPosition()
	world, err := pos.GetWorld()
	if err != nil {
		return
	}
	down := world.GetBlockAt(pos.FloorX(), pos.FloorY()-1, pos.FloorZ())
	if down.CanBeReplaced() {
		_ = world.SetBlock(pos, VanillaAir())

		if SpawnFallingBlockFunc != nil {
			SpawnFallingBlockFunc(blk)
		}
	}
}

func (FallableComponent) TickFalling() (Behavior, bool) { return nil, false }

func (FallableComponent) OnHitGround(blockEntity FallingBlockEntity) bool { return true }

func (FallableComponent) GetFallDamagePerBlock() float64 { return 0 }

func (FallableComponent) GetMaxFallDamage() float64 { return 0 }

func (FallableComponent) GetLandSound() (sound.Sound, bool) { return nil, false }
