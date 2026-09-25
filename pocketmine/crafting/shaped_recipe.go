package crafting

import (
	"fmt"
	"strings"

	"pocketmine-go/pocketmine/item"
)

// ShapedRecipe is a port of pocketmine\crafting\ShapedRecipe.
type ShapedRecipe struct {
	shape          []string
	ingredientList map[byte]RecipeIngredient
	results        []item.Item
	height         int
	width          int
}

// NewShapedRecipe is a port of ShapedRecipe::__construct. shape is 1 to 3 rows of equal length (at
// most 3); each character is an ingredient key, spaces are air. ingredients maps every character
// used to its ingredient. Recipes don't need to be square: don't add padding for empty rows or
// columns.
func NewShapedRecipe(shape []string, ingredients map[byte]RecipeIngredient, results []item.Item) (*ShapedRecipe, error) {
	r := &ShapedRecipe{ingredientList: map[byte]RecipeIngredient{}}
	r.height = len(shape)
	if r.height > 3 || r.height <= 0 {
		return nil, fmt.Errorf("Shaped recipes may only have 1, 2 or 3 rows, not %d", r.height)
	}

	r.width = len(shape[0])
	if r.width > 3 || r.width <= 0 {
		return nil, fmt.Errorf("Shaped recipes may only have 1, 2 or 3 columns, not %d", r.width)
	}

	for _, row := range shape {
		if len(row) != r.width {
			return nil, fmt.Errorf("Shaped recipe rows must all have the same length (expected %d, got %d)", r.width, len(row))
		}
		for x := 0; x < r.width; x++ {
			if _, ok := ingredients[row[x]]; row[x] != ' ' && !ok {
				return nil, fmt.Errorf("No item specified for symbol '%c'", row[x])
			}
		}
	}

	r.shape = append([]string(nil), shape...)

	joined := strings.Join(r.shape, "")
	for char, ingredient := range ingredients {
		if !strings.ContainsRune(joined, rune(char)) {
			return nil, fmt.Errorf("Symbol '%c' does not appear in the recipe shape", char)
		}
		r.ingredientList[char] = ingredient
	}

	r.results = cloneItems(results)
	return r, nil
}

func (r *ShapedRecipe) GetWidth() int  { return r.width }
func (r *ShapedRecipe) GetHeight() int { return r.height }

// GetResults is a port of ShapedRecipe::getResults.
func (r *ShapedRecipe) GetResults() []item.Item { return cloneItems(r.results) }

// GetResultsFor is a port of ShapedRecipe::getResultsFor.
func (r *ShapedRecipe) GetResultsFor(grid *CraftingGrid) []item.Item { return r.GetResults() }

// GetIngredientMap is a port of ShapedRecipe::getIngredientMap: [y][x], nil for air.
func (r *ShapedRecipe) GetIngredientMap() [][]RecipeIngredient {
	ingredients := make([][]RecipeIngredient, r.height)
	for y := 0; y < r.height; y++ {
		ingredients[y] = make([]RecipeIngredient, r.width)
		for x := 0; x < r.width; x++ {
			ingredients[y][x] = r.GetIngredient(x, y)
		}
	}
	return ingredients
}

// GetIngredientList is a port of ShapedRecipe::getIngredientList.
func (r *ShapedRecipe) GetIngredientList() []RecipeIngredient {
	var ingredients []RecipeIngredient
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			if ingredient := r.GetIngredient(x, y); ingredient != nil {
				ingredients = append(ingredients, ingredient)
			}
		}
	}
	return ingredients
}

// GetIngredient is a port of ShapedRecipe::getIngredient: nil for air.
func (r *ShapedRecipe) GetIngredient(x, y int) RecipeIngredient {
	return r.ingredientList[r.shape[y][x]]
}

// GetShape is a port of ShapedRecipe::getShape.
func (r *ShapedRecipe) GetShape() []string { return append([]string(nil), r.shape...) }

func (r *ShapedRecipe) matchInputMap(grid *CraftingGrid, reverse bool) bool {
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			gx := x
			if reverse {
				gx = r.width - x - 1
			}
			given := grid.GetIngredient(gx, y)
			required := r.GetIngredient(x, y)

			if required == nil {
				if !given.IsNull() {
					return false //hole, such as that in the center of a chest recipe, should not be filled
				}
			} else if !required.Accepts(given) {
				return false
			}
		}
	}
	return true
}

// MatchesCraftingGrid is a port of ShapedRecipe::matchesCraftingGrid.
func (r *ShapedRecipe) MatchesCraftingGrid(grid *CraftingGrid) bool {
	if r.width != grid.GetRecipeWidth() || r.height != grid.GetRecipeHeight() {
		return false
	}
	return r.matchInputMap(grid, false) || r.matchInputMap(grid, true)
}

var _ CraftingRecipe = (*ShapedRecipe)(nil)
