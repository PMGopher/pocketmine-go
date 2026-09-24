package effect

import (
	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// Effect is a port of pocketmine\entity\effect\Effect's public surface. Effect types are compared
// by identity (PHP keys effect collections by spl_object_id), so every Effect is a pointer and the
// VanillaEffects getters always return the same instance.
//
// Concrete effect types embed EffectBase and call Init(self) - the same self-dispatch pattern as
// block.Block - so EffectBase.CanTick reaches a subtype's GetApplyInterval override.
type Effect interface {
	// GetName returns the translation key or name of the effect (PHP's Translatable|string): either
	// a *lang.Translatable or a string.
	GetName() any
	GetColor() color.Color
	IsBad() bool
	GetDefaultDuration() int
	HasBubbles() bool
	CanTick(instance *EffectInstance) bool
	GetApplyInterval(instance *EffectInstance) int
	ApplyEffect(entity Living, instance *EffectInstance, potency float64, source entityevent.Entity)
	Add(entity Living, instance *EffectInstance)
	Remove(entity Living, instance *EffectInstance)
}

// EffectBase is a port of pocketmine\entity\effect\Effect's state and default method bodies.
type EffectBase struct {
	self Effect

	name            any
	color           color.Color
	bad             bool
	defaultDuration int
	hasBubbles      bool
}

// NewEffect is a port of Effect::__construct for a plain Effect with no special behaviour. name is
// a *lang.Translatable or a string; PHP's defaults are bad=false, defaultDuration=600,
// hasBubbles=true.
func NewEffect(name any, c color.Color, bad bool, defaultDuration int, hasBubbles bool) *EffectBase {
	e := &EffectBase{}
	e.initBase(name, c, bad, defaultDuration, hasBubbles)
	e.Init(e)
	return e
}

func (e *EffectBase) initBase(name any, c color.Color, bad bool, defaultDuration int, hasBubbles bool) {
	e.name = name
	e.color = c
	e.bad = bad
	e.defaultDuration = defaultDuration
	e.hasBubbles = hasBubbles
}

// Init finishes constructing e given the concrete type embedding it. Must be called exactly once.
func (e *EffectBase) Init(self Effect) { e.self = self }

func (e *EffectBase) GetName() any { return e.name }

func (e *EffectBase) GetColor() color.Color { return e.color }

// IsBad returns whether this effect is harmful.
func (e *EffectBase) IsBad() bool { return e.bad }

// GetDefaultDuration returns the default duration (in ticks) this effect will apply for if a
// duration is not specified.
func (e *EffectBase) GetDefaultDuration() int { return e.defaultDuration }

// HasBubbles returns whether this effect will give the subject potion bubbles.
func (e *EffectBase) HasBubbles() bool { return e.hasBubbles }

// CanTick is a port of Effect::canTick: whether the effect will do something on the current tick.
func (e *EffectBase) CanTick(instance *EffectInstance) bool {
	interval := e.self.GetApplyInterval(instance)
	return interval > 0 && instance.GetDuration()%interval == 0
}

// GetApplyInterval returns the number of ticks between applications of the effect.
func (e *EffectBase) GetApplyInterval(instance *EffectInstance) int { return 0 }

// ApplyEffect applies effect results to an entity. This will not be called unless CanTick returns
// true.
func (e *EffectBase) ApplyEffect(entity Living, instance *EffectInstance, potency float64, source entityevent.Entity) {
}

// Add applies effects to the entity when the effect is first added.
func (e *EffectBase) Add(entity Living, instance *EffectInstance) {}

// Remove removes the effect from the entity, resetting any changed values back to their original
// defaults.
func (e *EffectBase) Remove(entity Living, instance *EffectInstance) {}
