package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// BaseCake is a port of pocketmine\block\BaseCake. The FoodSource interface it implements in PHP
// (getFoodRestore/getSaturationRestore/requiresHunger/getAdditionalEffects) isn't declared as a
// Go interface here: nothing in this port dispatches on it polymorphically yet, so the methods
// just live directly on BaseCake, same reasoning as skipping a formal Tile-registry interface
// until something actually needs to look blocks up by it.
type BaseCake struct {
	Transparent
}

func (b *BaseCake) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (b *BaseCake) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetSide(math.Down, 1).GetTypeId() != AIR
}

func (b *BaseCake) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return b.canBeSupportedAt(blockReplace) && b.Transparent.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (b *BaseCake) OnNearbyBlockChange() {
	if !b.canBeSupportedAt(b.self) {
		if world, err := b.position.GetWorld(); err == nil {
			world.UseBreakOn(b.position.AsVector3())
		}
	} else {
		b.Transparent.OnNearbyBlockChange()
	}
}

// OnInteract is a port of BaseCake::onInteract, minus the actual eating: Player.ConsumeObject
// doesn't exist yet (it needs Human's hunger/saturation/effect-application machinery, none of
// which is ported), so this just reports whether a player was present, matching the PHP return
// value shape without performing the consumption.
func (b *BaseCake) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return player != nil
}

func (b *BaseCake) GetFoodRestore() int { return 2 }

func (b *BaseCake) GetSaturationRestore() float64 { return 0.4 }

func (b *BaseCake) RequiresHunger() bool { return true }

// GetAdditionalEffects is a port of BaseCake::getAdditionalEffects, which always returns an empty
// array in the base class - EffectInstance (entity/effect package) isn't ported, so there's
// nothing to return a slice of yet.
func (b *BaseCake) GetAdditionalEffects() {}

// cakeShaper lets OnConsume reach a concrete leaf's GetResidue override - same self-dispatch shape
// as fireShaper.
type cakeShaper interface {
	GetResidue() Behavior
}

// OnConsume is a port of BaseCake::onConsume.
func (b *BaseCake) OnConsume(consumer Living) {
	world, err := b.position.GetWorld()
	if err != nil {
		return
	}
	_ = world.SetBlock(b.position, b.self.(cakeShaper).GetResidue())
}
