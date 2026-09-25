package craftingjson

// PotionContainerChangeRecipeData is a port of
// pocketmine\crafting\json\PotionContainerChangeRecipeData.
type PotionContainerChangeRecipeData struct {
	InputItemName  string               `json:"input_item_name"`
	Ingredient     RecipeIngredientData `json:"ingredient"`
	OutputItemName string               `json:"output_item_name"`
}
