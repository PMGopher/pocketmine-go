package convert

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/lang"
)

func init() {
	inventory.CreativeInventoryLoader = loadCreativeInventory
}

// loadCreativeInventory is CreativeInventory::__construct's loading of the vanilla creative items:
// the vendored 1.26.50 creative content (bedrock.CreativeContent, the same data as BedrockData's
// creative item lists), in its categories and groups. Like PHP, items that can't be deserialized
// are left out (so far only the item types ItemTranslator maps can be; see FromNetworkID).
func loadCreativeInventory(inv *inventory.CreativeInventory) {
	content := bedrock.CreativeContent()

	groups := make([]*inventory.CreativeGroup, len(content.Groups))
	categories := make([]inventory.CreativeCategory, len(content.Groups))
	for i, g := range content.Groups {
		categories[i] = coreCreativeCategory(g.Category)
		if g.Name == "" || g.Icon.NetworkID == 0 {
			continue
		}
		icon, err := NetItemStackToCore(g.Icon)
		if err != nil || icon.IsNull() {
			continue
		}
		groups[i] = inventory.NewCreativeGroup(lang.NewTranslatable(g.Name, nil), icon)
	}

	for _, entry := range content.Items {
		it, err := NetItemStackToCore(entry.Item)
		if err != nil || it.IsNull() {
			continue
		}
		category := inventory.CreativeCategoryItems
		var group *inventory.CreativeGroup
		if int(entry.GroupIndex) < len(groups) {
			category = categories[entry.GroupIndex]
			group = groups[entry.GroupIndex]
		}
		inv.Add(it, category, group)
	}
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
