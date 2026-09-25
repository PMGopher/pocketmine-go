package crafting

import "pocketmine-go/pocketmine/item"

// CraftingRecipe is a port of pocketmine\crafting\CraftingRecipe.
type CraftingRecipe interface {
	// GetIngredientList returns the items needed to craft this recipe. This MUST NOT include Air
	// items or items with a zero count.
	GetIngredientList() []RecipeIngredient
	// GetResultsFor returns the results this recipe will produce when the inputs in the given
	// crafting grid are consumed.
	GetResultsFor(grid *CraftingGrid) []item.Item
	// MatchesCraftingGrid returns whether the given crafting grid meets the requirements to craft
	// this recipe.
	MatchesCraftingGrid(grid *CraftingGrid) bool
}

func cloneItems(items []item.Item) []item.Item {
	result := make([]item.Item, len(items))
	for i, it := range items {
		result[i] = it.Clone()
	}
	return result
}
