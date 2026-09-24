package entity

import (
	stdmath "math"

	"pocketmine-go/pocketmine/event"
)

// defaultFireChance mirrors Explosion::DEFAULT_FIRE_CHANCE (pocketmine/world can't be imported
// from here - see this package's doc comment).
const defaultFireChance = 1.0 / 3.0

// EntityPreExplodeEvent is a port of pocketmine\event\entity\EntityPreExplodeEvent - called before
// an entity explodes, when the explosion's impact has not yet been calculated.
type EntityPreExplodeEvent struct {
	EntityEvent
	event.CancellableTrait

	radius        float64
	fireChance    float64
	blockBreaking bool
}

// NewEntityPreExplodeEvent is a port of EntityPreExplodeEvent::__construct, panicking on the same
// invalid arguments as the PHP InvalidArgumentException.
func NewEntityPreExplodeEvent(entity Entity, radius, fireChance float64) *EntityPreExplodeEvent {
	if radius <= 0 {
		panic("Explosion radius must be positive")
	}
	checkFireChance(fireChance)
	return &EntityPreExplodeEvent{EntityEvent: EntityEvent{entity: entity}, radius: radius, fireChance: fireChance, blockBreaking: true}
}

func checkFireChance(fireChance float64) {
	if stdmath.IsNaN(fireChance) || stdmath.IsInf(fireChance, 0) {
		panic("fireChance cannot be NaN or infinite")
	}
	if fireChance < 0.0 || fireChance > 1.0 {
		panic("Fire chance must be between 0 and 1.")
	}
}

func (e *EntityPreExplodeEvent) Call() { event.Call(e) }

func (e *EntityPreExplodeEvent) GetRadius() float64 { return e.radius }

func (e *EntityPreExplodeEvent) SetRadius(radius float64) {
	if radius <= 0 {
		panic("Explosion radius must be positive")
	}
	e.radius = radius
}

func (e *EntityPreExplodeEvent) IsIncendiary() bool { return e.fireChance > 0 }

// SetIncendiary is a port of EntityPreExplodeEvent::setIncendiary.
func (e *EntityPreExplodeEvent) SetIncendiary(incendiary bool) {
	if !incendiary {
		e.fireChance = 0
	} else if e.fireChance <= 0 {
		e.fireChance = defaultFireChance
	}
}

func (e *EntityPreExplodeEvent) GetFireChance() float64 { return e.fireChance }

func (e *EntityPreExplodeEvent) SetFireChance(fireChance float64) {
	checkFireChance(fireChance)
	e.fireChance = fireChance
}

func (e *EntityPreExplodeEvent) IsBlockBreaking() bool { return e.blockBreaking }

func (e *EntityPreExplodeEvent) SetBlockBreaking(affectsBlocks bool) { e.blockBreaking = affectsBlocks }
