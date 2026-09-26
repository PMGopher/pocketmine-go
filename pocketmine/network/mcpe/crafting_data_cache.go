package mcpe

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/crafting"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/timings"
)

// RecipeIDOffset is CraftingDataCache::RECIPE_ID_OFFSET: the client doesn't like recipes with ID 0
// (as of 1.21.100) and complains about them in the content log. This doesn't actually affect the
// function of the recipe, but it is annoying, so this offset fixes it.
const RecipeIDOffset = 1

// recipeInputWildcardMeta is TypeConverter::RECIPE_INPUT_WILDCARD_META.
const recipeInputWildcardMeta = 0x7fff

// CraftingDataCache is a port of pocketmine\network\mcpe\cache\CraftingDataCache.
type CraftingDataCache struct {
	mu     sync.Mutex
	caches map[*crafting.CraftingManager]*packet.CraftingData
}

var craftingDataCache = &CraftingDataCache{caches: map[*crafting.CraftingManager]*packet.CraftingData{}}

// GetCraftingDataCache is CraftingDataCache::getInstance.
func GetCraftingDataCache() *CraftingDataCache { return craftingDataCache }

// GetCache is a port of CraftingDataCache::getCache.
func (c *CraftingDataCache) GetCache(manager *crafting.CraftingManager) *packet.CraftingData {
	c.mu.Lock()
	defer c.mu.Unlock()
	if pk, ok := c.caches[manager]; ok {
		return pk
	}
	manager.AddRecipeRegisteredCallback(func() {
		c.mu.Lock()
		delete(c.caches, manager)
		c.mu.Unlock()
	})
	pk := buildCraftingDataCache(manager)
	c.caches[manager] = pk
	return pk
}

// CoreRecipeIngredientToNet is a port of TypeConverter::coreRecipeIngredientToNet (nil is air).
func CoreRecipeIngredientToNet(ingredient crafting.RecipeIngredient) protocol.ItemDescriptorCount {
	switch i := ingredient.(type) {
	case nil:
		return protocol.ItemDescriptorCount{Descriptor: &protocol.InvalidItemDescriptor{}, Count: 0}
	case *crafting.MetaWildcardRecipeIngredient:
		return protocol.ItemDescriptorCount{Descriptor: &protocol.DefaultItemDescriptor{Name: i.GetItemId(), MetadataValue: recipeInputWildcardMeta}, Count: 1}
	case *crafting.ExactRecipeIngredient:
		it := i.GetItem()
		name, ok := convert.ItemTypeName(it)
		if !ok {
			panic(fmt.Sprintf("recipe ingredient %v has no network item name", it))
		}
		return protocol.ItemDescriptorCount{Descriptor: &protocol.DefaultItemDescriptor{Name: name, MetadataValue: 0}, Count: 1}
	case *crafting.TagWildcardRecipeIngredient:
		return protocol.ItemDescriptorCount{Descriptor: &protocol.ItemTagItemDescriptor{Tag: i.GetTagName()}, Count: 1}
	}
	panic(fmt.Sprintf("Unsupported recipe ingredient type %T", ingredient))
}

func recipeIDString(recipeNetID int) string {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], uint32(recipeNetID)) //TODO: this should probably be changed to something human-readable
	return string(b[:])
}

