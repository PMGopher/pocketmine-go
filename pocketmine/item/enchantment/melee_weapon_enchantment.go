package enchantment

import entityevent "pocketmine-go/pocketmine/event/entity"

// MeleeWeaponEnchantment is a port of the abstract pocketmine\item\enchantment\MeleeWeaponEnchantment
// - classes extending this represent enchantments which apply to melee weapons (Sharpness,
// Knockback, Fire Aspect).
type MeleeWeaponEnchantment interface {
	Enchantment

	// IsApplicableTo returns whether this melee enchantment has an effect on the target entity. For
	// example, it might only apply to undead mobs.
	IsApplicableTo(victim Entity) bool

	// GetDamageBonus returns the amount of additional damage caused by this enchantment to
	// applicable targets.
	GetDamageBonus(enchantmentLevel int) float64

	// OnPostAttack is called after damaging the entity to apply any post damage effects to the
	// target.
	OnPostAttack(attacker, victim Entity, enchantmentLevel int)
}

// SharpnessEnchantment is a port of pocketmine\item\enchantment\SharpnessEnchantment.
type SharpnessEnchantment struct{ EnchantmentBase }

func NewSharpnessEnchantment(name any, rarity, primaryItemFlags, secondaryItemFlags, maxLevel int, minEnchantingPower func(level int) int, enchantingPowerRange int) *SharpnessEnchantment {
	e := &SharpnessEnchantment{}
	e.initBase(name, rarity, primaryItemFlags, secondaryItemFlags, maxLevel, minEnchantingPower, enchantingPowerRange)
	return e
}

func (e *SharpnessEnchantment) IsApplicableTo(victim Entity) bool { return true }

func (e *SharpnessEnchantment) GetDamageBonus(enchantmentLevel int) float64 {
	return 0.5 * float64(enchantmentLevel+1)
}

func (e *SharpnessEnchantment) OnPostAttack(attacker, victim Entity, enchantmentLevel int) {}

// KnockbackEnchantment is a port of pocketmine\item\enchantment\KnockbackEnchantment.
type KnockbackEnchantment struct{ EnchantmentBase }

func NewKnockbackEnchantment(name any, rarity, primaryItemFlags, secondaryItemFlags, maxLevel int, minEnchantingPower func(level int) int, enchantingPowerRange int) *KnockbackEnchantment {
	e := &KnockbackEnchantment{}
	e.initBase(name, rarity, primaryItemFlags, secondaryItemFlags, maxLevel, minEnchantingPower, enchantingPowerRange)
	return e
}

func (e *KnockbackEnchantment) IsApplicableTo(victim Entity) bool {
	_, ok := victim.(knockBackable)
	return ok
}

func (e *KnockbackEnchantment) GetDamageBonus(enchantmentLevel int) float64 { return 0 }

// OnPostAttack is a port of KnockbackEnchantment::onPostAttack (the vertical limit is Living's
// default, as PHP omits the argument).
func (e *KnockbackEnchantment) OnPostAttack(attacker, victim Entity, enchantmentLevel int) {
	if living, ok := victim.(knockBackable); ok {
		diff := victim.GetPosition().SubtractVector(attacker.GetPosition())
		living.KnockBack(diff.X, diff.Z, float64(enchantmentLevel)*0.5, entityevent.DefaultKnockbackVerticalLimit)
	}
}

// FireAspectEnchantment is a port of pocketmine\item\enchantment\FireAspectEnchantment.
type FireAspectEnchantment struct{ EnchantmentBase }

func NewFireAspectEnchantment(name any, rarity, primaryItemFlags, secondaryItemFlags, maxLevel int, minEnchantingPower func(level int) int, enchantingPowerRange int) *FireAspectEnchantment {
	e := &FireAspectEnchantment{}
	e.initBase(name, rarity, primaryItemFlags, secondaryItemFlags, maxLevel, minEnchantingPower, enchantingPowerRange)
	return e
}

func (e *FireAspectEnchantment) IsApplicableTo(victim Entity) bool { return true }

func (e *FireAspectEnchantment) GetDamageBonus(enchantmentLevel int) float64 { return 0 }

func (e *FireAspectEnchantment) OnPostAttack(attacker, victim Entity, enchantmentLevel int) {
	victim.SetOnFire(enchantmentLevel * 4)
}
