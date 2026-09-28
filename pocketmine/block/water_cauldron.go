package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
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

// Items the water cauldron dyes and cleans. This package can't import item, so these are the
// parts of item.Dye, item.Armor, item.Banner and item.Item it needs.
type (
	dyeColorItem interface{ GetColor() blockutils.DyeColor }
	armorItem    interface {
		GetArmorSlot() int
		GetCustomColor() (color.Color, bool)
		SetCustomColor(c color.Color)
		ClearCustomColor()
	}
	bannerItem interface {
		GetPatterns() []blockutils.BannerPatternLayer
		SetPatterns(patterns []blockutils.BannerPatternLayer)
	}
	blockItem      interface{ GetBlock() Behavior }
	namedTagHolder interface {
		GetNamedTag() *nbt.CompoundTag
		SetNamedTag(tag *nbt.CompoundTag)
	}
)

// dyeColorOf is the dye colour match of WaterCauldron::onInteract (ok is false for other items).
func dyeColorOf(it Item) (blockutils.DyeColor, bool) {
	switch it.GetTypeId() {
	case itemTypeIDsLapisLazuli:
		return blockutils.DyeColorBlue, true
	case itemTypeIDsInkSac:
		return blockutils.DyeColorBlack, true
	case itemTypeIDsCocoaBeans:
		return blockutils.DyeColorBrown, true
	case itemTypeIDsBoneMeal:
		return blockutils.DyeColorWhite, true
	case itemTypeIDsDye:
		if dye, ok := it.(dyeColorItem); ok {
			return dye.GetColor(), true
		}
	}
	return 0, false
}

// OnInteract is a port of WaterCauldron::onInteract.
func (w *WaterCauldron) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	world, err := w.position.GetWorld()
	if err != nil {
		return true
	}
	center := w.position.Add(0.5, 0.5, 0.5)

	dyeColor, isDye := dyeColorOf(item)
	newColor := dyeColor.GetRgbValue()
	if isDye && (w.customWaterColor == nil || newColor.ToRGBA() != w.customWaterColor.ToRGBA()) {
		mixed := newColor
		if w.customWaterColor != nil {
			mixed = color.Mix(*w.customWaterColor, newColor)
		}
		_ = world.SetBlock(w.position, w.SetCustomWaterColor(&mixed))
		world.AddSound(center, sound.CauldronAddDyeSound{})

		item.Pop()
	} else if potion, ok := item.(waterPotionChecker); ok {
		if potion.IsWaterPotion() {
			w.SetCustomWaterColor(nil).addFillLevels(WaterCauldronWaterBottleFillAmount, item, vanillaItem("glass_bottle"), returnedItems)
		} else {
			w.mix(item, vanillaItem("glass_bottle"), returnedItems)
		}
	} else if armor, ok := item.(armorItem); ok {
		if w.customWaterColor != nil {
			customColor, hasCustomColor := armor.GetCustomColor()
			if isDyeableArmor(item.GetTypeId()) && (!hasCustomColor || customColor.ToRGBA() != w.customWaterColor.ToRGBA()) {
				armor.SetCustomColor(*w.customWaterColor)
				_ = world.SetBlock(w.position, w.withFillLevel(w.FillLevel-WaterCauldronDyeArmorUseAmount))
				world.AddSound(center, sound.CauldronDyeItemSound{})
			}
		} else if _, hasCustomColor := armor.GetCustomColor(); hasCustomColor {
			armor.ClearCustomColor()
			_ = world.SetBlock(w.position, w.withFillLevel(w.FillLevel-WaterCauldronCleanArmorUseAmount))
			world.AddSound(center, sound.CauldronCleanItemSound{})
		}
	} else if banner, ok := item.(bannerItem); ok {
		patterns := banner.GetPatterns()
		if len(patterns) > 0 && w.customWaterColor == nil {
			banner.SetPatterns(patterns[:len(patterns)-1])

			_ = world.SetBlock(w.position, w.withFillLevel(w.FillLevel-WaterCauldronCleanBannerUseAmount))
			world.AddSound(center, sound.CauldronCleanItemSound{})
		}
	} else if b, ok := item.(blockItem); ok && b.GetBlock().GetTypeId() == DYED_SHULKER_BOX { //ItemTypeIds::toBlockTypeId($item->getTypeId())
		if w.customWaterColor == nil {
			newItem, err := VanillaBlock("shulker_box").(interface{ AsItem() (Item, error) }).AsItem()
			if err == nil {
				if tagged, ok := item.(namedTagHolder); ok {
					newItem.(namedTagHolder).SetNamedTag(tagged.GetNamedTag())
				}

				item.Pop()
				appendItem(returnedItems, newItem)

				_ = world.SetBlock(w.position, w.withFillLevel(w.FillLevel-WaterCauldronCleanShulkerBoxUseAmount))
				world.AddSound(center, sound.CauldronCleanItemSound{})
			}
		}
	} else {
		switch item.GetTypeId() {
		case itemTypeIDsWaterBucket:
			w.SetCustomWaterColor(nil).addFillLevels(FillableCauldronMaxFillLevel, item, vanillaItem("bucket"), returnedItems)
		case itemTypeIDsBucket:
			w.removeFillLevels(FillableCauldronMaxFillLevel, item, vanillaItem("water_bucket"), returnedItems)
		case itemTypeIDsGlassBottle:
			// VanillaItems::POTION()->setType(PotionType::WATER): water is a new potion's type.
			w.removeFillLevels(WaterCauldronWaterBottleFillAmount, item, vanillaItem("potion"), returnedItems)
		case itemTypeIDsLavaBucket, itemTypeIDsPowderSnowBucket:
			w.mix(item, vanillaItem("bucket"), returnedItems)
		}
	}
	return true
}

// isDyeableArmor is WaterCauldron::onInteract's leather armour check.
func isDyeableArmor(typeID int) bool {
	//TODO: a DyeableArmor class would probably be a better idea, since not all types of armor are dyeable
	switch typeID {
	case itemTypeIDsLeatherCap, itemTypeIDsLeatherTunic, itemTypeIDsLeatherPants, itemTypeIDsLeatherBoots:
		return true
	}
	return false
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
