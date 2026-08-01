package item

import (
	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/nbt"
)

const tagCustomColor = "customColor"

// Armor is a port of pocketmine\item\Armor. Unlike Tool's subclasses, Armor itself is directly
// instantiable in the PHP original (not abstract) - different armor pieces are just Armor
// instances constructed with different ArmorTypeInfo, so this is both the base and the leaf here.
//
// getEnchantmentProtectionFactor isn't ported: it needs ProtectionEnchantment (item/enchantment
// package, not ported), and enchantments are always empty here anyway (see
// ItemBase.deserializeCompoundTag's doc comment). getUnbreakingDamageReduction is skipped for the
// same reason as Durable's (see its doc comment) - Armor's version has an extra 40%-of-the-time
// gate on top, but since the enchantment level is always 0 the whole method is moot either way.
// OnClickAir (equipping the armor) needs a real Player/ArmorInventory - see the Item interface's
// doc comment on Player/Entity-interaction methods.
type Armor struct {
	Durable

	ArmorInfo ArmorTypeInfo

	customColor    color.Color
	hasCustomColor bool
}

func NewArmor(identifier ItemIdentifier, name string, info ArmorTypeInfo) *Armor {
	a := &Armor{ArmorInfo: info}
	a.Init(a, identifier, name)
	return a
}

func (a *Armor) Clone() Item {
	c := *a
	c.rebind(&c)
	return &c
}

func (a *Armor) GetMaxDurability() int { return a.ArmorInfo.GetMaxDurability() }

func (a *Armor) GetDefensePoints() int { return a.ArmorInfo.GetDefensePoints() }

// GetArmorSlot mirrors Armor::getArmorSlot - the index into pocketmine\inventory\ArmorInventory,
// which isn't ported, so this is just the raw slot number for now.
func (a *Armor) GetArmorSlot() int { return a.ArmorInfo.GetArmorSlot() }

func (a *Armor) GetMaxStackSize() int { return 1 }

func (a *Armor) IsFireProof() bool { return a.ArmorInfo.IsFireProof() }

func (a *Armor) GetMaterial() ArmorMaterial { return a.ArmorInfo.GetMaterial() }

func (a *Armor) GetEnchantability() int { return a.ArmorInfo.GetMaterial().GetEnchantability() }

// GetCustomColor is a port of Armor::getCustomColor.
func (a *Armor) GetCustomColor() (color.Color, bool) { return a.customColor, a.hasCustomColor }

func (a *Armor) SetCustomColor(c color.Color) { a.customColor = c; a.hasCustomColor = true }

func (a *Armor) ClearCustomColor() { a.customColor = color.Color{}; a.hasCustomColor = false }

// deserializeCompoundTag/serializeCompoundTag extend Durable's own pair, the same self-dispatch
// participation described on Durable's - the ARGB round trip skips PHP's Binary::signInt/
// unsignInt: Go's int32 conversions already preserve the exact bit pattern (same reasoning as
// tile.Sign's colour round trip).
func (a *Armor) deserializeCompoundTag(tag *nbt.CompoundTag) {
	a.Durable.deserializeCompoundTag(tag)
	if colorTag, err := tag.GetInt(tagCustomColor); err == nil {
		a.customColor = color.FromARGB(int32(colorTag))
		a.hasCustomColor = true
	} else {
		a.hasCustomColor = false
	}
}

func (a *Armor) serializeCompoundTag(tag *nbt.CompoundTag) {
	a.Durable.serializeCompoundTag(tag)
	if a.hasCustomColor {
		tag.SetInt(tagCustomColor, nbt.IntTag(a.customColor.ToARGB()))
	} else {
		tag.RemoveTag(tagCustomColor)
	}
}
