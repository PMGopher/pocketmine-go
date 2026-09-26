package item

import "pocketmine-go/pocketmine/item/enchantment"

// EnchantItem is a port of EnchantingHelper::enchantItem (it lives here because it needs
// VanillaItems): a book becomes an enchanted book, any other item is cloned, and the
// enchantments are added to the result.
func EnchantItem(it Item, enchantments []*enchantment.EnchantmentInstance) Item {
	var resultItem Item
	if it.GetTypeId() == BOOK {
		resultItem = VanillaEnchantedBook()
	} else {
		resultItem = it.Clone()
	}

	for _, e := range enchantments {
		resultItem.AddEnchantment(e)
	}

	return resultItem
}
