package craftingjson

// SmithingTrimRecipeData is a port of pocketmine\crafting\json\SmithingTrimRecipeData.
type SmithingTrimRecipeData struct {
	Template RecipeIngredientData `json:"template"`
	Input    RecipeIngredientData `json:"input"`
	Addition RecipeIngredientData `json:"addition"`
	Block    string               `json:"block"`
}
