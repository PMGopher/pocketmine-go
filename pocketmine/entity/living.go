package entity

import (
	stdmath "math"

	"pocketmine-go/pocketmine/math"
)

// DefaultKnockbackForce/DefaultKnockbackVerticalLimit mirror Living::DEFAULT_KNOCKBACK_FORCE/
// DEFAULT_KNOCKBACK_VERTICAL_LIMIT.
const (
	DefaultKnockbackForce         = 0.4
	DefaultKnockbackVerticalLimit = 0.4
)

// Living is a port of pocketmine\entity\Living, narrowed to the same slice as Entity (see
// entity.go's package doc comment) - IsSneaking/IsLiving (needed by block.Living's local
// interface) plus everything inherited from Entity (health, fire, motion). Living.Attack layers
// real knockback on top of the base Entity.Attack - see its own doc comment for what's still
// deliberately left out (armor-enchantment damage reduction, effect-based cancellation, the
// attackTime/noDamageTicks invulnerability-cooldown gating, and the HurtAnimation/CriticalHit/
// MagicHit broadcasts, none of which have a subsystem anywhere else in this port yet either).
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

// Attack is a port of a slice of Living::attack: calls the base Entity.Attack first, then (if not
// cancelled, and the source is an EntityDamageByEntityEvent with a real damager) applies real
// knockback away from the damager - see the type's own doc comment for what's deliberately left
// out (the noDamageTicks/attackTime gating means this port applies knockback on every successful
// hit rather than only "cold" ones, a simplification rather than a guess: no per-tick
// invulnerability countdown exists anywhere in this port to gate on).
func (l *Living) Attack(source DamageSource) {
	l.Entity.Attack(source)
	if source.IsCancelled() {
		return
	}

	if byEntity, ok := source.(*EntityDamageByEntityEvent); ok && byEntity.damager != nil {
		dx := l.position.X - byEntity.damager.GetPosition().X
		dz := l.position.Z - byEntity.damager.GetPosition().Z
		l.KnockBack(dx, dz, byEntity.GetKnockBack(), byEntity.GetVerticalKnockBackLimit())
	}
}

// CalculateFallDamage is a port of Living::calculateFallDamage, minus the Jump Boost effect-level
// subtraction (no EffectManager exists in this port yet - matches every other "no EffectManager"
// gap elsewhere). Exported so callers driving fall tracking from outside this package (see
// player.Player's own fall-damage wiring) can compute the same damage value onHitGround already
// used internally, without duplicating the formula.
func (l *Living) CalculateFallDamage(fallDistance float64) float64 {
	return stdmath.Ceil(fallDistance - 3)
}

// onHitGround is a port of a slice of Living::onHitGround: applies real fall damage via the same
// Attack/EntityDamageEvent pipeline as any other damage source.
//
// Not ported: the fallBlock.onEntityLand() bouncy-block hook (e.g. slime blocks) - no such
// per-block landing-behavior hook exists anywhere in the block package yet, a whole separate
// subsystem, not a shortcut of convenience.
func (l *Living) onHitGround() *float64 {
	damage := l.CalculateFallDamage(l.fallDistance)
	if damage > 0 {
		ev := NewEntityDamageEvent(l, EntityDamageCauseFall, damage, nil)
		l.Attack(ev)
	}
	return nil
}

// KnockBack is a port of Living::knockBack, minus the knockback-resistance-attribute roll (no
// AttributeMap exists in this port yet - matches every other "no AttributeMap" gap elsewhere).
func (l *Living) KnockBack(x, z, force, verticalLimit float64) {
	f := stdmath.Sqrt(x*x + z*z)
	if f <= 0 {
		return
	}
	f = 1 / f

	motion := l.GetMotion()
	motionX := motion.X/2 + x*f*force
	motionY := motion.Y/2 + force
	motionZ := motion.Z/2 + z*f*force

	if motionY > verticalLimit {
		motionY = verticalLimit
	}

	l.SetMotion(math.NewVector3(motionX, motionY, motionZ))
}
