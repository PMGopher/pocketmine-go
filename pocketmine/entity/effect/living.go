// Package effect is a port of pocketmine\entity\effect.
//
// It sits below pocketmine/item (items build EffectInstances for food/potions) and below
// pocketmine/entity (Living owns an EffectManager), so the Living/Human entities effects act on are
// described by the local interfaces below rather than imported. *entity.Living (and every type
// embedding it) satisfies Living structurally. The two places PHP narrows further with instanceof
// (Human for hunger, Player for levitation) go through the HumanHungerManager/IsPlayer hooks, which
// pocketmine/entity and pocketmine/player install from their own init() - the same
// dependency-inversion pattern as block.NewItemBlockFunc.
package effect

import (
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
)

// Living is the surface effects need from pocketmine\entity\Living.
type Living interface {
	entityevent.Entity

	GetEffects() *EffectManager

	GetHealth() float64
	SetHealth(amount float64)
	GetMaxHealth() int
	SetMaxHealth(amount int)
	GetAbsorption() float64
	SetAbsorption(absorption float64)
	Heal(source *entityevent.EntityRegainHealthEvent)
	Attack(source entityevent.DamageSource)

	SetInvisible(value bool)
	SetNameTagVisible(value bool)

	GetMotion() math.Vector3
	AddMotion(x, y, z float64)
	SetHasGravity(v bool)

	GetMovementSpeed() float64
	SetMovementSpeed(v float64, fit bool)
}

// HungerManager is the surface HungerEffect/SaturationEffect need from
// pocketmine\entity\HungerManager. *entity.HungerManager satisfies it.
type HungerManager interface {
	Exhaust(amount float64, cause int) float64
	AddFood(amount float64)
	AddSaturation(amount float64)
}

// HumanHungerManager reports the hunger manager of l when l is a Human (PHP's
// `$entity instanceof Human` + getHungerManager()). Installed by pocketmine/entity's init(); nil
// until then, in which case no entity is treated as a Human.
var HumanHungerManager func(l Living) (HungerManager, bool)

// IsPlayer reports whether l is a pocketmine\player\Player (PHP's `$entity instanceof Player`).
// Installed by pocketmine/player's init(); nil until then, in which case nothing is a Player.
var IsPlayer func(l Living) bool

// OwningEntityOf returns the owning entity of source (Entity::getOwningEntity), or nil. Installed
// by pocketmine/entity's init(); nil until then, in which case no source has an owner.
var OwningEntityOf func(source entityevent.Entity) entityevent.Entity

func humanHungerManager(l Living) (HungerManager, bool) {
	if HumanHungerManager == nil {
		return nil, false
	}
	return HumanHungerManager(l)
}
