package player

// PlayerPostChunkSendEvent is a port of pocketmine\event\player\PlayerPostChunkSendEvent: called
// after a player is sent a chunk as part of their view radius.
type PlayerPostChunkSendEvent struct {
	PlayerEvent

	chunkX, chunkZ int
}

func NewPlayerPostChunkSendEvent(player Player, chunkX, chunkZ int) *PlayerPostChunkSendEvent {
	return &PlayerPostChunkSendEvent{PlayerEvent: PlayerEvent{player: player}, chunkX: chunkX, chunkZ: chunkZ}
}

func (e *PlayerPostChunkSendEvent) GetChunkX() int { return e.chunkX }
func (e *PlayerPostChunkSendEvent) GetChunkZ() int { return e.chunkZ }
