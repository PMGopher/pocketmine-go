package effect

import (
	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// RegenerationEffect is a port of pocketmine\entity\effect\RegenerationEffect.
type RegenerationEffect struct{ EffectBase }

func NewRegenerationEffect(name any, c color.Color) *RegenerationEffect {
	e := &RegenerationEffect{}
	e.initBase(name, c, false, 600, true)
	e.Init(e)
	return e
}

func (e *RegenerationEffect) GetApplyInterval(instance *EffectInstance) int {
	return max(1, 50>>instance.GetAmplifier())
}

func (e *RegenerationEffect) ApplyEffect(entity Living, instance *EffectInstance, potency float64, source entityevent.Entity) {
	if entity.GetHealth() < float64(entity.GetMaxHealth()) {
		ev := entityevent.NewEntityRegainHealthEvent(entity, 1, entityevent.RegainCauseMagic)
		entity.Heal(ev)
	}
}
