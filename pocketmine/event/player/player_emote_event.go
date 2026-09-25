package player

import "pocketmine-go/pocketmine/event"

// PlayerEmoteEvent is a port of pocketmine\event\player\PlayerEmoteEvent: called when a player
// uses an emote.
type PlayerEmoteEvent struct {
	PlayerEvent
	event.CancellableTrait

	emoteID string
}

func NewPlayerEmoteEvent(player Player, emoteID string) *PlayerEmoteEvent {
	return &PlayerEmoteEvent{PlayerEvent: PlayerEvent{player: player}, emoteID: emoteID}
}

func (e *PlayerEmoteEvent) GetEmoteId() string { return e.emoteID }

func (e *PlayerEmoteEvent) SetEmoteId(emoteID string) { e.emoteID = emoteID }
