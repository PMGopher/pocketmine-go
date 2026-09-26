package transaction

import (
	"fmt"
	"sort"

	"pocketmine-go/pocketmine/crafting"
	"pocketmine-go/pocketmine/event"
	inventoryevent "pocketmine-go/pocketmine/event/inventory"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

// craftingGridOwner is Player::getCraftingGrid for CraftingTransaction.
type craftingGridOwner interface {
	GetCraftingGrid() inventory.TemporaryInventory
}

// CraftingTransaction is a port of pocketmine\inventory\transaction\CraftingTransaction: a
// transaction that consumes crafting ingredients and produces the recipe's results.
//
// Crafting transactions come from ItemStackRequests: the client says which recipe it crafts, and
// the transaction checks that the items it moved match that recipe (or, without a recipe, finds
// one matching the outputs).
type CraftingTransaction struct {
	*InventoryTransaction

	recipe      crafting.CraftingRecipe
	repetitions *int

	inputs  []item.Item
	outputs []item.Item

	craftingManager *crafting.CraftingManager
}

// NewCraftingTransaction is a port of CraftingTransaction::__construct. recipe and repetitions
// may be nil/0 to have them worked out from the actions.
func NewCraftingTransaction(source Player, craftingManager *crafting.CraftingManager, actions []InventoryAction, recipe crafting.CraftingRecipe, repetitions int) (*CraftingTransaction, error) {
	t := &CraftingTransaction{InventoryTransaction: &InventoryTransaction{}, craftingManager: craftingManager, recipe: recipe}
	if repetitions > 0 {
		r := repetitions
		t.repetitions = &r
	}
	t.Init(t, source)
	for _, a := range actions {
		if err := t.AddAction(a); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (t *CraftingTransaction) GetRecipe() crafting.CraftingRecipe { return t.recipe }

// packItems is a port of CraftingTransaction::packItems.
func packItems(providedItems []item.Item) []item.Item {
	var packed []item.Item
	for len(providedItems) > 0 {
		it := providedItems[len(providedItems)-1]
		providedItems = providedItems[:len(providedItems)-1]
		remaining := providedItems[:0]
		for _, other := range providedItems {
			if it.CanStackWith(other) {
				it.SetCount(it.GetCount() + other.GetCount())
			} else {
				remaining = append(remaining, other)
			}
		}
		providedItems = remaining
		packed = append(packed, it)
	}
	return packed
}

// MatchIngredients is a port of CraftingTransaction::matchIngredients: checks the provided items
// satisfy the recipe ingredients expectedIterations times, with nothing left over.
func MatchIngredients(providedItems []item.Item, recipeIngredients []crafting.RecipeIngredient, expectedIterations int) error {
	if len(recipeIngredients) == 0 {
		return validationError("No recipe ingredients given")
	}
	if len(providedItems) == 0 {
		return validationError("No transaction items given")
	}

	cloned := make([]item.Item, len(providedItems))
	for i, it := range providedItems {
		cloned[i] = it.Clone()
	}
	packedProvidedItems := map[int]item.Item{}
	for i, it := range packItems(cloned) {
		packedProvidedItems[i] = it
	}
	packedIndexes := make([]int, 0, len(packedProvidedItems))
	for i := range packedProvidedItems {
		packedIndexes = append(packedIndexes, i)
	}
	sort.Ints(packedIndexes)
	packedProvidedItemMatches := map[int]int{}

	type ingredientMatch struct {
		ingredientIndex int
		acceptedItems   map[int]bool
	}
	var recipeIngredientMatches []ingredientMatch

	for ingredientIndex, recipeIngredient := range recipeIngredients {
		acceptedItems := map[int]bool{}
		for _, itemIndex := range packedIndexes {
			if recipeIngredient.Accepts(packedProvidedItems[itemIndex]) {
				packedProvidedItemMatches[itemIndex]++
				acceptedItems[itemIndex] = true
			}
		}
		if len(acceptedItems) == 0 {
			return validationError(fmt.Sprintf("No provided items satisfy ingredient requirement %s", recipeIngredient))
		}
		recipeIngredientMatches = append(recipeIngredientMatches, ingredientMatch{ingredientIndex, acceptedItems})
	}

	for _, itemIndex := range packedIndexes {
		if packedProvidedItemMatches[itemIndex] == 0 {
			return validationError(fmt.Sprintf("Provided item %v is not accepted by any recipe ingredient", packedProvidedItems[itemIndex]))
		}
	}

	//Most picky ingredients first - avoid picky ingredient getting their items stolen by wildcard ingredients
	//TODO: this is still insufficient when multiple wildcard ingredients have overlaps, but we don't (yet) have to
	//worry about those.
	sort.SliceStable(recipeIngredientMatches, func(a, b int) bool {
		return len(recipeIngredientMatches[a].acceptedItems) < len(recipeIngredientMatches[b].acceptedItems)
	})

outer:
	for _, match := range recipeIngredientMatches {
		needed := expectedIterations
		for _, itemIndex := range packedIndexes {
			it, ok := packedProvidedItems[itemIndex]
			if !ok || !match.acceptedItems[itemIndex] {
				continue
			}
			taken := min(needed, it.GetCount())
			needed -= taken
			it.SetCount(it.GetCount() - taken)
			if it.GetCount() == 0 {
				delete(packedProvidedItems, itemIndex)
			}
			if needed == 0 {
				//validation passed!
				continue outer
			}
		}
		actualIterations := expectedIterations - needed
		return validationError(fmt.Sprintf("Not enough items to satisfy recipe ingredient %s for %d (only have enough items for %d iterations)", recipeIngredients[match.ingredientIndex], expectedIterations, actualIterations))
	}

	if len(packedProvidedItems) > 0 {
		return validationError("Not all provided items were used")
	}
	return nil
}

// matchOutputs is a port of CraftingTransaction::matchOutputs: the number of times the recipe
// was crafted.
func (t *CraftingTransaction) matchOutputs(txItems, recipeItems []item.Item) (int, error) {
	if len(recipeItems) == 0 {
		return 0, validationError("No recipe items given")
	}
	if len(txItems) == 0 {
		return 0, validationError("No transaction items given")
	}
	txItems = append([]item.Item(nil), txItems...)
	recipeItems = append([]item.Item(nil), recipeItems...)

	iterations := 0
	for len(recipeItems) > 0 {
		recipeItem := recipeItems[len(recipeItems)-1]
		recipeItems = recipeItems[:len(recipeItems)-1]
		needCount := recipeItem.GetCount()
		remaining := recipeItems[:0]
		for _, other := range recipeItems {
			if other.CanStackWith(recipeItem) { //make sure they have the same wildcards set
				needCount += other.GetCount()
			} else {
				remaining = append(remaining, other)
			}
		}
		recipeItems = remaining

		haveCount := 0
		remainingTx := txItems[:0]
		for _, txItem := range txItems {
			if txItem.CanStackWith(recipeItem) {
				haveCount += txItem.GetCount()
			} else {
				remainingTx = append(remainingTx, txItem)
			}
		}
		txItems = remainingTx

		if haveCount%needCount != 0 {
			//wrong count for this output, should divide exactly
			return 0, validationError(fmt.Sprintf("Expected an exact multiple of required %v (given: %d, needed: %d)", recipeItem, haveCount, needCount))
		}

		multiplier := haveCount / needCount
		if multiplier < 1 {
			return 0, validationError(fmt.Sprintf("Expected more than zero items matching %v (given: %d, needed: %d)", recipeItem, haveCount, needCount))
		}
		if iterations == 0 {
			iterations = multiplier
		} else if multiplier != iterations {
			//wrong count for this output, should match previous outputs
			return 0, validationError(fmt.Sprintf("Expected %v x%d, but found x%d", recipeItem, iterations, multiplier))
		}
	}

	if len(txItems) > 0 {
		//all items should be destroyed in this process
		return 0, validationError(fmt.Sprintf("Expected 0 items left over, have %d", len(txItems)))
	}
	return iterations, nil
}

// craftingGrid is $this->source->getCraftingGrid() as a CraftingGrid.
func (t *CraftingTransaction) craftingGrid() *crafting.CraftingGrid {
	if owner, ok := t.source.(craftingGridOwner); ok {
		if g, ok := owner.GetCraftingGrid().(interface{ Grid() *crafting.CraftingGrid }); ok {
			return g.Grid()
		}
	}
	return nil
}

// validateRecipe is a port of CraftingTransaction::validateRecipe.
func (t *CraftingTransaction) validateRecipe(recipe crafting.CraftingRecipe, expectedRepetitions *int) (int, error) {
	//compute number of times recipe was crafted
	repetitions, err := t.matchOutputs(t.outputs, recipe.GetResultsFor(t.craftingGrid()))
	if err != nil {
		return 0, err
	}
	if expectedRepetitions != nil && repetitions != *expectedRepetitions {
		return 0, validationError(fmt.Sprintf("Expected %d repetitions, got %d", *expectedRepetitions, repetitions))
	}
	//assert that $repetitions x recipe ingredients should be consumed
	if err := MatchIngredients(t.inputs, recipe.GetIngredientList(), repetitions); err != nil {
		return 0, err
	}
	return repetitions, nil
}

// Validate is a port of CraftingTransaction::validate.
func (t *CraftingTransaction) Validate() error {
	if err := t.squashDuplicateSlotChanges(); err != nil {
		return err
	}
	if len(t.actions) < 1 {
		return validationError("Transaction must have at least one action to be executable")
	}

	outputs, inputs, err := t.MatchItems()
	if err != nil {
		return err
	}
	t.outputs, t.inputs = outputs, inputs

	if t.recipe == nil {
		failed := 0
		for recipe := range t.craftingManager.MatchRecipeByOutputs(t.outputs) {
			//compute number of times recipe was crafted
			repetitions, err := t.matchOutputs(t.outputs, recipe.GetResultsFor(t.craftingGrid()))
			if err == nil {
				//assert that $repetitions x recipe ingredients should be consumed
				err = MatchIngredients(t.inputs, recipe.GetIngredientList(), repetitions)
			}
			if err != nil {
				failed++
				continue
			}
			//Success!
			t.recipe = recipe
			t.repetitions = &repetitions
			break
		}
		if t.recipe == nil {
			return validationError(fmt.Sprintf("Unable to match a recipe to transaction (tried to match against %d recipes)", failed))
		}
		return nil
	}

	repetitions, err := t.validateRecipe(t.recipe, t.repetitions)
	if err != nil {
		return err
	}
	t.repetitions = &repetitions
	return nil
}

func toEventItems(items []item.Item) []inventoryevent.Item {
	result := make([]inventoryevent.Item, len(items))
	for i, it := range items {
		result[i] = it
	}
	return result
}

// CallExecuteEvent is a port of CraftingTransaction::callExecuteEvent.
func (t *CraftingTransaction) CallExecuteEvent() bool {
	repetitions := 0
	if t.repetitions != nil {
		repetitions = *t.repetitions
	}
	ev := inventoryevent.NewCraftItemEvent(t, t.recipe, repetitions, toEventItems(t.inputs), toEventItems(t.outputs))
	event.Call(ev)
	return !ev.IsCancelled()
}
