package enchantment

import (
	stdmath "math"

	entityevent "pocketmine-go/pocketmine/event/entity"
)

// ProtectionEnchantment is a port of pocketmine\item\enchantment\ProtectionEnchantment.
type ProtectionEnchantment struct {
	EnchantmentBase

	typeModifier          float64
	applicableDamageTypes map[int]bool
}

// NewProtectionEnchantment is a port of ProtectionEnchantment::__construct. A nil
// applicableDamageTypes means the enchantment applies to every damage cause.
func NewProtectionEnchantment(name any, rarity, primaryItemFlags, secondaryItemFlags, maxLevel int, typeModifier float64, applicableDamageTypes []int, minEnchantingPower func(level int) int, enchantingPowerRange int) *ProtectionEnchantment {
	e := &ProtectionEnchantment{typeModifier: typeModifier}
	e.initBase(name, rarity, primaryItemFlags, secondaryItemFlags, maxLevel, minEnchantingPower, enchantingPowerRange)
	if applicableDamageTypes != nil {
		e.applicableDamageTypes = make(map[int]bool, len(applicableDamageTypes))
		for _, cause := range applicableDamageTypes {
			e.applicableDamageTypes[cause] = true
		}
	}
	return e
}

// GetTypeModifier returns the multiplier by which this enchantment type's EPF increases with each
// enchantment level.
func (e *ProtectionEnchantment) GetTypeModifier() float64 { return e.typeModifier }

// GetProtectionFactor returns the base EPF this enchantment type offers for the given enchantment
// level.
func (e *ProtectionEnchantment) GetProtectionFactor(level int) int {
	return int(stdmath.Floor(float64(6+level*level) * e.typeModifier / 3))
}

// IsApplicable returns whether this enchantment type offers protection from the specified damage
// source's cause.
func (e *ProtectionEnchantment) IsApplicable(event entityevent.DamageSource) bool {
	return e.applicableDamageTypes == nil || e.applicableDamageTypes[event.GetCause()]
}
