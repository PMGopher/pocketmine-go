package effect

import (
	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
	playerevent "pocketmine-go/pocketmine/event/player"
)

// HungerEffect is a port of pocketmine\entity\effect\HungerEffect.
type HungerEffect struct{ EffectBase }

func NewHungerEffect(name any, c color.Color, bad bool) *HungerEffect {
	e := &HungerEffect{}
	e.initBase(name, c, bad, 600, true)
	e.Init(e)
	return e
}

func (e *HungerEffect) GetApplyInterval(instance *EffectInstance) int { return 1 }

func (e *HungerEffect) ApplyEffect(entity Living, instance *EffectInstance, potency float64, source entityevent.Entity) {
	if manager, ok := humanHungerManager(entity); ok {
		manager.Exhaust(0.1*float64(instance.GetEffectLevel()), playerevent.ExhaustCausePotion)
	}
}
