package crafting

import (
	"fmt"
	"testing"

	"pocketmine-go/pocketmine/data/bedrock"
)

func TestZZLoad(t *testing.T) {
	m, err := MakeCraftingManager(bedrock.Recipes, bedrock.RecipesDir)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("crafting", len(m.GetCraftingRecipeIndex()), "potion", len(m.GetPotionTypeRecipes()), "container", len(m.GetPotionContainerChangeRecipes()))
	for _, ft := range []FurnaceType{0, 1, 2, 3, 4} {
		fmt.Println("furnace", ft, len(m.GetFurnaceRecipeManager(ft).GetAll()))
	}
	for i, r := range m.GetCraftingRecipeIndex() {
		if i < 15 {
			switch r := r.(type) {
			case *ShapedRecipe:
				fmt.Println("shaped", r.GetShape(), r.GetResults())
			case *ShapelessRecipe:
				fmt.Println("shapeless", r.GetIngredientList(), r.GetResults())
			}
		}
	}
}
