package craftingjson

// SmithingTransformRecipeData is a port of pocketmine\crafting\json\SmithingTransformRecipeData.
type SmithingTransformRecipeData struct {
	Template RecipeIngredientData `json:"template"`
	Input    RecipeIngredientData `json:"input"`
	Addition RecipeIngredientData `json:"addition"`
	Output   ItemStackData        `json:"output"`
	Block    string               `json:"block"`
}
