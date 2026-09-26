package blockinventory

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/math"
)

func TestContainerInventorySizes(t *testing.T) {
	holder := block.NewPosition(0, 0, 0, &fakeWorld{})
	cases := []struct {
		name string
		inv  inventory.Inventory
		want int
	}{
		{"Anvil", NewAnvilInventory(holder), 2},
		{"Barrel", NewBarrelInventory(holder), 27},
		{"BrewingStand", NewBrewingStandInventory(holder, 5), 5},
		{"Furnace", NewFurnaceInventory(holder, tile.FurnaceTypeSmoker), 3},
		{"Hopper", NewHopperInventory(holder, 5), 5},
		{"ShulkerBox", NewShulkerBoxInventory(holder), 27},
		{"EnderChest", NewEnderChestInventory(holder, inventory.NewPlayerEnderInventory(nil, 27)), 27},
	}
	for _, c := range cases {
		if got := c.inv.GetSize(); got != c.want {
			t.Errorf("%s: GetSize() = %d, want %d", c.name, got, c.want)
		}
		if _, ok := c.inv.(BlockInventory); !ok {
			t.Errorf("%s is not a BlockInventory", c.name)
		}
	}
}

func TestShulkerBoxInventoryRejectsShulkerBoxes(t *testing.T) {
	inv := NewShulkerBoxInventory(block.NewPosition(0, 0, 0, &fakeWorld{}))
	shulker := item.NewItemBlock(item.NewItemIdentifier(-block.SHULKER_BOX), block.VanillaDirt())
	if inv.CanAddItem(shulker) {
		t.Error("a shulker box item can be added to a shulker box inventory")
	}
	if !inv.CanAddItem(item.VanillaIronIngot()) {
		t.Error("an iron ingot can't be added to a shulker box inventory")
	}
}

// Listeners (and the network inventory manager, keyed by inventory) must be told about the
// block inventory itself, not the SimpleInventory it embeds.
func TestBlockInventoryListenersSeeOuterInventory(t *testing.T) {
	holder := block.NewPosition(0, 0, 0, &fakeWorld{})
	for _, inv := range []inventory.Inventory{
		NewChestInventory(holder), NewFurnaceInventory(holder, tile.FurnaceTypeFurnace), NewCraftingTableInventory(holder),
		NewEnderChestInventory(holder, inventory.NewPlayerEnderInventory(nil, 27)),
	} {
		var seen inventory.Inventory
		inv.GetListeners().Add(inventory.NewCallbackInventoryListener(func(changed inventory.Inventory, _ int, _ item.Item) { seen = changed }, nil))
		inv.SetItem(0, item.VanillaIronIngot())
		if seen != inv {
			t.Errorf("%T: listener got %T", inv, seen)
		}
	}
}

func TestTemporaryBlockInventories(t *testing.T) {
	holder := block.NewPosition(0, 0, 0, &fakeWorld{})
	for _, inv := range []inventory.Inventory{NewAnvilInventory(holder), NewEnchantInventory(holder), NewCraftingTableInventory(holder)} {
		if _, ok := inv.(inventory.TemporaryInventory); !ok {
			t.Errorf("%T should be a TemporaryInventory (its items are returned when it closes)", inv)
		}
	}
	if _, ok := inventory.Inventory(NewChestInventory(holder)).(inventory.TemporaryInventory); ok {
		t.Error("ChestInventory must not be a TemporaryInventory")
	}
}

type fakeEnchantingViewer struct {
	options []*enchantment.EnchantingOption
}

func (v *fakeEnchantingViewer) GetID() int                               { return 1 }
func (v *fakeEnchantingViewer) GetPosition() math.Vector3                { return math.Vector3{} }
func (v *fakeEnchantingViewer) IsClosed() bool                           { return false }
func (v *fakeEnchantingViewer) GetName() string                          { return "viewer" }
func (v *fakeEnchantingViewer) GetDisplayName() string                   { return "viewer" }
func (v *fakeEnchantingViewer) GetEnchantmentSeed() int                  { return 12345 }
func (v *fakeEnchantingViewer) GetInvManager() inventory.InventorySyncer { return v }
func (v *fakeEnchantingViewer) OnSlotChange(inventory.Inventory, int)    {}
func (v *fakeEnchantingViewer) SyncContents(inventory.Inventory)         {}
func (v *fakeEnchantingViewer) SyncEnchantingTableOptions(options []*enchantment.EnchantingOption) {
	v.options = options
}

func TestEnchantInventoryGeneratesOptionsForViewers(t *testing.T) {
	inv := NewEnchantInventory(block.NewPosition(0, 64, 0, &fakeWorld{}))
	viewer := &fakeEnchantingViewer{}
	inv.OnOpen(viewer)

	inv.SetItem(EnchantSlotInput, item.VanillaBook())
	if len(viewer.options) != 3 {
		t.Fatalf("the viewer got %d options, want 3", len(viewer.options))
	}
	if inv.GetOption(0) != viewer.options[0] || inv.GetOption(3) != nil {
		t.Error("GetOption doesn't return the options sent to the viewer")
	}
	output := inv.GetOutput(0)
	if output == nil || output.GetTypeId() != item.ENCHANTED_BOOK || len(output.GetEnchantments()) != len(viewer.options[0].GetEnchantments()) {
		t.Errorf("a book should enchant into an enchanted book with the option's enchantments, got %v", output)
	}

	viewer.options = nil
	inv.SetItem(EnchantSlotLapis, item.VanillaIronIngot())
	if viewer.options != nil {
		t.Error("changing the lapis slot regenerated the options")
	}
}
