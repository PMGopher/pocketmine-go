package block

import (
	"pocketmine-go/pocketmine/entity/effect"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
)

// testEntity is a block-package test double for Entity: pocketmine/entity imports this package, so
// block tests can't use the real entity types. It mirrors the real Entity's observable behaviour
// that block tests assert on (damage through Attack, fire ticks, fall distance).
type testEntity struct {
	health          float64
	fallDistance    float64
	fireTicks       int
	onGround        bool
	motion          math.Vector3
	position        math.Vector3
	lastDamageCause entityevent.DamageSource
}

func newTestEntity() *testEntity { return &testEntity{health: 20} }

func (e *testEntity) GetID() int                           { return 1 }
func (e *testEntity) GetPosition() math.Vector3            { return e.position }
func (e *testEntity) IsClosed() bool                       { return false }
func (e *testEntity) ResetFallDistance()                   { e.fallDistance = 0 }
func (e *testEntity) SetOnGround(onGround bool)            { e.onGround = onGround }
func (e *testEntity) GetFallDistance() float64             { return e.fallDistance }
func (e *testEntity) SetFallDistance(fallDistance float64) { e.fallDistance = fallDistance }
func (e *testEntity) GetBoundingBox() math.AxisAlignedBB   { return math.OneAABB() }
func (e *testEntity) GetMotion() math.Vector3              { return e.motion }
func (e *testEntity) IsOnFire() bool                       { return e.fireTicks > 0 }
func (e *testEntity) Extinguish()                          { e.fireTicks = 0 }
func (e *testEntity) ExtinguishWithCause(cause int)        { e.fireTicks = 0 }
func (e *testEntity) CanBeMovedByCurrents() bool           { return true }
func (e *testEntity) GetHealth() float64                   { return e.health }
func (e *testEntity) GetFireTicks() int                    { return e.fireTicks }

func (e *testEntity) GetLastDamageCause() entityevent.DamageSource { return e.lastDamageCause }

// SetOnFire mirrors Entity::setOnFire.
func (e *testEntity) SetOnFire(seconds int) {
	if ticks := seconds * 20; ticks > e.fireTicks {
		e.fireTicks = ticks
	}
}

// Attack mirrors the base Entity::attack.
func (e *testEntity) Attack(source entityevent.DamageSource) {
	source.Call()
	if source.IsCancelled() {
		return
	}
	e.lastDamageCause = source
	e.health -= source.GetFinalDamage()
}

// testLiving is the Living counterpart of testEntity.
type testLiving struct {
	testEntity
	sneaking bool
}

func newTestLiving() *testLiving { return &testLiving{testEntity: testEntity{health: 20}} }

func (l *testLiving) IsLiving() bool            { return true }
func (l *testLiving) IsSneaking() bool          { return l.sneaking }
func (l *testLiving) SetSneaking(sneaking bool) { l.sneaking = sneaking }

var (
	_ Entity = (*testEntity)(nil)
	_ Living = (*testLiving)(nil)
)

// testConsumer is an effect.Living test double for Consumable.OnConsume calls.
type testConsumer struct {
	testLiving
}

func (c *testConsumer) GetEffects() *effect.EffectManager                { return nil }
func (c *testConsumer) SetHealth(amount float64)                         { c.health = amount }
func (c *testConsumer) GetMaxHealth() int                                { return 20 }
func (c *testConsumer) SetMaxHealth(amount int)                          {}
func (c *testConsumer) GetAbsorption() float64                           { return 0 }
func (c *testConsumer) SetAbsorption(absorption float64)                 {}
func (c *testConsumer) Heal(source *entityevent.EntityRegainHealthEvent) {}
func (c *testConsumer) SetInvisible(value bool)                          {}
func (c *testConsumer) SetNameTagVisible(value bool)                     {}
func (c *testConsumer) AddMotion(x, y, z float64)                        {}
func (c *testConsumer) SetHasGravity(v bool)                             {}
func (c *testConsumer) GetMovementSpeed() float64                        { return 0.1 }
func (c *testConsumer) SetMovementSpeed(v float64, fit bool)             {}

var _ effect.Living = (*testConsumer)(nil)
