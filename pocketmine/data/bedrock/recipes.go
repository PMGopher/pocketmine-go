package bedrock

import "embed"

// Recipes holds pmmp BedrockData's recipes/*.json (the files CraftingManagerFromDataHelper::make
// reads: shaped/shapeless crafting, potion type and potion container change recipes), at the
// BedrockData version PocketMine-MP 5.44.4 pins (6.7.0+bedrock-1.26.30; CC0, see
// assets/LICENSE-bedrockdata).
//
//go:embed assets/recipes/*.json
var Recipes embed.FS

// RecipesDir is the directory of Recipes holding the recipe files (BedrockDataFiles::RECIPES).
const RecipesDir = "assets/recipes"
