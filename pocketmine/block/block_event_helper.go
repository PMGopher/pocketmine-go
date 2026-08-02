package block

import "pocketmine-go/pocketmine/event"

// BlockChangeEvent is a port of pocketmine\event\block\BaseBlockChangeEvent (itself extending
// BlockEvent). Only Melt (below) needs a concrete subclass of this shape right now, so this
// stands in directly for BlockMeltEvent rather than porting a whole hierarchy of near-identical
// *ChangeEvent types (Form/Die/Spread/Fade/Grow) that nothing calls yet - same "port what's
// actually needed" reasoning as tile.FurnaceType living directly in the tile package.
type BlockChangeEvent struct {
	event.CancellableTrait

	Block    Behavior
	NewState Behavior
}

func (e *BlockChangeEvent) GetBlock() Behavior { return e.Block }

func (e *BlockChangeEvent) GetNewState() Behavior { return e.NewState }

// Melt is a port of pocketmine\block\utils\BlockEventHelper::melt. Unlike the PHP original, this
// always constructs and fires the event rather than checking hasHandlers() first - a pure
// performance optimization this port doesn't need to replicate.
func Melt(oldState Behavior, newState Behavior) bool {
	ev := &BlockChangeEvent{Block: oldState, NewState: newState}
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}
	world, err := oldState.GetPosition().GetWorld()
	if err != nil {
		return false
	}
	if err := world.SetBlock(oldState.GetPosition(), ev.NewState); err != nil {
		return false
	}
	return true
}

// BlockFormEvent is a port of pocketmine\event\block\BlockFormEvent.
type BlockFormEvent struct {
	BlockChangeEvent

	CausingBlock Behavior
}

func (e *BlockFormEvent) GetCausingBlock() Behavior { return e.CausingBlock }

// Form is a port of pocketmine\block\utils\BlockEventHelper::form.
func Form(oldState Behavior, newState Behavior, causingBlock Behavior) bool {
	ev := &BlockFormEvent{BlockChangeEvent: BlockChangeEvent{Block: oldState, NewState: newState}, CausingBlock: causingBlock}
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}
	world, err := oldState.GetPosition().GetWorld()
	if err != nil {
		return false
	}
	if err := world.SetBlock(oldState.GetPosition(), ev.NewState); err != nil {
		return false
	}
	return true
}
