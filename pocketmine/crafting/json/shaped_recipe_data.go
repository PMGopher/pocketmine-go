package craftingjson

// ShapedRecipeData is a port of pocketmine\crafting\json\ShapedRecipeData.
type ShapedRecipeData struct {
	Shape                []string                        `json:"shape"`
	Input                map[string]RecipeIngredientData `json:"input"`
	Output               []ItemStackData                 `json:"output"`
	Block                string                          `json:"block"`
	Priority             int                             `json:"priority"`
	UnlockingIngredients []RecipeIngredientData          `json:"unlockingIngredients,omitempty"`
}
