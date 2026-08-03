// Package entity is a port of pocketmine\entity, started from a deliberately narrow slice: just
// the self-contained physical/vital state (position, motion, bounding box, fall distance, fire,
// health) and the Attack/damage-event pipeline needed to make the block package's existing
// Entity/Living/Player local interfaces (see block/world.go) satisfiable by a real type, and to
// un-gap the several OnEntityInside/OnAttack methods across the block package that were
// previously stubbed out for lack of Entity.Attack.
//
// NOT ported in this first slice — Entity.php and Living.php are enormous (1700+ and 1000+ lines
// respectively) and cover far more than fits in one pass:
//   - World integration: entities aren't tracked by a world, spawned, or ticked. SetPosition is a
//     bare field write, not the full addEntity/removeEntity/onEntityMoved dance.
//   - Movement simulation: gravity, drag, block collision, updateMovement's network-dirty
//     bookkeeping, checkBlockIntersections (which is what would actually call
//     Behavior.OnEntityInside from a real game loop).
//   - Network sync, NBT save/load, viewers/spawning-to-players, effects (EffectManager),
//     attributes (AttributeMap) beyond a bare Health/MaxHealth pair, armor, name tags.
//   - Living.Attack's knockback/armor-enchantment/effect-modifier pipeline — only the simpler
//     base Entity.Attack (fire-immunity cancel check, call the event, reduce health) is ported;
//     Living doesn't override it yet, so it inherits the base behavior via embedding.
//
// Every event type this slice needs (EntityDamageEvent and friends) lives in this package too,
// rather than mirroring PHP's separate pocketmine\event\entity namespace — those events hold an
// Entity reference and Entity needs to construct/fire them (extinguish(), setMotion()), so
// keeping them in one package sidesteps an import cycle a separate package would create (same
// reasoning as tile.FurnaceType living directly in the tile package instead of a dedicated
// crafting package).
package entity

import (
	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/math"
)

// entityShaper is the self-dispatch interface letting Entity's own methods reach a concrete
// subtype's overrides — same shape as blockShaper/tileShaper elsewhere in this port. IsFireProof,
// onDeath, and onHitGround are the only overridable hooks this slice's ported methods call
// internally.
type entityShaper interface {
	IsFireProof() bool
	onDeath()
	onHitGround() *float64
}

// Entity is a port of pocketmine\entity\Entity (see package doc comment for scope).
type Entity struct {
	self entityShaper

	position     math.Vector3
	motion       math.Vector3
	boundingBox  math.AxisAlignedBB
	onGround     bool
	fallDistance float64
	fireTicks    int

	health          float64
	maxHealth       int
	lastDamageCause DamageSource

	closed bool
}

// NewEntity constructs an Entity at the given position with the given bounding box. Real Entity
// construction in PHP derives the bounding box from EntitySizeInfo and a Location - both skipped
// here (see package doc comment), so the caller provides the bounding box directly.
func NewEntity(position math.Vector3, boundingBox math.AxisAlignedBB) *Entity {
	e := &Entity{position: position, boundingBox: boundingBox, health: 20, maxHealth: 20}
	e.Init(e)
	return e
}

// Init finishes constructing e, given self (the concrete entity type embedding this Entity). Must
// be called exactly once, immediately after construction — same convention as block.Block.Init.
func (e *Entity) Init(self entityShaper) { e.self = self }

func (e *Entity) rebind(self entityShaper) { e.self = self }

func (e *Entity) IsClosed() bool { return e.closed }

// GetPosition is a port of Entity::getPosition, simplified to a bare Vector3 (no World
// reference) — matching the shape block.Entity's local interface already committed to.
func (e *Entity) GetPosition() math.Vector3 { return e.position }

// SetPosition is a simplified port of Entity::setPosition: just the field write. The real
// version's world-transfer bookkeeping (despawnFromAll/removeEntity/addEntity/onEntityMoved) and
// bounding-box recalculation need a world that tracks entities, not ported yet.
func (e *Entity) SetPosition(pos math.Vector3) bool {
	if e.closed {
		return false
	}
	e.position = pos
	return true
}

func (e *Entity) GetBoundingBox() math.AxisAlignedBB { return e.boundingBox }

func (e *Entity) GetMotion() math.Vector3 { return e.motion }

// SetMotion is a port of Entity::setMotion, minus the justCreated-skip and updateMovement's
// network-dirty bookkeeping (movement simulation isn't ported). The cancellable EntityMotionEvent
// is fully real.
func (e *Entity) SetMotion(motion math.Vector3) bool {
	ev := NewEntityMotionEvent(e, motion)
	Call(ev)
	if ev.IsCancelled() {
		return false
	}
	e.motion = motion
	return true
}

// AddMotion is a port of Entity::addMotion.
func (e *Entity) AddMotion(x, y, z float64) {
	e.motion = e.motion.Add(x, y, z)
}

func (e *Entity) IsOnGround() bool { return e.onGround }

func (e *Entity) SetOnGround(onGround bool) { e.onGround = onGround }

func (e *Entity) GetFallDistance() float64 { return e.fallDistance }

func (e *Entity) SetFallDistance(fallDistance float64) { e.fallDistance = fallDistance }

