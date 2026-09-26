package crafting

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/item"
)

func loadVanillaManager(t *testing.T) *CraftingManager {
	t.Helper()
	m, err := MakeCraftingManager(bedrock.Recipes, bedrock.RecipesDir)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMakeCraftingManagerLoadsKnownRecipes(t *testing.T) {
	m := loadVanillaManager(t)
	if len(m.GetCraftingRecipeIndex()) == 0 {
		t.Fatal("no crafting recipes loaded")
	}
	if len(m.GetFurnaceRecipeManager(tile.FurnaceTypeFurnace).GetAll()) == 0 {
		t.Error("no furnace recipes loaded")
	}
	// Every loaded recipe only uses items that exist: its results are real items.
	for _, r := range m.GetCraftingRecipeIndex() {
		for _, result := range r.GetResultsFor(nil) {
			if result.IsNull() {
				t.Errorf("recipe %T has an air result", r)
			}
		}
	}
}

func TestFurnaceRecipeFromData(t *testing.T) {
	m := loadVanillaManager(t)
	recipe := m.GetFurnaceRecipeManager(tile.FurnaceTypeFurnace).Match(item.VanillaRawBeef())
	if recipe == nil || recipe.GetResult().GetTypeId() != item.VanillaSteak().GetTypeId() {
		t.Fatalf("raw beef should smelt into steak, got %v", recipe)
	}
}

func TestShapelessRecipeMatchesGridAndOutputs(t *testing.T) {
	m := NewCraftingManager()
	registered := 0
	m.AddRecipeRegisteredCallback(func() { registered++ })
	r, err := NewShapelessRecipe([]RecipeIngredient{NewMetaWildcardRecipeIngredient("minecraft:iron_ingot"), NewTagWildcardRecipeIngredient("minecraft:egg")}, []item.Item{item.VanillaFlintAndSteel()}, ShapelessRecipeTypeCrafting)
	if err != nil {
		t.Fatal(err)
	}
	m.RegisterShapelessRecipe(r)
	if registered != 1 {
		t.Error("registering a recipe should fire the callbacks")
	}
	grid := NewCraftingGrid(CraftingGridSizeSmall)
	grid.SetItem(0, item.VanillaEgg())
	grid.SetItem(3, item.VanillaIronIngot())
	if m.MatchRecipe(grid, []item.Item{item.VanillaFlintAndSteel()}) != r {
		t.Fatal("the shapeless recipe wasn't matched")
	}
	grid.SetItem(1, item.VanillaEgg())
	if r.MatchesCraftingGrid(grid) {
		t.Error("an extra item shouldn't match")
	}
	if m.GetCraftingRecipeFromIndex(0) != r || m.GetCraftingRecipeFromIndex(1) != nil {
		t.Error("recipe index lookup is wrong")
	}
}

func TestShapedRecipeMatchesMirrored(t *testing.T) {
	stick := item.VanillaStick()
	ingredient, err := NewExactRecipeIngredient(stick)
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewShapedRecipe([]string{"A ", "AA"}, map[byte]RecipeIngredient{'A': ingredient}, []item.Item{item.VanillaBowl()})
	if err != nil {
		t.Fatal(err)
	}
	grid := NewCraftingGrid(CraftingGridSizeBig)
	// Mirrored shape: " A" / "AA".
	grid.SetItem(1, item.VanillaStick())
	grid.SetItem(3, item.VanillaStick())
	grid.SetItem(4, item.VanillaStick())
	if !r.MatchesCraftingGrid(grid) {
		t.Error("the mirrored shape should match")
	}
	if _, err := NewShapedRecipe([]string{"AB"}, map[byte]RecipeIngredient{'A': ingredient}, nil); err == nil {
		t.Error("a symbol without an ingredient should be rejected")
	}
}

func TestTagWildcardIngredient(t *testing.T) {
	ingredient := NewTagWildcardRecipeIngredient("minecraft:egg")
	if !ingredient.Accepts(item.VanillaEgg()) {
		t.Error("an egg should be accepted by minecraft:egg")
	}
	if ingredient.Accepts(item.VanillaStick()) {
		t.Error("a stick shouldn't be accepted by minecraft:egg")
	}
}
