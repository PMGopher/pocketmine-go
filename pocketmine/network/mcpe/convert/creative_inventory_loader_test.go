package convert

import (
	"fmt"
	"testing"

	"pocketmine-go/pocketmine/data/bedrock"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

// Items that place a block but aren't block items (doors, signs, beds, cake, seeds, ...) come with
// a block runtime ID in the captured creative content; they must still be loaded.
func TestCreativeInventoryHasNonBlockItemsThatPlaceBlocks(t *testing.T) {
	inv := inventory.NewCreativeInventory()
	loadCreativeInventory(inv)
	for _, name := range []string{"wheat_seeds", "oak_sign"} {
		if inv.GetItemIndex(item.VanillaItem(name)) < 0 {
			t.Errorf("item %s is missing from the creative inventory", name)
		}
	}
	for _, name := range []string{"oak_door", "cake", "brewing_stand", "campfire", "flower_pot", "cauldron", "bed"} {
		it, err := block.VanillaBlock(name).(interface{ AsItem() (block.Item, error) }).AsItem()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if inv.GetItemIndex(it.(item.Item)) < 0 {
			t.Errorf("block item %s is missing from the creative inventory", name)
		}
	}
	if entries, _ := inv.GetAllEntries(); len(entries) < 1420 {
		t.Errorf("creative inventory has %d entries, want at least 1420", len(entries))
	}
}

// A creative shulker box must not carry a "facing" in its NBT: the shulker box tile copies the
// item's whole NBT when placed (ShulkerBox::copyDataFromItem), so Dragonfly's creative entry
// (facing:0) made every shulker box taken from the creative inventory face down.
func TestCreativeShulkerBoxHasNoFacing(t *testing.T) {
	inv := inventory.NewCreativeInventory()
	loadCreativeInventory(inv)
	it, err := block.VanillaBlock("shulker_box").(interface{ AsItem() (block.Item, error) }).AsItem()
	if err != nil {
		t.Fatal(err)
	}
	idx := inv.GetItemIndex(it.(item.Item))
	if idx < 0 {
		t.Fatal("no shulker box in the creative inventory")
	}
	if tag := inv.GetItem(idx).GetNamedTag(); tag != nil {
		if _, ok := tag.GetTag("facing"); ok {
			t.Errorf("creative shulker box NBT has a facing: %v", tag)
		}
	}
}

// Every creative entry must be a different item as the client sees it (no duplicates: beds of every
// colour used to all load as the white bed, enchanted books with enchantments PocketMine-MP lacks as
// blank books), and must round-trip to its own item ID and meta (block items with block runtime ID
// 0 used to load as cyan terracotta).
func TestCreativeInventoryHasNoDuplicatesOrWrongItems(t *testing.T) {
	inv := inventory.NewCreativeInventory()
	loadCreativeInventory(inv)
	entries, _ := inv.GetAllEntries()
	seen := map[string]bool{}
	for _, e := range entries {
		st := CoreItemStackToNet(e.GetItem())
		key := fmt.Sprintf("%d/%d/%d/%v", st.NetworkID, st.MetadataValue, st.BlockRuntimeID, st.NBTData)
		if seen[key] {
			t.Errorf("duplicate creative entry %s", e.GetItem().GetName())
		}
		seen[key] = true
	}
	for _, e := range bedrock.CreativeContent().Items {
		it, ok := loadCreativeItem(e.Item)
		if !ok {
			continue
		}
		back := CoreItemStackToNet(it)
		if back.NetworkID != e.Item.NetworkID || back.MetadataValue != e.Item.MetadataValue {
			name, _ := bedrock.ItemNameForRuntimeID(e.Item.NetworkID)
			t.Errorf("%s:%d loads as %s", name, e.Item.MetadataValue, it.GetName())
		}
	}
}