func (e *Entity) ResetFallDistance() { e.fallDistance = 0 }

// onHitGround is a port of Entity::onHitGround - the base no-op override (only Living overrides
// this, to apply fall damage).
func (e *Entity) onHitGround() *float64 { return nil }

// UpdateFallState is a port of Entity::updateFallState. distanceThisTick is the raw Y delta of this
// movement step (newY - oldY, negative while falling) - real PHP computes this from its own
// physics simulation (Entity::move); this port has no server-side physics (position comes from the
// client's own PlayerAuthInput reports instead - see cmd/pocketmine-go's own doc comments on why),
// so the caller supplies it directly. Returns the new vertical velocity onHitGround reports (nil
// for every concrete type in this port so far - see Living.onHitGround's own doc comment on the
// bouncy-block hook this doesn't implement).
func (e *Entity) UpdateFallState(distanceThisTick float64, onGround bool) *float64 {
	if distanceThisTick < e.fallDistance {
		e.fallDistance -= distanceThisTick
	} else {
		e.fallDistance = 0
	}
	if onGround && e.fallDistance > 0 {
		newVerticalVelocity := e.self.onHitGround()
		e.ResetFallDistance()
		return newVerticalVelocity
	}
	return nil
}

// CanBeMovedByCurrents is a port of Entity::canBeMovedByCurrents.
func (e *Entity) CanBeMovedByCurrents() bool { return true }

func (e *Entity) IsOnFire() bool { return e.fireTicks > 0 }

// SetOnFire is a port of Entity::setOnFire.
func (e *Entity) SetOnFire(seconds int) {
	ticks := seconds * 20
	if ticks > e.GetFireTicks() {
		e.SetFireTicks(ticks)
	}
}

func (e *Entity) GetFireTicks() int { return e.fireTicks }

// SetFireTicks is a port of Entity::setFireTicks. Panics on a negative value, matching the PHP
// original's InvalidArgumentException (same convention as e.g. block.Hopper.SetFacing).
func (e *Entity) SetFireTicks(fireTicks int) {
	if fireTicks < 0 {
		panic("Fire ticks cannot be negative")
	}
	if fireTicks > binaryutils.Int16Max {
		fireTicks = binaryutils.Int16Max
	}
	if !e.self.IsFireProof() {
		e.fireTicks = fireTicks
	}
}

// Extinguish is a port of Entity::extinguish() with its default $cause argument
// (EntityExtinguishEvent::CAUSE_CUSTOM) — Go has no default parameters, and this exact zero-arg
// signature is what block.Entity's local interface already commits to. Use ExtinguishWithCause
// for the other causes.
func (e *Entity) Extinguish() { e.ExtinguishWithCause(EntityExtinguishCauseCustom) }

// ExtinguishWithCause is a port of Entity::extinguish($cause).
func (e *Entity) ExtinguishWithCause(cause int) {
	ev := NewEntityExtinguishEvent(e, cause)
	Call(ev)
	e.fireTicks = 0
}

// IsFireProof is a port of Entity::isFireProof - the default (false); concrete entity types
// override it via self-dispatch.
func (e *Entity) IsFireProof() bool { return false }

func (e *Entity) IsAlive() bool { return e.health > 0 }

func (e *Entity) GetHealth() float64 { return e.health }

// SetHealth is a port of Entity::setHealth.
func (e *Entity) SetHealth(amount float64) {
	if amount == e.health {
		return
	}
	if amount <= 0 {
		if e.IsAlive() {
			e.Kill()
		}
	} else if amount <= float64(e.GetMaxHealth()) || amount < e.health {
		e.health = amount
	} else {
		e.health = float64(e.GetMaxHealth())
	}
}

func (e *Entity) GetMaxHealth() int { return e.maxHealth }

func (e *Entity) SetMaxHealth(amount int) { e.maxHealth = amount }

func (e *Entity) SetLastDamageCause(source DamageSource) { e.lastDamageCause = source }

func (e *Entity) GetLastDamageCause() DamageSource { return e.lastDamageCause }

// Kill is a port of Entity::kill, minus scheduleUpdate (tick scheduling isn't ported).
func (e *Entity) Kill() {
	if e.IsAlive() {
		e.health = 0
		e.self.onDeath()
	}
}

// onDeath is a port of Entity::onDeath - a no-op hook overridden by concrete subtypes (protected
// in PHP; unexported here since nothing outside this package needs to call it directly).
func (e *Entity) onDeath() {}

// Attack is a port of the base Entity::attack (not Living::attack, which layers knockback/armor/
// effects on top - not ported, see package doc comment). Fire-immune entities cancel fire-cause
// damage; otherwise the event is called, and unless cancelled, health is reduced by the event's
// final damage.
func (e *Entity) Attack(source DamageSource) {
	if e.self.IsFireProof() && isFireCause(source.GetCause()) {
		source.Cancel()
	}
	source.Call()
	if source.IsCancelled() {
		return
	}
	e.SetLastDamageCause(source)
	e.SetHealth(e.GetHealth() - source.GetFinalDamage())
}

func isFireCause(cause int) bool {
	return cause == EntityDamageCauseFire || cause == EntityDamageCauseFireTick || cause == EntityDamageCauseLava
}
