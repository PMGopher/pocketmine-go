package player

import (
	"fmt"

	"pocketmine-go/pocketmine/event"
)

// PlayerEnchantingOptionsRequestEvent is a port of
// pocketmine\event\player\PlayerEnchantingOptionsRequestEvent: called when a player inserts an
// item into an enchanting table's input slot. The options provided by the event will be shown on
// the enchanting table menu. The inventory is a block/inventory.EnchantInventory and options are
// item/enchantment.EnchantingOption values.
type PlayerEnchantingOptionsRequestEvent struct {
	PlayerEvent
	event.CancellableTrait

	inventory any
	options   []any
}

func NewPlayerEnchantingOptionsRequestEvent(player Player, inventory any, options []any) *PlayerEnchantingOptionsRequestEvent {
	return &PlayerEnchantingOptionsRequestEvent{PlayerEvent: PlayerEvent{player: player}, inventory: inventory, options: options}
}

func (e *PlayerEnchantingOptionsRequestEvent) GetInventory() any { return e.inventory }

func (e *PlayerEnchantingOptionsRequestEvent) GetOptions() []any { return e.options }

func (e *PlayerEnchantingOptionsRequestEvent) SetOptions(options []any) {
	if optionCount := len(options); optionCount > 3 {
		panic(fmt.Sprintf("The maximum number of options for an enchanting table is 3, but %d have been passed", optionCount))
	}
	e.options = options
}
