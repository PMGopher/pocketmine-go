package crafting

import (
	"fmt"

	"pocketmine-go/pocketmine/item"
)

// ShapelessRecipeType is a port of pocketmine\crafting\ShapelessRecipeType.
type ShapelessRecipeType int

const (
	ShapelessRecipeTypeCrafting ShapelessRecipeType = iota
	ShapelessRecipeTypeStonecutter
	ShapelessRecipeTypeSmithing
	ShapelessRecipeTypeCartography
)

// ShapelessRecipe is a port of pocketmine\crafting\ShapelessRecipe.
type ShapelessRecipe struct {
	ingredients []RecipeIngredient
	results     []item.Item
	recipeType  ShapelessRecipeType
}

// NewShapelessRecipe is a port of ShapelessRecipe::__construct: no more than 9 ingredients.
func NewShapelessRecipe(ingredients []RecipeIngredient, results []item.Item, recipeType ShapelessRecipeType) (*ShapelessRecipe, error) {
	if len(ingredients) > 9 {
		return nil, fmt.Errorf("Shapeless recipes cannot have more than 9 ingredients")
	}
	return &ShapelessRecipe{ingredients: append([]RecipeIngredient(nil), ingredients...), results: cloneItems(results), recipeType: recipeType}, nil
}

// GetResults is a port of ShapelessRecipe::getResults.
func (r *ShapelessRecipe) GetResults() []item.Item { return cloneItems(r.results) }

// GetResultsFor is a port of ShapelessRecipe::getResultsFor.
func (r *ShapelessRecipe) GetResultsFor(grid *CraftingGrid) []item.Item { return r.GetResults() }

func (r *ShapelessRecipe) GetType() ShapelessRecipeType { return r.recipeType }

// GetIngredientList is a port of ShapelessRecipe::getIngredientList.
func (r *ShapelessRecipe) GetIngredientList() []RecipeIngredient {
	return append([]RecipeIngredient(nil), r.ingredients...)
}

func (r *ShapelessRecipe) GetIngredientCount() int { return len(r.ingredients) }

// MatchesCraftingGrid is a port of ShapelessRecipe::matchesCraftingGrid.
func (r *ShapelessRecipe) MatchesCraftingGrid(grid *CraftingGrid) bool {
	//don't pack the ingredients - shapeless recipes require that each ingredient be in a separate slot
	input := grid.GetContents(false)

outer:
	for _, ingredient := range r.ingredients {
		for j, haveItem := range input {
			if ingredient.Accepts(haveItem) {
				delete(input, j)
				continue outer
			}
		}
		return false //failed to match the needed item to a given item
	}

	return len(input) == 0 //crafting grid should be empty apart from the given ingredient stacks
}

var _ CraftingRecipe = (*ShapelessRecipe)(nil)
