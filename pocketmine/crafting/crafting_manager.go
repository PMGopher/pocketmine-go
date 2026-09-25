package crafting

import (
	"iter"
	"sort"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
)

// CraftingManager is a port of pocketmine\crafting\CraftingManager.
type CraftingManager struct {
	// shapedRecipes/shapelessRecipes are output hash => recipes.
	shapedRecipes    map[string][]*ShapedRecipe
	shapelessRecipes map[string][]*ShapelessRecipe

	craftingRecipeIndex []CraftingRecipe

	furnaceRecipeManagers map[FurnaceType]*FurnaceRecipeManager

	potionTypeRecipes            []*PotionTypeRecipe
	potionContainerChangeRecipes []*PotionContainerChangeRecipe

	// brewingRecipeCache is input state ID => ingredient state ID => recipe.
	brewingRecipeCache map[int]map[int]BrewingRecipe

	recipeRegisteredCallbacks []func()
}

// NewCraftingManager is a port of CraftingManager::__construct.
func NewCraftingManager() *CraftingManager {
	m := &CraftingManager{
		shapedRecipes:         map[string][]*ShapedRecipe{},
		shapelessRecipes:      map[string][]*ShapelessRecipe{},
		furnaceRecipeManagers: map[FurnaceType]*FurnaceRecipeManager{},
		brewingRecipeCache:    map[int]map[int]BrewingRecipe{},
	}
	for _, furnaceType := range tile.AllFurnaceTypes {
		manager := NewFurnaceRecipeManager()
		manager.AddRecipeRegisteredCallback(func(*FurnaceRecipe) { m.fireRecipeRegistered() })
		m.furnaceRecipeManagers[furnaceType] = manager
	}
	return m
}

// AddRecipeRegisteredCallback is getRecipeRegisteredCallbacks()->add().
func (m *CraftingManager) AddRecipeRegisteredCallback(callback func()) {
	m.recipeRegisteredCallbacks = append(m.recipeRegisteredCallbacks, callback)
}

func (m *CraftingManager) fireRecipeRegistered() {
	for _, callback := range m.recipeRegisteredCallbacks {
		callback()
	}
}

// hashOutput is a port of CraftingManager::hashOutput: the item's state ID and NBT.
func hashOutput(output item.Item) string {
	var b strings.Builder
	b.WriteString(strconv.Itoa(output.GetStateId()))
	b.WriteByte(0)
	if root, err := nbt.NewTreeRoot(output.GetNamedTag(), ""); err == nil {
		if data, err := nbt.NewLittleEndianSerializer().Write(root); err == nil {
			b.Write(data)
		}
	}
	return b.String()
}

// hashOutputs is a port of CraftingManager::hashOutputs.
func hashOutputs(outputs []item.Item) string {
	if len(outputs) == 1 {
		return hashOutput(outputs[0])
	}
	unique := map[string]bool{}
	for _, o := range outputs {
		//count is not written because the outputs might be from multiple repetitions of a single recipe
		//this reduces the accuracy of the hash, but it won't matter in most cases.
		unique[hashOutput(o)] = true
	}
	keys := make([]string, 0, len(unique))
	for k := range unique {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, "")
}

// GetShapelessRecipes is a port of CraftingManager::getShapelessRecipes.
func (m *CraftingManager) GetShapelessRecipes() map[string][]*ShapelessRecipe {
	return m.shapelessRecipes
}

// GetShapedRecipes is a port of CraftingManager::getShapedRecipes.
func (m *CraftingManager) GetShapedRecipes() map[string][]*ShapedRecipe { return m.shapedRecipes }

// GetCraftingRecipeIndex is a port of CraftingManager::getCraftingRecipeIndex.
func (m *CraftingManager) GetCraftingRecipeIndex() []CraftingRecipe {
	return append([]CraftingRecipe(nil), m.craftingRecipeIndex...)
}

// GetCraftingRecipeFromIndex is a port of CraftingManager::getCraftingRecipeFromIndex.
func (m *CraftingManager) GetCraftingRecipeFromIndex(index int) CraftingRecipe {
	if index < 0 || index >= len(m.craftingRecipeIndex) {
		return nil
	}
	return m.craftingRecipeIndex[index]
}

// GetFurnaceRecipeManager is a port of CraftingManager::getFurnaceRecipeManager.
func (m *CraftingManager) GetFurnaceRecipeManager(furnaceType FurnaceType) *FurnaceRecipeManager {
	return m.furnaceRecipeManagers[furnaceType]
}

