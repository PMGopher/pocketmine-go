package item

import (
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/nbt"
)

// NBT keys for the "ench" list, a port of Item::TAG_ENCH/TAG_ENCH_ID/TAG_ENCH_LVL.
const (
	TagEnch    = "ench"
	tagEnchID  = "id"
	tagEnchLvl = "lvl"
)

// enchantments is a port of ItemEnchantmentHandlingTrait's state: enchantments keyed by type
// (identity, like PHP's spl_object_id), plus insertion order so GetEnchantments/NBT output keep
// PHP's array ordering.
type enchantments struct {
	byType map[enchantment.Enchantment]*enchantment.EnchantmentInstance
	order  []enchantment.Enchantment
}

func (e *enchantments) clone() enchantments {
	c := enchantments{order: append([]enchantment.Enchantment(nil), e.order...)}
	if e.byType != nil {
		c.byType = make(map[enchantment.Enchantment]*enchantment.EnchantmentInstance, len(e.byType))
		for k, v := range e.byType {
			c.byType[k] = v
		}
	}
	return c
}

// HasEnchantments is a port of ItemEnchantmentHandlingTrait::hasEnchantments.
func (b *ItemBase) HasEnchantments() bool { return len(b.enchantments.byType) > 0 }

// HasEnchantment is a port of ItemEnchantmentHandlingTrait::hasEnchantment. level -1 (PHP's
// default) matches any level.
func (b *ItemBase) HasEnchantment(e enchantment.Enchantment, level int) bool {
	instance, ok := b.enchantments.byType[e]
	return ok && (level == -1 || instance.GetLevel() == level)
}

// GetEnchantment is a port of ItemEnchantmentHandlingTrait::getEnchantment (nil if absent).
func (b *ItemBase) GetEnchantment(e enchantment.Enchantment) *enchantment.EnchantmentInstance {
	return b.enchantments.byType[e]
}

// RemoveEnchantment is a port of ItemEnchantmentHandlingTrait::removeEnchantment. level -1 (PHP's
// default) removes the enchantment whatever its level.
func (b *ItemBase) RemoveEnchantment(e enchantment.Enchantment, level int) {
	instance := b.GetEnchantment(e)
	if instance != nil && (level == -1 || instance.GetLevel() == level) {
		delete(b.enchantments.byType, e)
		for i, t := range b.enchantments.order {
			if t == e {
				b.enchantments.order = append(b.enchantments.order[:i:i], b.enchantments.order[i+1:]...)
				break
			}
		}
	}
}

// RemoveEnchantments is a port of ItemEnchantmentHandlingTrait::removeEnchantments.
func (b *ItemBase) RemoveEnchantments() { b.enchantments = enchantments{} }

// AddEnchantment is a port of ItemEnchantmentHandlingTrait::addEnchantment - replaces any existing
// enchantment of the same type.
func (b *ItemBase) AddEnchantment(instance *enchantment.EnchantmentInstance) {
	t := instance.GetType()
	if b.enchantments.byType == nil {
		b.enchantments.byType = map[enchantment.Enchantment]*enchantment.EnchantmentInstance{}
	}
	if _, exists := b.enchantments.byType[t]; !exists {
		b.enchantments.order = append(b.enchantments.order, t)
	}
	b.enchantments.byType[t] = instance
}

// GetEnchantments is a port of ItemEnchantmentHandlingTrait::getEnchantments, in insertion order.
func (b *ItemBase) GetEnchantments() []*enchantment.EnchantmentInstance {
	result := make([]*enchantment.EnchantmentInstance, 0, len(b.enchantments.order))
	for _, t := range b.enchantments.order {
		result = append(result, b.enchantments.byType[t])
	}
	return result
}

// GetEnchantmentLevel is a port of ItemEnchantmentHandlingTrait::getEnchantmentLevel: the level of
// the enchantment on this item, or 0 if the item doesn't have it.
func (b *ItemBase) GetEnchantmentLevel(e enchantment.Enchantment) int {
	if instance := b.GetEnchantment(e); instance != nil {
		return instance.GetLevel()
	}
	return 0
}

// deserializeEnchantments is the "ench" half of Item::deserializeCompoundTag.
func (b *ItemBase) deserializeEnchantments(tag *nbt.CompoundTag) {
	b.RemoveEnchantments()
	list, ok, _ := tag.GetListTag(TagEnch)
	if !ok {
		return
	}
	for _, t := range list.Values() {
		entry, ok := t.(*nbt.CompoundTag)
		if !ok {
			continue
		}
		magicNumber := int(entry.GetShortOr(tagEnchID, -1))
		level := int(entry.GetShortOr(tagEnchLvl, 0))
		if level <= 0 {
			continue
		}
		if enchType, ok := bedrock.EnchantmentIdMap().FromID(magicNumber); ok {
			b.AddEnchantment(enchantment.NewEnchantmentInstance(enchType, level))
		}
	}
}

// serializeEnchantments is the "ench" half of Item::serializeCompoundTag.
func (b *ItemBase) serializeEnchantments(tag *nbt.CompoundTag) {
	if len(b.enchantments.order) == 0 {
		tag.RemoveTag(TagEnch)
		return
	}
	idMap := bedrock.EnchantmentIdMap()
	values := make([]nbt.Tag, 0, len(b.enchantments.order))
	for _, instance := range b.GetEnchantments() {
		values = append(values, nbt.NewCompoundTag().
			SetShort(tagEnchID, nbt.ShortTag(idMap.ToID(instance.GetType()))).
			SetShort(tagEnchLvl, nbt.ShortTag(instance.GetLevel())))
	}
	list, err := nbt.NewListTag(values, nbt.TagCompound)
	if err != nil {
		panic(err)
	}
	tag.SetTag(TagEnch, list)
}
