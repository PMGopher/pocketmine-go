package player

import (
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/nbt"
)

// PlayerDataSaveEvent is a port of pocketmine\event\player\PlayerDataSaveEvent: called when a
// player's data is about to be saved to disk.
type PlayerDataSaveEvent struct {
	event.CancellableTrait

	data       *nbt.CompoundTag
	playerName string
	player     Player
}

// NewPlayerDataSaveEvent creates the event; player is nil when the player isn't online.
func NewPlayerDataSaveEvent(data *nbt.CompoundTag, playerName string, player Player) *PlayerDataSaveEvent {
	return &PlayerDataSaveEvent{data: data, playerName: playerName, player: player}
}

// GetSaveData returns the data to be written to disk as a CompoundTag.
func (e *PlayerDataSaveEvent) GetSaveData() *nbt.CompoundTag { return e.data }

func (e *PlayerDataSaveEvent) SetSaveData(data *nbt.CompoundTag) { e.data = data }

// GetPlayerName returns the username of the player whose data is being saved. This is not
// necessarily an online player.
func (e *PlayerDataSaveEvent) GetPlayerName() string { return e.playerName }

// GetPlayer returns the player whose data is being saved, if online, or nil.
func (e *PlayerDataSaveEvent) GetPlayer() Player { return e.player }
