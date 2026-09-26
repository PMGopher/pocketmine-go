package crafting

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"path"

	"pocketmine-go/pocketmine/block/tile"
	craftingjson "pocketmine-go/pocketmine/crafting/json"
	"pocketmine-go/pocketmine/data/bedrock"
	blockconvert "pocketmine-go/pocketmine/data/bedrock/block/convert"
	bedrockitem "pocketmine-go/pocketmine/data/bedrock/item"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network/mcpe/convert"
)

// SavedDataLoadingError is pocketmine\data\SavedDataLoadingException for recipe data.
type SavedDataLoadingError struct{ Message string }

func (e *SavedDataLoadingError) Error() string { return e.Message }

// deserializeIngredient is a port of CraftingManagerFromDataHelper::deserializeIngredient: nil
// for an unknown item.
func deserializeIngredient(data craftingjson.RecipeIngredientData) (RecipeIngredient, error) {
	if data.Count != nil && *data.Count != 1 {
		//every case we've seen so far where this isn't the case, it's been a bug and the count was ignored anyway
		//e.g. gold blocks crafted from 9 ingots, but each input item individually had a count of 9
		return nil, &SavedDataLoadingError{Message: "Recipe inputs should have a count of exactly 1"}
	}
	if data.Tag != nil {
		return NewTagWildcardRecipeIngredient(*data.Tag), nil
	}

	if data.Meta != nil && *data.Meta == craftingjson.WildcardMetaValue {
		//this could be an unimplemented item, but it doesn't really matter, since the item shouldn't be able to
		//be obtained anyway - filtering unknown items is only really important for outputs, to prevent players
		//obtaining them
		return NewMetaWildcardRecipeIngredient(data.Name), nil
	}

	itemStack, err := deserializeItemStackFromFields(data.Name, data.Meta, data.Count, data.BlockStates, nil)
	if err != nil {
		return nil, err
	}
	if itemStack == nil {
		//probably unknown item
		return nil, nil
	}
	ingredient, err := NewExactRecipeIngredient(itemStack)
	if err != nil {
		return nil, &SavedDataLoadingError{Message: err.Error()}
	}
	return ingredient, nil
}

// DeserializeItemStack is a port of CraftingManagerFromDataHelper::deserializeItemStack: nil for
// an unknown item.
func DeserializeItemStack(data craftingjson.ItemStackData) (item.Item, error) {
	//count, name, block_name, block_states, meta, nbt, can_place_on, can_destroy
	return deserializeItemStackFromFields(data.Name, data.Meta, data.Count, data.BlockStates, data.NBT)
}

