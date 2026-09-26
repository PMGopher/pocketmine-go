package player

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item/enchantment"
)

// PlayerItemEnchantEvent is a port of pocketmine\event\player\PlayerItemEnchantEvent: called when
// a player enchants an item using an enchanting table. The transaction is an
// inventory/transaction.EnchantingTransaction.
type PlayerItemEnchantEvent struct {
	PlayerEvent
	event.CancellableTrait

	transaction any
	option      *enchantment.EnchantingOption
	inputItem   Item
	outputItem  Item
	cost        int
}

func NewPlayerItemEnchantEvent(player Player, transaction any, option *enchantment.EnchantingOption, inputItem, outputItem Item, cost int) *PlayerItemEnchantEvent {
	return &PlayerItemEnchantEvent{PlayerEvent: PlayerEvent{player: player}, transaction: transaction, option: option, inputItem: inputItem, outputItem: outputItem, cost: cost}
}

// GetTransaction returns the inventory transaction involved in this enchant event.
func (e *PlayerItemEnchantEvent) GetTransaction() any { return e.transaction }

// GetOption returns the enchantment option used.
func (e *PlayerItemEnchantEvent) GetOption() *enchantment.EnchantingOption { return e.option }

// GetInputItem returns the item to be enchanted.
func (e *PlayerItemEnchantEvent) GetInputItem() Item { return entityevent.CloneItem(e.inputItem) }

// GetOutputItem returns the enchanted item.
func (e *PlayerItemEnchantEvent) GetOutputItem() Item { return entityevent.CloneItem(e.outputItem) }

// GetCost returns the number of XP levels and lapis that will be subtracted after enchanting if
// the player is not in creative mode.
func (e *PlayerItemEnchantEvent) GetCost() int { return e.cost }
