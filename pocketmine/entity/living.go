package entity

import "pocketmine-go/pocketmine/math"

// Living is a port of pocketmine\entity\Living, narrowed to the same slice as Entity (see
// entity.go's package doc comment) - IsSneaking/IsLiving (needed by block.Living's local
// interface) plus everything inherited from Entity (health, fire, motion, Attack). Living::attack
// isn't overridden here: the real version layers knockback, armor-enchantment damage reduction,
// and effect-based cancellation on top of Entity::attack, none of which are ported yet, so this
// type inherits the base Entity.Attack via embedding.
type Living struct {
	Entity

	sneaking bool
}

func NewLiving(position math.Vector3, boundingBox math.AxisAlignedBB) *Living {
	l := &Living{
		Entity: Entity{position: position, boundingBox: boundingBox, health: 20, maxHealth: 20},
	}
	l.Init(l)
	return l
}

// IsLiving satisfies block.Living's local interface marker.
func (l *Living) IsLiving() bool { return true }

func (l *Living) IsSneaking() bool { return l.sneaking }

func (l *Living) SetSneaking(sneaking bool) { l.sneaking = sneaking }
