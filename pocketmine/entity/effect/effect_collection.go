package effect

import (
	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/utils"
)

// EffectAddHook is called whenever an effect is added to an EffectCollection (PHP's effect add
// hook closure). replacesOldEffect is whether an existing effect of the same type was replaced.
type EffectAddHook func(effect *EffectInstance, replacesOldEffect bool)

// EffectRemoveHook is called whenever an effect is removed from an EffectCollection.
type EffectRemoveHook func(effect *EffectInstance)

// EffectCollection is a port of pocketmine\entity\effect\EffectCollection.
//
// PHP keys effects by spl_object_id($effectType) in an insertion-ordered array; Go maps aren't
// ordered, so the insertion order is tracked separately (it's observable through All(), e.g. the
// order effects are saved to NBT and ticked in).
//
// Hooks are stored as pointers so the same identity-based add/remove semantics as PHP's
// ObjectSet<Closure> apply: keep the pointer you added to be able to remove it again.
type EffectCollection struct {
	effects map[Effect]*EffectInstance
	order   []Effect

	effectAddHooks    *utils.ObjectSet[*EffectAddHook]
	effectRemoveHooks *utils.ObjectSet[*EffectRemoveHook]

	bubbleColor        color.Color
	onlyAmbientEffects bool

	effectFilterForBubbles func(e *EffectInstance) bool

	// removeImpl/addImpl are the self-dispatch hooks letting Clear() and Add() reach EffectManager's
	// overrides (PHP's $this->remove() inside EffectCollection::clear()).
	removeImpl func(effectType Effect)
}

// NewEffectCollection is a port of EffectCollection::__construct.
func NewEffectCollection() *EffectCollection {
	c := &EffectCollection{}
	c.initCollection()
	return c
}

func (c *EffectCollection) initCollection() {
	c.effects = map[Effect]*EffectInstance{}
	c.bubbleColor = color.NewColor(0, 0, 0, 0)
	c.effectAddHooks = utils.NewObjectSet[*EffectAddHook]()
	c.effectRemoveHooks = utils.NewObjectSet[*EffectRemoveHook]()
	c.SetEffectFilterForBubbles(func(e *EffectInstance) bool { return e.IsVisible() && e.GetType().HasBubbles() })
	c.removeImpl = c.remove
}

// All returns an array of effects currently active on the mob, in the order they were added.
func (c *EffectCollection) All() []*EffectInstance {
	result := make([]*EffectInstance, 0, len(c.order))
	for _, t := range c.order {
		result = append(result, c.effects[t])
	}
	return result
}

// Clear removes all effects.
func (c *EffectCollection) Clear() {
	for _, effect := range c.All() {
		c.removeImpl(effect.GetType())
	}
}

// Remove removes an effect from the collection.
func (c *EffectCollection) Remove(effectType Effect) { c.removeImpl(effectType) }

func (c *EffectCollection) remove(effectType Effect) {
	effect, ok := c.effects[effectType]
	if !ok {
		return
	}
	c.delete(effectType)
	for hook := range c.effectRemoveHooks.All() {
		(*hook)(effect)
	}
	c.recalculateEffectColor()
}

func (c *EffectCollection) delete(effectType Effect) {
	delete(c.effects, effectType)
	for i, t := range c.order {
		if t == effectType {
			c.order = append(c.order[:i:i], c.order[i+1:]...)
			break
		}
	}
}

// Get returns the effect instance applied to this collection, or nil if not applied.
func (c *EffectCollection) Get(effect Effect) *EffectInstance { return c.effects[effect] }

// Has returns whether the specified effect is active.
func (c *EffectCollection) Has(effect Effect) bool {
	_, ok := c.effects[effect]
	return ok
}

// CanAdd is a port of EffectCollection::canAdd: an effect can't replace one of the same type with
// a higher amplifier, or the same amplifier and a longer duration.
func (c *EffectCollection) CanAdd(effect *EffectInstance) bool {
	if oldEffect, ok := c.effects[effect.GetType()]; ok {
		amplifier := absInt(effect.GetAmplifier())
		if amplifier < oldEffect.GetAmplifier() ||
			(amplifier == absInt(oldEffect.GetAmplifier()) && effect.GetDuration() < oldEffect.GetDuration()) {
			return false
		}
	}
	return true
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// Add is a port of EffectCollection::add: adds an effect to the collection, replacing a weaker
// effect of the same type. Returns whether the effect was added.
func (c *EffectCollection) Add(effect *EffectInstance) bool { return c.add(effect) }

func (c *EffectCollection) add(effect *EffectInstance) bool {
	if !c.CanAdd(effect) {
		return false
	}
	effectType := effect.GetType()
	_, replacesOldEffect := c.effects[effectType]
	if !replacesOldEffect {
		c.order = append(c.order, effectType)
	}
	c.effects[effectType] = effect

	for hook := range c.effectAddHooks.All() {
		(*hook)(effect, replacesOldEffect)
	}

	c.recalculateEffectColor()
	return true
}

// SetEffectFilterForBubbles sets a filter that decides which effect instances contribute to the
// bubble color.
func (c *EffectCollection) SetEffectFilterForBubbles(filter func(e *EffectInstance) bool) {
	c.effectFilterForBubbles = filter
}

// recalculateEffectColor is a port of EffectCollection::recalculateEffectColor.
func (c *EffectCollection) recalculateEffectColor() {
	var colors []color.Color
	ambient := true
	for _, effect := range c.All() {
		if c.effectFilterForBubbles(effect) {
			level := effect.GetEffectLevel()
			col := effect.GetColor()
			for i := 0; i < level; i++ {
				colors = append(colors, col)
			}
			if !effect.IsAmbient() {
				ambient = false
			}
		}
	}

	if len(colors) > 0 {
		c.bubbleColor = color.Mix(colors[0], colors[1:]...)
		c.onlyAmbientEffects = ambient
	} else {
		c.bubbleColor = color.NewColor(0, 0, 0, 0)
		c.onlyAmbientEffects = false
	}
}

func (c *EffectCollection) GetBubbleColor() color.Color { return c.bubbleColor }

func (c *EffectCollection) HasOnlyAmbientEffects() bool { return c.onlyAmbientEffects }

func (c *EffectCollection) GetEffectAddHooks() *utils.ObjectSet[*EffectAddHook] {
	return c.effectAddHooks
}

func (c *EffectCollection) GetEffectRemoveHooks() *utils.ObjectSet[*EffectRemoveHook] {
	return c.effectRemoveHooks
}
