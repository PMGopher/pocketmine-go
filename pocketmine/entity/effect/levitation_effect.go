package effect

import (
	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// LevitationEffect is a port of pocketmine\entity\effect\LevitationEffect.
type LevitationEffect struct{ EffectBase }

func NewLevitationEffect(name any, c color.Color) *LevitationEffect {
	e := &LevitationEffect{}
	e.initBase(name, c, false, 600, true)
	e.Init(e)
	return e
}

func (e *LevitationEffect) GetApplyInterval(instance *EffectInstance) int { return 1 }

func (e *LevitationEffect) ApplyEffect(entity Living, instance *EffectInstance, potency float64, source entityevent.Entity) {
	if IsPlayer == nil || !IsPlayer(entity) { //TODO: ugly hack, player motion isn't updated properly by the server yet :(
		entity.AddMotion(0, (float64(instance.GetEffectLevel())/20-entity.GetMotion().Y)/5, 0)
	}
}

func (e *LevitationEffect) Add(entity Living, instance *EffectInstance) {
	entity.SetHasGravity(false)
}

func (e *LevitationEffect) Remove(entity Living, instance *EffectInstance) {
	entity.SetHasGravity(true)
}
