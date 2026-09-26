package block

import (
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// LavaCauldron is a port of pocketmine\block\LavaCauldron. writeStateToWorld's tile reset isn't
// ported (no Cauldron tile yet).
type LavaCauldron struct {
	FillableCauldron
}

func NewLavaCauldron(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *LavaCauldron {
	l := &LavaCauldron{newFillableCauldron(idInfo, name, typeInfo)}
	l.Init(l)
	return l
}

func (l *LavaCauldron) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *LavaCauldron) GetLightLevel() int { return 15 }

func (l *LavaCauldron) GetFillSound() sound.Sound { return sound.CauldronFillLavaSound{} }

func (l *LavaCauldron) GetEmptySound() sound.Sound { return sound.CauldronEmptyLavaSound{} }

// OnInteract is a port of LavaCauldron::onInteract.
func (l *LavaCauldron) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	switch item.GetTypeId() {
	case itemTypeIDsBucket:
		l.removeFillLevels(FillableCauldronMaxFillLevel, item, vanillaItem("lava_bucket"), returnedItems)
	case itemTypeIDsPowderSnowBucket, itemTypeIDsWaterBucket:
		l.mix(item, vanillaItem("bucket"), returnedItems)
	case itemTypeIDsLingeringPotion, itemTypeIDsPotion, itemTypeIDsSplashPotion:
		l.mix(item, vanillaItem("glass_bottle"), returnedItems)
	}
	return true
}

func (l *LavaCauldron) HasEntityCollision() bool { return true }

// OnEntityInside is a port of LavaCauldron::onEntityInside.
func (l *LavaCauldron) OnEntityInside(entity Entity) bool {
	dmgEv := entityevent.NewEntityDamageByBlockEvent(l.self, entity, entityevent.CauseLava, 4, nil)
	entity.Attack(dmgEv)

	combustEv := entityevent.NewEntityCombustByBlockEvent(l.self, entity, 8)
	combustEv.Call()
	if !combustEv.IsCancelled() {
		entity.SetOnFire(combustEv.GetDuration())
	}
	return true
}
