package block

// BlockGrowEvent is a port of pocketmine\event\block\BlockGrowEvent: called when plants or
// crops grow.
type BlockGrowEvent struct {
	BaseBlockChangeEvent

	player Player
}

// NewBlockGrowEvent creates the event; player is nil when the block grows by itself.
func NewBlockGrowEvent(block, newState Block, player Player) *BlockGrowEvent {
	return &BlockGrowEvent{BaseBlockChangeEvent: NewBaseBlockChangeEventBase(block, newState), player: player}
}

// GetPlayer returns the player that caused the growth (e.g. with bone meal), or nil.
func (e *BlockGrowEvent) GetPlayer() Player { return e.player }
