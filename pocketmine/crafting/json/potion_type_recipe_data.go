package craftingjson

// PotionTypeRecipeData is a port of pocketmine\crafting\json\PotionTypeRecipeData.
type PotionTypeRecipeData struct {
	Input      RecipeIngredientData `json:"input"`
	Ingredient RecipeIngredientData `json:"ingredient"`
	Output     ItemStackData        `json:"output"`
}
