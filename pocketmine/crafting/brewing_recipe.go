package crafting

import (
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/network/mcpe/convert"
)

// BrewingRecipe is a port of pocketmine\crafting\BrewingRecipe.
type BrewingRecipe interface {
	// GetResultFor returns the brewed item, or nil if input isn't accepted.
	GetResultFor(input item.Item) item.Item
	GetIngredient() RecipeIngredient
}

// PotionTypeRecipe is a port of pocketmine\crafting\PotionTypeRecipe.
type PotionTypeRecipe struct {
	input      RecipeIngredient
	ingredient RecipeIngredient
	output     item.Item
}

func NewPotionTypeRecipe(input, ingredient RecipeIngredient, output item.Item) *PotionTypeRecipe {
	return &PotionTypeRecipe{input: input, ingredient: ingredient, output: output.Clone()}
}

func (r *PotionTypeRecipe) GetInput() RecipeIngredient      { return r.input }
func (r *PotionTypeRecipe) GetIngredient() RecipeIngredient { return r.ingredient }
func (r *PotionTypeRecipe) GetOutput() item.Item            { return r.output.Clone() }

// GetResultFor is a port of PotionTypeRecipe::getResultFor.
func (r *PotionTypeRecipe) GetResultFor(input item.Item) item.Item {
	if r.input.Accepts(input) {
		return r.GetOutput()
	}
	return nil
}

// PotionContainerChangeRecipe is a port of pocketmine\crafting\PotionContainerChangeRecipe.
type PotionContainerChangeRecipe struct {
	inputItemID  string
	ingredient   RecipeIngredient
	outputItemID string
}

func NewPotionContainerChangeRecipe(inputItemID string, ingredient RecipeIngredient, outputItemID string) *PotionContainerChangeRecipe {
	return &PotionContainerChangeRecipe{inputItemID: inputItemID, ingredient: ingredient, outputItemID: outputItemID}
}

func (r *PotionContainerChangeRecipe) GetInputItemId() string          { return r.inputItemID }
func (r *PotionContainerChangeRecipe) GetIngredient() RecipeIngredient { return r.ingredient }
func (r *PotionContainerChangeRecipe) GetOutputItemId() string         { return r.outputItemID }

// GetResultFor is a port of PotionContainerChangeRecipe::getResultFor: the input item's data
// deserialized as the output item type. (Meta variants can't be deserialized yet, see
// convert.DeserializeItemType.)
func (r *PotionContainerChangeRecipe) GetResultFor(input item.Item) item.Item {
	//TODO: this is a really awful hack, but there isn't another way for now
	//this relies on transforming the serialized item's ID, relying on the target item type's data being the same as the input.
	name, ok := convert.ItemTypeName(input)
	if !ok || name != r.inputItemID {
		return nil
	}
	result, ok := convert.DeserializeItemType(r.outputItemID, 0)
	if !ok {
		return nil
	}
	if input.HasNamedTag() {
		result.SetNamedTag(input.GetNamedTag())
	}
	return result
}

var (
	_ BrewingRecipe = (*PotionTypeRecipe)(nil)
	_ BrewingRecipe = (*PotionContainerChangeRecipe)(nil)
)
