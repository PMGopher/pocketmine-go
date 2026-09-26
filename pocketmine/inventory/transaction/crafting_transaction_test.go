package transaction

import (
	"testing"

	"pocketmine-go/pocketmine/crafting"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

func withCount(it item.Item, n int) item.Item {
	it.SetCount(n)
	return it
}

// flintAndSteelManager has one recipe: iron ingot + egg -> flint and steel.
func flintAndSteelManager(t *testing.T) (*crafting.CraftingManager, crafting.CraftingRecipe) {
	t.Helper()
	m := crafting.NewCraftingManager()
	r, err := crafting.NewShapelessRecipe([]crafting.RecipeIngredient{
		crafting.NewMetaWildcardRecipeIngredient("minecraft:iron_ingot"),
		crafting.NewTagWildcardRecipeIngredient("minecraft:egg"),
	}, []item.Item{item.VanillaFlintAndSteel()}, crafting.ShapelessRecipeTypeCrafting)
	if err != nil {
		t.Fatal(err)
	}
	m.RegisterShapelessRecipe(r)
	return m, r
}

func TestCraftingTransactionMatchesRecipeFromOutputs(t *testing.T) {
	m, r := flintAndSteelManager(t)
	inv := inventory.NewSimpleInventory(9)
	inv.SetItem(0, withCount(item.VanillaIronIngot(), 2))
	inv.SetItem(1, withCount(item.VanillaEgg(), 2))

	actions := []InventoryAction{
		NewSlotChangeAction(inv, 0, withCount(item.VanillaIronIngot(), 2), withCount(item.VanillaIronIngot(), 1)),
		NewSlotChangeAction(inv, 1, withCount(item.VanillaEgg(), 2), withCount(item.VanillaEgg(), 1)),
		NewSlotChangeAction(inv, 2, inventory.Air(), item.VanillaFlintAndSteel()),
	}
	tx, err := NewCraftingTransaction(&fakePlayer{finite: true}, m, actions, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if tx.GetRecipe() != r {
		t.Error("the recipe wasn't matched from the outputs")
	}
	if inv.GetItem(2).GetTypeId() != item.VanillaFlintAndSteel().GetTypeId() || inv.GetItem(0).GetCount() != 1 {
		t.Error("the transaction wasn't applied")
	}
}

func TestCraftingTransactionRejectsMissingIngredient(t *testing.T) {
	m, r := flintAndSteelManager(t)
	inv := inventory.NewSimpleInventory(9)
	inv.SetItem(0, item.VanillaIronIngot())
	actions := []InventoryAction{
		NewSlotChangeAction(inv, 0, item.VanillaIronIngot(), inventory.Air()),
		NewSlotChangeAction(inv, 2, inventory.Air(), item.VanillaFlintAndSteel()),
	}
	tx, err := NewCraftingTransaction(&fakePlayer{finite: true}, m, actions, r, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Execute(); err == nil {
		t.Fatal("crafting without the egg should fail validation")
	}
	if !inv.GetItem(2).IsNull() {
		t.Error("a failed transaction changed the inventory")
	}
}

func TestMatchIngredientsRepetitions(t *testing.T) {
	ingredients := []crafting.RecipeIngredient{crafting.NewMetaWildcardRecipeIngredient("minecraft:iron_ingot")}
	if err := MatchIngredients([]item.Item{withCount(item.VanillaIronIngot(), 3)}, ingredients, 3); err != nil {
		t.Errorf("3 ingots for 3 repetitions: %v", err)
	}
	if err := MatchIngredients([]item.Item{withCount(item.VanillaIronIngot(), 2)}, ingredients, 3); err == nil {
		t.Error("2 ingots for 3 repetitions should fail")
	}
	if err := MatchIngredients([]item.Item{withCount(item.VanillaIronIngot(), 4)}, ingredients, 3); err == nil {
		t.Error("a leftover ingot should fail")
	}
}
