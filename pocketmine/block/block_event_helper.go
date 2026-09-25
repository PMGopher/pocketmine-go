package block

import (
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
)

// This file is a port of pocketmine\block\utils\BlockEventHelper: each helper fires the matching
// pocketmine\event\block event and, unless it's cancelled, sets the event's new state in the
// world. Unlike the PHP original, the events are always constructed and fired rather than
// checking hasHandlers() first - a pure performance optimization this port doesn't need to
// replicate.

// setNewState writes the (possibly handler-changed) new state of a block change event.
func setNewState(oldState Behavior, newState blockevent.Block) bool {
	world, err := oldState.GetPosition().GetWorld()
	if err != nil {
		return false
	}
	return world.SetBlock(oldState.GetPosition(), newState.(Behavior)) == nil
}

// Melt is a port of BlockEventHelper::melt.
func Melt(oldState Behavior, newState Behavior) bool {
	ev := blockevent.NewBlockMeltEvent(oldState, newState)
	event.Call(ev)
	return !ev.IsCancelled() && setNewState(oldState, ev.GetNewState())
}

// Die is a port of BlockEventHelper::die.
func Die(oldState Behavior, newState Behavior) bool {
	ev := blockevent.NewBlockDeathEvent(oldState, newState)
	event.Call(ev)
	return !ev.IsCancelled() && setNewState(oldState, ev.GetNewState())
}

// Spread is a port of BlockEventHelper::spread.
func Spread(oldState Behavior, newState Behavior, source Behavior) bool {
	ev := blockevent.NewBlockSpreadEvent(oldState, source, newState)
	event.Call(ev)
	return !ev.IsCancelled() && setNewState(oldState, ev.GetNewState())
}

// Grow is a port of BlockEventHelper::grow. causingPlayer may be nil.
func Grow(oldState Behavior, newState Behavior, causingPlayer Player) bool {
	var player blockevent.Player
	if causingPlayer != nil {
		player = causingPlayer
	}
	ev := blockevent.NewBlockGrowEvent(oldState, newState, player)
	event.Call(ev)
	return !ev.IsCancelled() && setNewState(oldState, ev.GetNewState())
}

// Form is a port of BlockEventHelper::form.
func Form(oldState Behavior, newState Behavior, causingBlock Behavior) bool {
	ev := blockevent.NewBlockFormEvent(oldState, newState, causingBlock)
	event.Call(ev)
	return !ev.IsCancelled() && setNewState(oldState, ev.GetNewState())
}