// buildCraftingDataCache is a port of CraftingDataCache::buildCraftingDataCache.
func buildCraftingDataCache(manager *crafting.CraftingManager) *packet.CraftingData {
	timings.Init()
	timings.CraftingDataCacheRebuild.StartTiming()
	defer timings.CraftingDataCacheRebuild.StopTiming()

	pk := &packet.CraftingData{ClearRecipes: true}
	// RecipeUnlockingRequirement(null): no requirement.
	noUnlockingRequirement := protocol.Optional[protocol.RecipeUnlockRequirement]{}

	recipeNetID := RecipeIDOffset
	for index, recipe := range manager.GetCraftingRecipeIndex() {
		//the client doesn't like recipes with an ID of 0, so we need to offset them
		recipeNetID = index + RecipeIDOffset
		switch r := recipe.(type) {
		case *crafting.ShapelessRecipe:
			typeTag := "crafting_table"
			switch r.GetType() {
			case crafting.ShapelessRecipeTypeStonecutter:
				typeTag = "stonecutter"
			case crafting.ShapelessRecipeTypeCartography:
				typeTag = "cartography_table"
			case crafting.ShapelessRecipeTypeSmithing:
				typeTag = "smithing_table"
			}
			var inputs []protocol.ItemDescriptorCount
			for _, ingredient := range r.GetIngredientList() {
				inputs = append(inputs, CoreRecipeIngredientToNet(ingredient))
			}
			var outputs []protocol.ItemStack
			for _, result := range r.GetResults() {
				outputs = append(outputs, convert.CoreItemStackToNet(result))
			}
			pk.ShapelessRecipes = append(pk.ShapelessRecipes, protocol.ShapelessRecipe{
				RecipeID:          recipeIDString(recipeNetID),
				Input:             inputs,
				Output:            outputs,
				UUID:              uuid.Nil,
				Block:             typeTag,
				Priority:          50,
				UnlockRequirement: noUnlockingRequirement,
				RecipeNetworkID:   uint32(recipeNetID),
			})
		case *crafting.ShapedRecipe:
			var inputs []protocol.ItemDescriptorCount
			for row := 0; row < r.GetHeight(); row++ {
				for column := 0; column < r.GetWidth(); column++ {
					inputs = append(inputs, CoreRecipeIngredientToNet(r.GetIngredient(column, row)))
				}
			}
			var outputs []protocol.ItemStack
			for _, result := range r.GetResults() {
				outputs = append(outputs, convert.CoreItemStackToNet(result))
			}
			pk.ShapedRecipes = append(pk.ShapedRecipes, protocol.ShapedRecipe{
				RecipeID:          recipeIDString(recipeNetID),
				Width:             int32(r.GetWidth()),
				Height:            int32(r.GetHeight()),
				Input:             inputs,
				Output:            outputs,
				UUID:              uuid.Nil,
				Block:             "crafting_table",
				Priority:          50,
				AssumeSymmetry:    true,
				UnlockRequirement: noUnlockingRequirement,
				RecipeNetworkID:   uint32(recipeNetID),
			})
		default:
			//TODO: probably special recipe types
		}
	}

	furnaceBlockNames := map[tile.FurnaceType]string{
		tile.FurnaceTypeFurnace:      "furnace",
		tile.FurnaceTypeBlastFurnace: "blast_furnace",
		tile.FurnaceTypeSmoker:       "smoker",
		tile.FurnaceTypeCampfire:     "campfire",
		tile.FurnaceTypeSoulCampfire: "soul_campfire",
	}
	for _, furnaceType := range tile.AllFurnaceTypes {
		for _, recipe := range manager.GetFurnaceRecipeManager(furnaceType).GetAll() {
			recipeNetID++
			pk.ShapelessRecipes = append(pk.ShapelessRecipes, protocol.ShapelessRecipe{
				RecipeID:          recipeIDString(recipeNetID),
				Input:             []protocol.ItemDescriptorCount{CoreRecipeIngredientToNet(recipe.GetInput())},
				Output:            []protocol.ItemStack{convert.CoreItemStackToNet(recipe.GetResult())},
				UUID:              uuid.Nil,
				Block:             furnaceBlockNames[furnaceType],
				Priority:          50,
				UnlockRequirement: noUnlockingRequirement,
				RecipeNetworkID:   uint32(recipeNetID), //not used, but we need to fill them with something unique regardless
			})
		}
	}

	intIDMeta := func(ingredient crafting.RecipeIngredient) (int32, int32) {
		descriptor, ok := CoreRecipeIngredientToNet(ingredient).Descriptor.(*protocol.DefaultItemDescriptor)
		if !ok {
			panic("potion recipe ingredients must be item descriptors")
		}
		id, _ := bedrock.ItemRuntimeIDFor(descriptor.Name)
		return id, descriptor.MetadataValue
	}
	for _, recipe := range manager.GetPotionTypeRecipes() {
		inputID, inputMeta := intIDMeta(recipe.GetInput())
		ingredientID, ingredientMeta := intIDMeta(recipe.GetIngredient())
		output := convert.CoreItemStackToNet(recipe.GetOutput())
		pk.PotionRecipes = append(pk.PotionRecipes, protocol.PotionRecipe{
			InputPotionID:        inputID,
			InputPotionMetadata:  inputMeta,
			ReagentItemID:        ingredientID,
			ReagentItemMetadata:  ingredientMeta,
			OutputPotionID:       output.NetworkID,
			OutputPotionMetadata: int32(output.MetadataValue),
		})
	}
	for _, recipe := range manager.GetPotionContainerChangeRecipes() {
		inputID, _ := bedrock.ItemRuntimeIDFor(recipe.GetInputItemId())
		ingredientID, _ := intIDMeta(recipe.GetIngredient())
		outputID, _ := bedrock.ItemRuntimeIDFor(recipe.GetOutputItemId())
		pk.PotionContainerChangeRecipes = append(pk.PotionContainerChangeRecipes, protocol.PotionContainerChangeRecipe{
			InputItemID:   inputID,
			ReagentItemID: ingredientID,
			OutputItemID:  outputID,
		})
	}
	return pk
}
