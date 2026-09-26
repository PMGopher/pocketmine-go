package block

import (
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

const (
	WaterCauldronWaterBottleFillAmount    = 2
	WaterCauldronDyeArmorUseAmount        = 1
	WaterCauldronCleanArmorUseAmount      = 1
	WaterCauldronCleanBannerUseAmount     = 1
	WaterCauldronCleanShulkerBoxUseAmount = 1
	//TODO: I'm not sure if this was intended to be 2 (to match java) but in Bedrock you can extinguish yourself 6 times ...
	WaterCauldronEntityExtinguishUseAmount = 1
)

// WaterCauldron is a port of pocketmine\block\WaterCauldron.
//
// Not ported yet: the custom water colour (it lives in the Cauldron tile, which isn't ported, as
// do readStateFromWorld's potion-cauldron conversion and writeStateToWorld), and the dye, armour,
// banner and shulker box interactions that depend on it (they need Dye/Armor/Banner item types
// this package can't see).
type WaterCauldron struct {
	FillableCauldron
}

func NewWaterCauldron(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *WaterCauldron {
	w := &WaterCauldron{newFillableCauldron(idInfo, name, typeInfo)}
	w.Init(w)
	return w
}

func (w *WaterCauldron) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WaterCauldron) GetFillSound() sound.Sound { return sound.CauldronFillWaterSound{} }

func (w *WaterCauldron) GetEmptySound() sound.Sound { return sound.CauldronEmptyWaterSound{} }

// OnInteract is a port of WaterCauldron::onInteract (see the type's doc comment for what's missing).
func (w *WaterCauldron) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if potion, ok := item.(waterPotionChecker); ok {
		if potion.IsWaterPotion() {
			w.addFillLevels(WaterCauldronWaterBottleFillAmount, item, vanillaItem("glass_bottle"), returnedItems)
		} else {
			w.mix(item, vanillaItem("glass_bottle"), returnedItems)
		}
		return true
	}
	switch item.GetTypeId() {
	case itemTypeIDsWaterBucket:
		w.addFillLevels(FillableCauldronMaxFillLevel, item, vanillaItem("bucket"), returnedItems)
	case itemTypeIDsBucket:
		w.removeFillLevels(FillableCauldronMaxFillLevel, item, vanillaItem("water_bucket"), returnedItems)
	case itemTypeIDsGlassBottle:
		w.removeFillLevels(WaterCauldronWaterBottleFillAmount, item, vanillaItem("water_potion"), returnedItems)
	case itemTypeIDsLavaBucket, itemTypeIDsPowderSnowBucket:
		w.mix(item, vanillaItem("bucket"), returnedItems)
	}
	return true
}

func (w *WaterCauldron) HasEntityCollision() bool { return true }

// OnEntityInside is a port of WaterCauldron::onEntityInside: burning entities are put out.
func (w *WaterCauldron) OnEntityInside(entity Entity) bool {
	if entity.IsOnFire() {
		entity.ExtinguishWithCause(entityevent.ExtinguishCauseWaterCauldron)
		//TODO: particles

		if world, err := w.position.GetWorld(); err == nil {
			_ = world.SetBlock(w.position, w.withFillLevel(w.FillLevel-WaterCauldronEntityExtinguishUseAmount))
		}
	}
	return true
}

// OnNearbyBlockChange is a port of WaterCauldron::onNearbyBlockChange: water above refills it.
func (w *WaterCauldron) OnNearbyBlockChange() {
	if w.FillLevel < FillableCauldronMaxFillLevel {
		world, err := w.position.GetWorld()
		if err != nil {
			return
		}
		if w.GetSide(math.Up, 1).GetTypeId() == WATER {
			w.SetFillLevel(FillableCauldronMaxFillLevel)
			_ = world.SetBlock(w.position, w)
			world.AddSound(w.position.Add(0.5, 0.5, 0.5), w.GetFillSound())
		}
	}
}
