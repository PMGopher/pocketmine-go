package craftingjson

// WildcardMetaValue is RecipeIngredientData::WILDCARD_META_VALUE.
const WildcardMetaValue = 32767

// RecipeIngredientData is a port of pocketmine\crafting\json\RecipeIngredientData.
type RecipeIngredientData struct {
	Name             string  `json:"name,omitempty"`
	Meta             *int    `json:"meta,omitempty"`
	BlockStates      *string `json:"block_states,omitempty"`
	Tag              *string `json:"tag,omitempty"`
	MolangExpression *string `json:"molang_expression,omitempty"`
	MolangVersion    *int    `json:"molang_version,omitempty"`
	Count            *int    `json:"count,omitempty"`
}
