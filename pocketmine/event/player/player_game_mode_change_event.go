package player

import "pocketmine-go/pocketmine/event"

// GameMode is the surface PlayerGameModeChangeEvent needs from pocketmine\player\GameMode
// (player.GameMode satisfies it).
type GameMode interface {
	GetEnglishName() string
}

// PlayerGameModeChangeEvent is a port of pocketmine\event\player\PlayerGameModeChangeEvent:
// called when a player has its gamemode changed.
type PlayerGameModeChangeEvent struct {
	PlayerEvent
	event.CancellableTrait

	newGamemode GameMode
}

func NewPlayerGameModeChangeEvent(player Player, newGamemode GameMode) *PlayerGameModeChangeEvent {
	return &PlayerGameModeChangeEvent{PlayerEvent: PlayerEvent{player: player}, newGamemode: newGamemode}
}

func (e *PlayerGameModeChangeEvent) GetNewGamemode() GameMode { return e.newGamemode }
