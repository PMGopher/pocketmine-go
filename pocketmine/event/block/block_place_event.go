package block

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// BlockPlaceEvent is a port of pocketmine\event\block\BlockPlaceEvent: called when a player
// initiates a block placement action. More than one block may be changed by a single placement
// action, for example when placing a door.
//
// PHP's constructor positions every block of the transaction in the world; here the caller
// (World.UseItemOn) does that before creating the event, since this package can't reach blocks.
type BlockPlaceEvent struct {
	event.CancellableTrait

	player       Player
	transaction  BlockTransaction
	blockAgainst Block
	item         Item
}

func NewBlockPlaceEvent(player Player, transaction BlockTransaction, blockAgainst Block, item Item) *BlockPlaceEvent {
	return &BlockPlaceEvent{player: player, transaction: transaction, blockAgainst: blockAgainst, item: item}
}

// GetPlayer returns the player who is placing the block.
func (e *BlockPlaceEvent) GetPlayer() Player { return e.player }

// GetItem returns (a clone of) the item in the player's hand.
func (e *BlockPlaceEvent) GetItem() Item { return entityevent.CloneItem(e.item) }

// GetTransaction returns a BlockTransaction object containing all the block positions that will
// be changed by this event, and the states they will be changed to.
func (e *BlockPlaceEvent) GetTransaction() BlockTransaction { return e.transaction }

func (e *BlockPlaceEvent) GetBlockAgainst() Block { return e.blockAgainst }