// deserializeItemStackFromFields is a port of
// CraftingManagerFromDataHelper::deserializeItemStackFromFields.
func deserializeItemStackFromFields(name string, meta, count *int, blockStatesRaw, nbtRaw *string) (item.Item, error) {
	m, c := 0, 1
	if meta != nil {
		m = *meta
	}
	if count != nil {
		c = *count
	}
	var blockStateData *bedrock.BlockStateData
	if blockName, ok := bedrockitem.GetBlockItemIdMap().LookupBlockID(name); ok {
		if m != 0 {
			return nil, &SavedDataLoadingError{Message: "Meta should not be specified for blockitems"}
		}
		states := map[string]any{}
		if blockStatesRaw != nil {
			raw, err := base64.StdEncoding.DecodeString(*blockStatesRaw)
			if err != nil {
				return nil, &SavedDataLoadingError{Message: fmt.Sprintf("invalid block states for %s: %v", name, err)}
			}
			root, _, err := nbt.NewLittleEndianSerializer().Read(raw, 0, 0)
			if err != nil {
				return nil, &SavedDataLoadingError{Message: fmt.Sprintf("invalid block states for %s: %v", name, err)}
			}
			tag, ok := root.GetTag().(*nbt.CompoundTag)
			if !ok {
				return nil, &SavedDataLoadingError{Message: fmt.Sprintf("block states for %s aren't a compound", name)}
			}
			if states, err = bedrock.BlockStatesFromNbt(tag); err != nil {
				return nil, &SavedDataLoadingError{Message: fmt.Sprintf("invalid block states for %s: %v", name, err)}
			}
		}
		blockStateData = &bedrock.BlockStateData{Name: blockName, States: states, Version: blockconvert.CurrentBlockStateVersion}
	}

	var namedTag *nbt.CompoundTag
	if nbtRaw != nil {
		raw, err := base64.StdEncoding.DecodeString(*nbtRaw)
		if err != nil {
			return nil, &SavedDataLoadingError{Message: fmt.Sprintf("invalid NBT for %s: %v", name, err)}
		}
		root, _, err := nbt.NewLittleEndianSerializer().Read(raw, 0, 0)
		if err != nil {
			return nil, &SavedDataLoadingError{Message: fmt.Sprintf("invalid NBT for %s: %v", name, err)}
		}
		tag, ok := root.GetTag().(*nbt.CompoundTag)
		if !ok {
			return nil, &SavedDataLoadingError{Message: fmt.Sprintf("NBT for %s isn't a compound", name)}
		}
		namedTag = tag
	}

	it, ok := convert.DeserializeItemType(name, m, blockStateData)
	if !ok {
		//probably unknown item
		return nil, nil
	}
	it.SetCount(c)
	if namedTag != nil {
		it.SetNamedTag(namedTag)
	}
	return it, nil
}

func loadRecipeFile[T any](fsys fs.FS, dir, name string) ([]T, error) {
	data, err := fs.ReadFile(fsys, path.Join(dir, name))
	if err != nil {
		return nil, err
	}
	result, err := craftingjson.LoadArray[T](data)
	if err != nil {
		return nil, &SavedDataLoadingError{Message: fmt.Sprintf("%s: %v", name, err)}
	}
	return result, nil
}

