package block

import (
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/color"
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
// Not ported yet: the dye, armour, banner and shulker box interactions (they need
// Dye/Armor/Banner item types this package can't see); the custom water colour they set is kept in
// the Cauldron tile.
type WaterCauldron struct {
	FillableCauldron

	customWaterColor *color.Color
}

func NewWaterCauldron(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *WaterCauldron {
	w := &WaterCauldron{FillableCauldron: newFillableCauldron(idInfo, name, typeInfo)}
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

// GetCustomWaterColor is a port of WaterCauldron::getCustomWaterColor (nil for none).
func (w *WaterCauldron) GetCustomWaterColor() *color.Color { return w.customWaterColor }

// SetCustomWaterColor is a port of WaterCauldron::setCustomWaterColor.
func (w *WaterCauldron) SetCustomWaterColor(customWaterColor *color.Color) *WaterCauldron {
	w.customWaterColor = customWaterColor
	return w
}

// ReadStateFromWorld is a port of WaterCauldron::readStateFromWorld.
func (w *WaterCauldron) ReadStateFromWorld() Behavior {
	if result := w.Block.ReadStateFromWorld(); result != w.self {
		return result
	}
	t, _ := w.tileAt()
	cauldronTile, _ := t.(*tile.Cauldron)
	if cauldronTile != nil {
		if potion, ok := cauldronTile.GetPotionItem().(Item); ok {
			//TODO: HACK! we keep potion cauldrons as a separate block type due to different behaviour, but in the
			//blockstate they are typically indistinguishable from water cauldrons. This hack converts cauldrons into
			//their appropriate type.
			potionCauldron := VanillaBlock("potion_cauldron").(*PotionCauldron)
			potionCauldron.SetFillLevel(w.FillLevel)
			potionCauldron.SetPotionItem(potion)
			return potionCauldron
		}
	}
	w.customWaterColor = nil
	if cauldronTile != nil {
		w.customWaterColor = cauldronTile.GetCustomWaterColor()
	}
	return w.self
}

// WriteStateToWorld is a port of WaterCauldron::writeStateToWorld.
func (w *WaterCauldron) WriteStateToWorld() {
	w.Block.WriteStateToWorld()
	if t, ok := w.tileAt(); ok {
		if cauldronTile, ok := t.(*tile.Cauldron); ok {
			cauldronTile.SetCustomWaterColor(w.customWaterColor)
			cauldronTile.SetPotionItem(nil)
		}
	}
}
