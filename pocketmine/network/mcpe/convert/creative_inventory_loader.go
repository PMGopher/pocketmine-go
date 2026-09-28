package convert

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	bedrockitem "pocketmine-go/pocketmine/data/bedrock/item"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/lang"
)

func init() {
	inventory.CreativeInventoryLoader = loadCreativeInventory
}

// loadCreativeInventory is CreativeInventory::__construct's loading of the vanilla creative items:
// the vendored creative content (bedrock.CreativeContent, captured from the vanilla server, the
// same source BedrockData's creative item lists are made from), in its categories and groups.
// Like PHP, items that can't be deserialized are left out. Three things differ from BedrockData's
// JSON, which is why this does more than PHP's loader:
//   - the packet gives some block items (e.g. poplar doors) block runtime ID 0, meaning "none", but
//     0 is also the first palette state (cyan terracotta); such entries use the block's own first
//     state instead, so a block PocketMine-MP doesn't have is left out instead of loading as
//     cyan terracotta;
//   - an item that only loads partially (an enchanted book whose enchantment PocketMine-MP doesn't
//     have loads as a blank book, and there are many of them) is left out instead of adding
//     duplicates;
//   - a group whose icon PocketMine-MP can't load (e.g. pottery sherds) gets its first loadable
//     item as icon, instead of scattering its items outside any group.
func loadCreativeInventory(inv *inventory.CreativeInventory) {
	content := bedrock.CreativeContent()

	type loaded struct {
		it    item.Item
		group uint32
	}
	var items []loaded
	firstItem := map[uint32]item.Item{}
	for _, entry := range content.Items {
		it, ok := loadCreativeItem(entry.Item)
		if !ok {
			continue
		}
		items = append(items, loaded{it, entry.GroupIndex})
		if _, seen := firstItem[entry.GroupIndex]; !seen {
			firstItem[entry.GroupIndex] = it
		}
	}

	groups := make([]*inventory.CreativeGroup, len(content.Groups))
	categories := make([]inventory.CreativeCategory, len(content.Groups))
	for i, g := range content.Groups {
		categories[i] = coreCreativeCategory(g.Category)
		if g.Name == "" {
			continue
		}
		icon, ok := loadCreativeItem(g.Icon)
		if !ok {
			if icon, ok = firstItem[uint32(i)]; !ok {
				continue // no item of this group loads
			}
		}
		groups[i] = inventory.NewCreativeGroup(lang.NewTranslatable(g.Name, nil), icon.Clone())
	}

	for _, l := range items {
		category := inventory.CreativeCategoryItems
		var group *inventory.CreativeGroup
		if int(l.group) < len(groups) {
			category = categories[l.group]
			group = groups[l.group]
		}
		inv.Add(l.it, category, group)
	}
}

// loadCreativeItem deserializes one creative entry (see loadCreativeInventory).
func loadCreativeItem(stack protocol.ItemStack) (item.Item, bool) {
	if stack.NetworkID == 0 {
		return nil, false
	}
	if stack.BlockRuntimeID == noBlockRuntimeID {
		if name, ok := bedrock.ItemNameForRuntimeID(stack.NetworkID); ok {
			if blockID, isBlockItem := bedrockitem.GetBlockItemIdMap().LookupBlockID(name); isBlockItem {
				if bedrock.BlockStates()[0].Name != blockID {
					ids := bedrock.RuntimeIDsForName(blockID)
					if len(ids) == 0 {
						return nil, false
					}
					stack.BlockRuntimeID = ids[0]
				}
			}
		}
	}
	it, err := NetItemStackToCore(stack)
	if err != nil || it.IsNull() {
		return nil, false
	}
	if ench, ok := stack.NBTData["ench"].([]any); ok && len(ench) > 0 {
		if enchanted, ok := it.(interface {
			GetEnchantments() []*enchantment.EnchantmentInstance
		}); !ok || len(enchanted.GetEnchantments()) < len(ench) {
			return nil, false
		}
	}
	return it, true
}

// coreCreativeCategory maps CreativeContentPacket::CATEGORY_* to CreativeCategory.
func coreCreativeCategory(category byte) inventory.CreativeCategory {
	switch category {
	case protocol.CreativeCategoryConstruction:
		return inventory.CreativeCategoryConstruction
	case protocol.CreativeCategoryNature:
		return inventory.CreativeCategoryNature
	case protocol.CreativeCategoryEquipment:
		return inventory.CreativeCategoryEquipment
	}
	return inventory.CreativeCategoryItems
}

// ProtocolCreativeCategory maps CreativeCategory to CreativeContentPacket::CATEGORY_*.
func ProtocolCreativeCategory(category inventory.CreativeCategory) byte {
	switch category {
	case inventory.CreativeCategoryConstruction:
		return protocol.CreativeCategoryConstruction
	case inventory.CreativeCategoryNature:
		return protocol.CreativeCategoryNature
	case inventory.CreativeCategoryEquipment:
		return protocol.CreativeCategoryEquipment
	}
	return protocol.CreativeCategoryItems
}