// MakeCraftingManager is a port of CraftingManagerFromDataHelper::make: a CraftingManager with the
// recipes in dir of fsys (pmmp BedrockData's recipes directory). Recipes with unknown items are
// skipped.
func MakeCraftingManager(fsys fs.FS, dir string) (*CraftingManager, error) {
	result := NewCraftingManager()

	shapeless, err := loadRecipeFile[craftingjson.ShapelessRecipeData](fsys, dir, "shapeless_crafting.json")
	if err != nil {
		return nil, err
	}
shapelessLoop:
	for _, recipe := range shapeless {
		var shapelessType *ShapelessRecipeType
		var furnaceType *FurnaceType
		set := func(t ShapelessRecipeType) { shapelessType = &t }
		setFurnace := func(t FurnaceType) { furnaceType = &t }
		switch recipe.Block {
		case "crafting_table":
			set(ShapelessRecipeTypeCrafting)
		case "stonecutter":
			set(ShapelessRecipeTypeStonecutter)
		case "smithing_table":
			set(ShapelessRecipeTypeSmithing)
		case "cartography_table":
			set(ShapelessRecipeTypeCartography)
		case "furnace":
			setFurnace(tile.FurnaceTypeFurnace)
		case "blast_furnace":
			setFurnace(tile.FurnaceTypeBlastFurnace)
		case "smoker":
			setFurnace(tile.FurnaceTypeSmoker)
		case "campfire":
			setFurnace(tile.FurnaceTypeCampfire)
		case "soul_campfire":
			setFurnace(tile.FurnaceTypeSoulCampfire)
		default:
			continue
		}
		var inputs []RecipeIngredient
		for _, inputData := range recipe.Input {
			input, err := deserializeIngredient(inputData)
			if err != nil {
				return nil, err
			}
			if input == nil { //unknown input item
				continue shapelessLoop
			}
			inputs = append(inputs, input)
		}
		var outputs []item.Item
		for _, outputData := range recipe.Output {
			output, err := DeserializeItemStack(outputData)
			if err != nil {
				return nil, err
			}
			if output == nil { //unknown output item
				continue shapelessLoop
			}
			outputs = append(outputs, output)
		}
		//TODO: check unlocking requirements - our current system doesn't support this

		if furnaceType != nil {
			if len(inputs) != 1 || len(outputs) != 1 {
				return nil, &SavedDataLoadingError{Message: "Furnace recipes must have exactly 1 input and 1 output"}
			}
			result.GetFurnaceRecipeManager(*furnaceType).Register(NewFurnaceRecipe(outputs[0], inputs[0]))
		} else {
			r, err := NewShapelessRecipe(inputs, outputs, *shapelessType)
			if err != nil {
				return nil, &SavedDataLoadingError{Message: err.Error()}
			}
			result.RegisterShapelessRecipe(r)
		}
	}

	shaped, err := loadRecipeFile[craftingjson.ShapedRecipeData](fsys, dir, "shaped_crafting.json")
	if err != nil {
		return nil, err
	}
shapedLoop:
	for _, recipe := range shaped {
		if recipe.Block != "crafting_table" { //TODO: filter others out for now to avoid breaking economics
			continue
		}
		inputs := map[byte]RecipeIngredient{}
		for symbol, inputData := range recipe.Input {
			if len(symbol) != 1 {
				return nil, &SavedDataLoadingError{Message: fmt.Sprintf("Invalid shaped recipe symbol %q", symbol)}
			}
			input, err := deserializeIngredient(inputData)
			if err != nil {
				return nil, err
			}
			if input == nil { //unknown input item
				continue shapedLoop
			}
			inputs[symbol[0]] = input
		}
		var outputs []item.Item
		for _, outputData := range recipe.Output {
			output, err := DeserializeItemStack(outputData)
			if err != nil {
				return nil, err
			}
			if output == nil { //unknown output item
				continue shapedLoop
			}
			outputs = append(outputs, output)
		}
		//TODO: check unlocking requirements - our current system doesn't support this
		r, err := NewShapedRecipe(recipe.Shape, inputs, outputs)
		if err != nil {
			return nil, &SavedDataLoadingError{Message: err.Error()}
		}
		result.RegisterShapedRecipe(r)
	}

	potionTypes, err := loadRecipeFile[craftingjson.PotionTypeRecipeData](fsys, dir, "potion_type.json")
	if err != nil {
		return nil, err
	}
	for _, recipe := range potionTypes {
		input, err := deserializeIngredient(recipe.Input)
		if err != nil {
			return nil, err
		}
		ingredient, err := deserializeIngredient(recipe.Ingredient)
		if err != nil {
			return nil, err
		}
		output, err := DeserializeItemStack(recipe.Output)
		if err != nil {
			return nil, err
		}
		if input == nil || ingredient == nil || output == nil {
			continue
		}
		result.RegisterPotionTypeRecipe(NewPotionTypeRecipe(input, ingredient, output))
	}

	containerChanges, err := loadRecipeFile[craftingjson.PotionContainerChangeRecipeData](fsys, dir, "potion_container_change.json")
	if err != nil {
		return nil, err
	}
	for _, recipe := range containerChanges {
		ingredient, err := deserializeIngredient(recipe.Ingredient)
		if err != nil {
			return nil, err
		}
		if ingredient == nil {
			continue
		}
		//TODO: this is a really awful way to just check if an ID is recognized ...
		inputKnown, _ := deserializeItemStackFromFields(recipe.InputItemName, nil, nil, nil, nil)
		outputKnown, _ := deserializeItemStackFromFields(recipe.OutputItemName, nil, nil, nil, nil)
		if inputKnown == nil || outputKnown == nil {
			//unknown item
			continue
		}
		result.RegisterPotionContainerChangeRecipe(NewPotionContainerChangeRecipe(recipe.InputItemName, ingredient, recipe.OutputItemName))
	}

	//TODO: smithing

	return result, nil
}
