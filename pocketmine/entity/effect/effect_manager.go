package effect

// EffectManager is a port of pocketmine\entity\effect\EffectManager - the EffectCollection owned by
// a Living, which additionally fires EntityEffectAddEvent/EntityEffectRemoveEvent and calls each
// effect type's Add/Remove/ApplyEffect hooks on the entity.
type EffectManager struct {
	EffectCollection

	entity Living
}

// NewEffectManager is a port of EffectManager::__construct.
func NewEffectManager(entity Living) *EffectManager {
	m := &EffectManager{entity: entity}
	m.initCollection()
	m.removeImpl = m.Remove
	return m
}

// Remove is a port of EffectManager::remove: removes an effect from the mob, unless an
// EntityEffectRemoveEvent listener cancels it (in which case the add hooks are re-run so the
// client stays in sync).
func (m *EffectManager) Remove(effectType Effect) {
	effect, ok := m.effects[effectType]
	if !ok {
		return
	}
	ev := NewEntityEffectRemoveEvent(m.entity, effect)
	ev.Call()
	if ev.IsCancelled() {
		for hook := range m.effectAddHooks.All() {
			(*hook)(ev.GetEffect(), true)
		}
		return
	}

	effect.GetType().Remove(m.entity, effect)
	m.remove(effectType)
}

// Add is a port of EffectManager::add: adds an effect to the mob. If a weaker effect of the same
// type is already applied, it will be replaced. If a weaker or equal-strength effect is already
// applied but has a shorter duration, it will be replaced. Returns whether the effect has been
// successfully applied.
func (m *EffectManager) Add(effect *EffectInstance) bool {
	oldEffect := m.effects[effect.GetType()]

	ev := NewEntityEffectAddEvent(m.entity, effect, oldEffect)
	if !m.CanAdd(effect) {
		ev.Cancel()
	}

	ev.Call()
	if ev.IsCancelled() {
		return false
	}

	if oldEffect != nil {
		oldEffect.GetType().Remove(m.entity, oldEffect)
	}

	effect.GetType().Add(m.entity, effect)

	return m.add(effect)
}

// Clear removes all effects from the mob, through Remove (so each removal can be cancelled).
func (m *EffectManager) Clear() {
	for _, effect := range m.All() {
		m.Remove(effect.GetType())
	}
}

// Tick is a port of EffectManager::tick. Returns whether any effects are still active.
func (m *EffectManager) Tick(tickDiff int) bool {
	for _, instance := range m.All() {
		effectType := instance.GetType()
		if effectType.CanTick(instance) {
			effectType.ApplyEffect(m.entity, instance, 1.0, nil)
		}
		instance.DecreaseDuration(tickDiff)
		if instance.HasExpired() {
			m.Remove(instance.GetType())
		}
	}

	return len(m.effects) > 0
}
