package block

import "pocketmine-go/pocketmine/event"

// BlockChangeEvent is a port of pocketmine\event\block\BaseBlockChangeEvent (itself extending
// BlockEvent) - it also stands in directly for BlockMeltEvent (Melt below) and BlockDeathEvent
// (Die below), since neither PHP class adds anything to the base shape. BlockFormEvent/
// BlockGrowEvent (below) are real, separate subclasses since they each add one field; Spread/Fade
// aren't ported since nothing calls them yet - same "port what's actually needed" reasoning as
// tile.FurnaceType living directly in the tile package.
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

// Die is a port of pocketmine\block\utils\BlockEventHelper::die. It reuses BlockChangeEvent
// directly as BlockDeathEvent's stand-in (see the doc comment above) and is otherwise identical to
// Melt.
func Die(oldState Behavior, newState Behavior) bool {
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

// BlockSpreadEvent is a port of pocketmine\event\block\BlockSpreadEvent.
type BlockSpreadEvent struct {
	BlockChangeEvent

	Source Behavior
}

func (e *BlockSpreadEvent) GetSource() Behavior { return e.Source }

// Spread is a port of pocketmine\block\utils\BlockEventHelper::spread.
func Spread(oldState Behavior, newState Behavior, source Behavior) bool {
	ev := &BlockSpreadEvent{BlockChangeEvent: BlockChangeEvent{Block: oldState, NewState: newState}, Source: source}
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

// BlockGrowEvent is a port of pocketmine\event\block\BlockGrowEvent.
type BlockGrowEvent struct {
	BlockChangeEvent

	Player Player // nil when the block grows by itself
}

func (e *BlockGrowEvent) GetPlayer() Player { return e.Player }

// Grow is a port of pocketmine\block\utils\BlockEventHelper::grow.
func Grow(oldState Behavior, newState Behavior, causingPlayer Player) bool {
	ev := &BlockGrowEvent{BlockChangeEvent: BlockChangeEvent{Block: oldState, NewState: newState}, Player: causingPlayer}
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
