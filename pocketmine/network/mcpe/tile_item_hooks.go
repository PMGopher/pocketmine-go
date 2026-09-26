package mcpe

import (
	blockinventory "pocketmine-go/pocketmine/block/inventory"
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network/mcpe/convert"
)

// init gives block/inventory and block/tile TypeConverter::getInstance()->getItemTranslator()->toNetworkNbt() for
// the campfire's spawn data (block/inventory can't import convert: crafting, which it imports,
// imports convert... and convert would need block/inventory).
func init() {
	var translator *convert.ItemTranslator
	blockinventory.ToNetworkNbtFunc = func(it item.Item) *nbt.CompoundTag {
		if translator == nil {
			translator = convert.NewItemTranslator()
		}
		tag, err := translator.ToNetworkNbt(it)
		if err != nil {
			return nil
		}
		return tag
	}
	tile.ItemToNetworkNbtFunc = func(it tile.Item) *nbt.CompoundTag {
		real, ok := it.(item.Item)
		if !ok {
			return nil
		}
		return blockinventory.ToNetworkNbtFunc(real)
	}
}
