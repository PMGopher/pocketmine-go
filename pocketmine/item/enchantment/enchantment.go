// Package enchantment is a port of the parts of pocketmine\item\enchantment that items and entities
// need at runtime: Enchantment and its Protection/MeleeWeapon subtypes, EnchantmentInstance,
// Rarity, ItemFlags, IncompatibleEnchantmentRegistry, VanillaEnchantments and
// StringToEnchantmentParser.
//
// Not ported yet: the enchanting-table machinery (EnchantingHelper beyond GenerateSeed,
// EnchantingOption, AvailableEnchantmentRegistry, ItemEnchantmentTags/TagRegistry) - it needs the
// unported inventory transaction/enchanting-table flow to be useful.
//
// This package sits below pocketmine/item (items store EnchantmentInstances) and pocketmine/entity,
// so the entities MeleeWeaponEnchantment acts on are the local Entity interface below.
package enchantment

import (
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// Entity is the surface MeleeWeaponEnchantment implementations need from pocketmine\entity\Entity.
type Entity interface {
	entityevent.Entity
	SetOnFire(seconds int)
}

// knockBackable is the optional surface KnockbackEnchantment needs to recognise a Living (PHP's
// `$victim instanceof Living`). *entity.Living satisfies it.
type knockBackable interface {
	KnockBack(x, z, force, verticalLimit float64)
}

// Enchantment is a port of pocketmine\item\enchantment\Enchantment's public surface. Enchantments
// are compared by identity (PHP keys item enchantments by spl_object_id), so every Enchantment is a
// pointer and the VanillaEnchantments getters return the same instance each time.
type Enchantment interface {
	// GetName returns a translation key (*lang.Translatable) or a plain string name.
	GetName() any
	GetRarity() int
	GetPrimaryItemFlags() int
	GetSecondaryItemFlags() int
	HasPrimaryItemType(flag int) bool
	HasSecondaryItemType(flag int) bool
	GetMaxLevel() int
	IsCompatibleWith(other Enchantment) bool
	GetMinEnchantingPower(level int) int
	GetMaxEnchantingPower(level int) int
}

// EnchantmentBase is a port of pocketmine\item\enchantment\Enchantment's state and methods.
type EnchantmentBase struct {
	name                 any
	rarity               int
	primaryItemFlags     int
	secondaryItemFlags   int
	maxLevel             int
	minEnchantingPower   func(level int) int
	enchantingPowerRange int
}

// NewEnchantment is a port of Enchantment::__construct. A nil minEnchantingPower means PHP's default
// (`fn(int $level) => 1`); PHP's default enchantingPowerRange is 50.
func NewEnchantment(name any, rarity, primaryItemFlags, secondaryItemFlags, maxLevel int, minEnchantingPower func(level int) int, enchantingPowerRange int) *EnchantmentBase {
	e := &EnchantmentBase{}
	e.initBase(name, rarity, primaryItemFlags, secondaryItemFlags, maxLevel, minEnchantingPower, enchantingPowerRange)
	return e
}

func (e *EnchantmentBase) initBase(name any, rarity, primaryItemFlags, secondaryItemFlags, maxLevel int, minEnchantingPower func(level int) int, enchantingPowerRange int) {
	if minEnchantingPower == nil {
		minEnchantingPower = func(int) int { return 1 }
	}
	e.name = name
	e.rarity = rarity
	e.primaryItemFlags = primaryItemFlags
	e.secondaryItemFlags = secondaryItemFlags
	e.maxLevel = maxLevel
	e.minEnchantingPower = minEnchantingPower
	e.enchantingPowerRange = enchantingPowerRange
}

// GetName returns a translation key for this enchantment's name.
func (e *EnchantmentBase) GetName() any { return e.name }

// GetRarity returns an int constant indicating how rare this enchantment type is.
func (e *EnchantmentBase) GetRarity() int { return e.rarity }

// GetPrimaryItemFlags returns a bitset indicating what item types can have this item applied from
// an enchanting table.
//
// Deprecated: kept for compatibility, like PHP.
func (e *EnchantmentBase) GetPrimaryItemFlags() int { return e.primaryItemFlags }

// GetSecondaryItemFlags returns a bitset indicating what item types cannot have this item applied
// from an enchanting table, but can from an anvil.
//
// Deprecated: kept for compatibility, like PHP.
func (e *EnchantmentBase) GetSecondaryItemFlags() int { return e.secondaryItemFlags }

// HasPrimaryItemType returns whether this enchantment can apply to the item type from an
// enchanting table.
//
// Deprecated: kept for compatibility, like PHP.
func (e *EnchantmentBase) HasPrimaryItemType(flag int) bool { return e.primaryItemFlags&flag != 0 }

// HasSecondaryItemType returns whether this enchantment can apply to the item type from an anvil,
// if it is not a primary item.
//
// Deprecated: kept for compatibility, like PHP.
func (e *EnchantmentBase) HasSecondaryItemType(flag int) bool {
	return e.secondaryItemFlags&flag != 0
}

// GetMaxLevel returns the maximum level of this enchantment that can be found on an enchantment
// table.
func (e *EnchantmentBase) GetMaxLevel() int { return e.maxLevel }

// GetMinEnchantingPower returns the minimum enchanting power value required for the particular
// level of the enchantment to be available in an enchanting table.
func (e *EnchantmentBase) GetMinEnchantingPower(level int) int { return e.minEnchantingPower(level) }

// GetMaxEnchantingPower returns the maximum enchanting power value allowed for the particular level
// of the enchantment to be available in an enchanting table.
func (e *EnchantmentBase) GetMaxEnchantingPower(level int) int {
	return e.GetMinEnchantingPower(level) + e.enchantingPowerRange
}

// IsCompatibleWith returns whether this enchantment can be applied to the item along with the given
// enchantment. The receiver is the base struct, so this compares through the concrete Enchantment
// registered for it - see compatibilityKey.
func (e *EnchantmentBase) IsCompatibleWith(other Enchantment) bool {
	return GetIncompatibleEnchantmentRegistry().areCompatibleKeys(compatibilityKeyOf(e), compatibilityKeyOf(other))
}

// base exposes the embedded EnchantmentBase so identity-keyed registries can key by it regardless of
// which concrete wrapper type is passed.
func (e *EnchantmentBase) base() *EnchantmentBase { return e }

type hasBase interface{ base() *EnchantmentBase }

// compatibilityKeyOf maps any Enchantment (a bare *EnchantmentBase or a concrete subtype embedding
// one) to the single identity the incompatibility registry is keyed by.
func compatibilityKeyOf(e any) *EnchantmentBase {
	if b, ok := e.(hasBase); ok {
		return b.base()
	}
	return nil
}
