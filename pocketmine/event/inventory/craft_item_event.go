package inventory

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// CraftingTransaction is the surface CraftItemEvent needs from
// pocketmine\inventory\transaction\CraftingTransaction.
type CraftingTransaction interface {
	GetSourcePlayer() Player
}

// CraftItemEvent is a port of pocketmine\event\inventory\CraftItemEvent. The recipe is a
// crafting.CraftingRecipe.
type CraftItemEvent struct {
	event.CancellableTrait

	transaction CraftingTransaction
	recipe      any
	repetitions int
	inputs      []Item
	outputs     []Item
}

func NewCraftItemEvent(transaction CraftingTransaction, recipe any, repetitions int, inputs, outputs []Item) *CraftItemEvent {
	return &CraftItemEvent{transaction: transaction, recipe: recipe, repetitions: repetitions, inputs: inputs, outputs: outputs}
}

// GetTransaction returns the inventory transaction involved in this crafting event.
func (e *CraftItemEvent) GetTransaction() CraftingTransaction { return e.transaction }

// GetRecipe returns the recipe crafted.
func (e *CraftItemEvent) GetRecipe() any { return e.recipe }

// GetRepetitions returns the number of times the recipe was crafted. This is usually 1, but
// might be more in the case of recipe book shift-clicks (which craft lots of items in a batch).
func (e *CraftItemEvent) GetRepetitions() int { return e.repetitions }

// GetInputs returns a list of items destroyed as ingredients of the recipe.
func (e *CraftItemEvent) GetInputs() []Item { return cloneItems(e.inputs) }

// GetOutputs returns a list of items created by crafting the recipe.
func (e *CraftItemEvent) GetOutputs() []Item { return cloneItems(e.outputs) }

func (e *CraftItemEvent) GetPlayer() Player { return e.transaction.GetSourcePlayer() }

func cloneItems(items []Item) []Item {
	result := make([]Item, len(items))
	for i, it := range items {
		result[i] = entityevent.CloneItem(it)
	}
	return result
}
