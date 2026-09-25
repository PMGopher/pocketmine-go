package block

import (
	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// BlockBreakEvent is a port of pocketmine\event\block\BlockBreakEvent: called when a player
// destroys a block somewhere in the world.
type BlockBreakEvent struct {
	BlockEvent
	event.CancellableTrait

	player     Player
	item       Item
	instaBreak bool
	blockDrops []Item
	xpDrops    int
}

func NewBlockBreakEvent(player Player, block Block, item Item, instaBreak bool, drops []Item, xpDrops int) *BlockBreakEvent {
	return &BlockBreakEvent{BlockEvent: BlockEvent{block: block}, player: player, item: item, instaBreak: instaBreak, blockDrops: drops, xpDrops: xpDrops}
}

// GetPlayer returns the player who is destroying the block.
func (e *BlockBreakEvent) GetPlayer() Player { return e.player }

// GetItem returns (a clone of) the item used to destroy the block.
func (e *BlockBreakEvent) GetItem() Item { return entityevent.CloneItem(e.item) }

// GetInstaBreak returns whether the block may be broken in less than the amount of time
// calculated. This is usually true for creative players.
func (e *BlockBreakEvent) GetInstaBreak() bool { return e.instaBreak }

func (e *BlockBreakEvent) SetInstaBreak(instaBreak bool) { e.instaBreak = instaBreak }

func (e *BlockBreakEvent) GetDrops() []Item { return e.blockDrops }

func (e *BlockBreakEvent) SetDrops(drops []Item) { e.blockDrops = drops }

// GetXpDropAmount returns how much XP will be dropped by breaking this block.
func (e *BlockBreakEvent) GetXpDropAmount() int { return e.xpDrops }

// SetXpDropAmount sets how much XP will be dropped by breaking this block.
func (e *BlockBreakEvent) SetXpDropAmount(amount int) {
	if amount < 0 {
		panic("Amount must be at least zero")
	}
	e.xpDrops = amount
}
