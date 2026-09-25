package craftingjson

// ShapelessRecipeData is a port of pocketmine\crafting\json\ShapelessRecipeData.
type ShapelessRecipeData struct {
	Input                []RecipeIngredientData `json:"input"`
	Output               []ItemStackData        `json:"output"`
	Block                string                 `json:"block"`
	Priority             int                    `json:"priority"`
	UnlockingIngredients []RecipeIngredientData `json:"unlockingIngredients,omitempty"`
}
