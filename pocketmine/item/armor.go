package item

import (
	"math/rand/v2"

	"pocketmine-go/pocketmine/color"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/utils"
)

const tagCustomColor = "customColor"

// Armor is a port of pocketmine\item\Armor. Unlike Tool's subclasses, Armor itself is directly
// instantiable in the PHP original (not abstract) - different armor pieces are just Armor
// instances constructed with different ArmorTypeInfo, so this is both the base and the leaf here.

type Armor struct {
	Durable

	ArmorInfo ArmorTypeInfo

	customColor    color.Color
	hasCustomColor bool
}

func NewArmor(identifier ItemIdentifier, name string, info ArmorTypeInfo, enchantmentTags ...string) *Armor {
	a := &Armor{ArmorInfo: info}
	a.Init(a, identifier, name)
	a.enchantmentTags = enchantmentTags
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

// GetEnchantmentProtectionFactor is a port of Armor::getEnchantmentProtectionFactor: the total
// enchantment protection factor this armour piece offers from all applicable protection
// enchantments on the item.
func (a *Armor) GetEnchantmentProtectionFactor(event entityevent.DamageSource) int {
	epf := 0
	for _, instance := range a.GetEnchantments() {
		if t, ok := instance.GetType().(*enchantment.ProtectionEnchantment); ok && t.IsApplicable(event) {
			epf += t.GetProtectionFactor(instance.GetLevel())
		}
	}
	return epf
}

// getUnbreakingDamageReduction is a port of Armor::getUnbreakingDamageReduction: unbreaking only
// applies to armor 40% of the time at best.
func (a *Armor) getUnbreakingDamageReduction(amount int) int {
	unbreakingLevel := a.GetEnchantmentLevel(enchantment.VanillaUnbreaking())
	if unbreakingLevel <= 0 {
		return 0
	}
	negated := 0
	chance := 1 / float64(unbreakingLevel+1)
	for i := 0; i < amount; i++ {
		if rand.IntN(100)+1 > 60 && utils.GetRandomFloat() > chance {
			negated++
		}
	}
	return negated
}

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
