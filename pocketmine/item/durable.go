package item

import (
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/utils"
)

const (
	tagUnbreakable = "Unbreakable"
	tagDamage      = "Damage"
)

// durableShaper lets concrete tool/armor types (FlintSteel, and eventually Axe/Sword/Armor/etc.)
// report their own maximum durability - Durable.ApplyDamage/IsBroken reach it via a type
// assertion on the promoted self field, the same narrow self-dispatch shape used throughout
// block/ (e.g. candleShaper) for PHP abstract-base-with-internal-self-calls situations.
type durableShaper interface {
	GetMaxDurability() int
}

// Durable is a port of pocketmine\item\Durable. Its ApplyDamage(int) bool method is what
// satisfies block.Durable (the forward-compatible marker declared in block/candle_component.go),
// so a real Durable-embedding item (like FlintSteel) now works with Candle's lighting logic
// instead of only a hypothetical one.
type Durable struct {
	ItemBase

	Damage      int
	unbreakable bool
}

func (d *Durable) IsUnbreakable() bool { return d.unbreakable }

func (d *Durable) SetUnbreakable(value bool) { d.unbreakable = value }

// ApplyDamage is a port of Durable::applyDamage.
func (d *Durable) ApplyDamage(amount int) bool {
	if d.IsUnbreakable() || d.IsBroken() {
		return false
	}
	amount -= d.self.(unbreakingReducer).getUnbreakingDamageReduction(amount)

	maxDurability := d.self.(durableShaper).GetMaxDurability()
	d.Damage = min(d.Damage+amount, maxDurability)
	if d.IsBroken() {
		d.onBroken()
	}
	return true
}

// unbreakingReducer is the self-dispatch hook letting Armor override getUnbreakingDamageReduction
// (PHP's protected method override).
type unbreakingReducer interface {
	getUnbreakingDamageReduction(amount int) int
}

// getUnbreakingDamageReduction is a port of Durable::getUnbreakingDamageReduction: each point of
// damage has a 1/(level+1) chance of applying with Unbreaking.
func (d *Durable) getUnbreakingDamageReduction(amount int) int {
	unbreakingLevel := d.self.GetEnchantmentLevel(enchantment.VanillaUnbreaking())
	if unbreakingLevel <= 0 {
		return 0
	}
	negated := 0
	chance := 1 / float64(unbreakingLevel+1)
	for i := 0; i < amount; i++ {
		if utils.GetRandomFloat() > chance {
			negated++
		}
	}
	return negated
}

func (d *Durable) GetDamage() int { return d.Damage }

// SetDamage panics if damage is out of range, mirroring the PHP original's
// InvalidArgumentException (a programmer error at the call site) - same convention as
// block.Bed.SetOccupied-adjacent SetX methods elsewhere in this port (e.g. AgeComponent.SetAge).
func (d *Durable) SetDamage(damage int) {
	maxDurability := d.self.(durableShaper).GetMaxDurability()
	if damage < 0 || damage > maxDurability {
		panic("Damage must be in range 0 - max durability")
	}
	d.Damage = damage
}

// IsBroken is a port of Durable::isBroken.
func (d *Durable) IsBroken() bool {
	return d.Damage >= d.self.(durableShaper).GetMaxDurability() || d.self.IsNull()
}

// onBroken is a port of Durable::onBroken.
func (d *Durable) onBroken() {
	d.self.Pop()
	d.SetDamage(0)
}

// deserializeCompoundTag/serializeCompoundTag are Durable's participation in the compoundTagCodec
// self-dispatch chain (see ItemBase.GetNamedTag's doc comment) - a concrete type embedding
// Durable without further NBT fields of its own can rely on these being reached automatically,
// the same way a concrete type embedding just ItemBase relies on ItemBase's own pair.
func (d *Durable) deserializeCompoundTag(tag *nbt.CompoundTag) {
	d.ItemBase.deserializeCompoundTag(tag)
	d.unbreakable = tag.GetByteOr(tagUnbreakable, 0) != 0

	damage := int(tag.GetIntOr(tagDamage, nbt.IntTag(d.Damage)))
	maxDurability := d.self.(durableShaper).GetMaxDurability()
	if damage != d.Damage && damage >= 0 && damage <= maxDurability {
		d.SetDamage(damage)
	}
}

func (d *Durable) serializeCompoundTag(tag *nbt.CompoundTag) {
	d.ItemBase.serializeCompoundTag(tag)
	if d.unbreakable {
		tag.SetByte(tagUnbreakable, 1)
	} else {
		tag.RemoveTag(tagUnbreakable)
	}
	if d.Damage != 0 {
		tag.SetInt(tagDamage, nbt.IntTag(d.Damage))
	} else {
		tag.RemoveTag(tagDamage)
	}
}