// GetPotionTypeRecipes is a port of CraftingManager::getPotionTypeRecipes.
func (m *CraftingManager) GetPotionTypeRecipes() []*PotionTypeRecipe {
	return append([]*PotionTypeRecipe(nil), m.potionTypeRecipes...)
}

// GetPotionContainerChangeRecipes is a port of CraftingManager::getPotionContainerChangeRecipes.
func (m *CraftingManager) GetPotionContainerChangeRecipes() []*PotionContainerChangeRecipe {
	return append([]*PotionContainerChangeRecipe(nil), m.potionContainerChangeRecipes...)
}

// RegisterShapedRecipe is a port of CraftingManager::registerShapedRecipe.
func (m *CraftingManager) RegisterShapedRecipe(recipe *ShapedRecipe) {
	hash := hashOutputs(recipe.GetResults())
	m.shapedRecipes[hash] = append(m.shapedRecipes[hash], recipe)
	m.craftingRecipeIndex = append(m.craftingRecipeIndex, recipe)
	m.fireRecipeRegistered()
}

// RegisterShapelessRecipe is a port of CraftingManager::registerShapelessRecipe.
func (m *CraftingManager) RegisterShapelessRecipe(recipe *ShapelessRecipe) {
	hash := hashOutputs(recipe.GetResults())
	m.shapelessRecipes[hash] = append(m.shapelessRecipes[hash], recipe)
	m.craftingRecipeIndex = append(m.craftingRecipeIndex, recipe)
	m.fireRecipeRegistered()
}

// RegisterPotionTypeRecipe is a port of CraftingManager::registerPotionTypeRecipe.
func (m *CraftingManager) RegisterPotionTypeRecipe(recipe *PotionTypeRecipe) {
	m.potionTypeRecipes = append(m.potionTypeRecipes, recipe)
	m.fireRecipeRegistered()
}

// RegisterPotionContainerChangeRecipe is a port of
// CraftingManager::registerPotionContainerChangeRecipe.
func (m *CraftingManager) RegisterPotionContainerChangeRecipe(recipe *PotionContainerChangeRecipe) {
	m.potionContainerChangeRecipes = append(m.potionContainerChangeRecipes, recipe)
	m.fireRecipeRegistered()
}

// MatchRecipe is a port of CraftingManager::matchRecipe.
func (m *CraftingManager) MatchRecipe(grid *CraftingGrid, outputs []item.Item) CraftingRecipe {
	//TODO: try to match special recipes before anything else (first they need to be implemented!)
	outputHash := hashOutputs(outputs)
	for _, recipe := range m.shapedRecipes[outputHash] {
		if recipe.MatchesCraftingGrid(grid) {
			return recipe
		}
	}
	for _, recipe := range m.shapelessRecipes[outputHash] {
		if recipe.MatchesCraftingGrid(grid) {
			return recipe
		}
	}
	return nil
}

// MatchRecipeByOutputs is a port of CraftingManager::matchRecipeByOutputs.
func (m *CraftingManager) MatchRecipeByOutputs(outputs []item.Item) iter.Seq[CraftingRecipe] {
	//TODO: try to match special recipes before anything else (first they need to be implemented!)
	outputHash := hashOutputs(outputs)
	return func(yield func(CraftingRecipe) bool) {
		for _, recipe := range m.shapedRecipes[outputHash] {
			if !yield(recipe) {
				return
			}
		}
		for _, recipe := range m.shapelessRecipes[outputHash] {
			if !yield(recipe) {
				return
			}
		}
	}
}

// MatchBrewingRecipe is a port of CraftingManager::matchBrewingRecipe.
func (m *CraftingManager) MatchBrewingRecipe(input, ingredient item.Item) BrewingRecipe {
	inputHash := input.GetStateId()
	ingredientHash := ingredient.GetStateId()
	if cached := m.brewingRecipeCache[inputHash][ingredientHash]; cached != nil {
		return cached
	}
	cache := func(recipe BrewingRecipe) BrewingRecipe {
		if m.brewingRecipeCache[inputHash] == nil {
			m.brewingRecipeCache[inputHash] = map[int]BrewingRecipe{}
		}
		m.brewingRecipeCache[inputHash][ingredientHash] = recipe
		return recipe
	}
	for _, recipe := range m.potionContainerChangeRecipes {
		if recipe.GetIngredient().Accepts(ingredient) && recipe.GetResultFor(input) != nil {
			return cache(recipe)
		}
	}
	for _, recipe := range m.potionTypeRecipes {
		if recipe.GetIngredient().Accepts(ingredient) && recipe.GetResultFor(input) != nil {
			return cache(recipe)
		}
	}
	return nil
}
