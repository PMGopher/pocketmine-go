package crafting

import (
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/item"
)

// FurnaceType is pocketmine\crafting\FurnaceType (ported as tile.FurnaceType, which the furnace
// tiles need and this package can import).
type FurnaceType = tile.FurnaceType

// FurnaceRecipe is a port of pocketmine\crafting\FurnaceRecipe.
type FurnaceRecipe struct {
	result     item.Item
	ingredient RecipeIngredient
}

func NewFurnaceRecipe(result item.Item, ingredient RecipeIngredient) *FurnaceRecipe {
	return &FurnaceRecipe{result: result.Clone(), ingredient: ingredient}
}

func (r *FurnaceRecipe) GetInput() RecipeIngredient { return r.ingredient }
func (r *FurnaceRecipe) GetResult() item.Item       { return r.result.Clone() }

// FurnaceRecipeManager is a port of pocketmine\crafting\FurnaceRecipeManager.
type FurnaceRecipeManager struct {
	furnaceRecipes            []*FurnaceRecipe
	lookupCache               map[int]*FurnaceRecipe
	recipeRegisteredCallbacks []func(*FurnaceRecipe)
}

func NewFurnaceRecipeManager() *FurnaceRecipeManager {
	return &FurnaceRecipeManager{lookupCache: map[int]*FurnaceRecipe{}}
}

// AddRecipeRegisteredCallback is getRecipeRegisteredCallbacks()->add().
func (m *FurnaceRecipeManager) AddRecipeRegisteredCallback(callback func(*FurnaceRecipe)) {
	m.recipeRegisteredCallbacks = append(m.recipeRegisteredCallbacks, callback)
}

// GetAll is a port of FurnaceRecipeManager::getAll.
func (m *FurnaceRecipeManager) GetAll() []*FurnaceRecipe {
	return append([]*FurnaceRecipe(nil), m.furnaceRecipes...)
}

// Register is a port of FurnaceRecipeManager::register.
func (m *FurnaceRecipeManager) Register(recipe *FurnaceRecipe) {
	m.furnaceRecipes = append(m.furnaceRecipes, recipe)
	for _, callback := range m.recipeRegisteredCallbacks {
		callback(recipe)
	}
}

// Match is a port of FurnaceRecipeManager::match.
func (m *FurnaceRecipeManager) Match(input item.Item) *FurnaceRecipe {
	index := input.GetStateId()
	if recipe, ok := m.lookupCache[index]; ok {
		return recipe
	}
	for _, recipe := range m.furnaceRecipes {
		if recipe.GetInput().Accepts(input) {
			//remember that this item is accepted by this recipe, so we don't need to bruteforce it again
			m.lookupCache[index] = recipe
			return recipe
		}
	}
	return nil
}
