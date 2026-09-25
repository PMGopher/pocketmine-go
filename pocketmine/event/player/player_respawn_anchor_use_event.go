package player

import "pocketmine-go/pocketmine/event"

// PlayerRespawnAnchorUseEvent actions, PlayerRespawnAnchorUseEvent::ACTION_*.
const (
	RespawnAnchorActionExplode  = 0
	RespawnAnchorActionSetSpawn = 1
)

// PlayerRespawnAnchorUseEvent is a port of pocketmine\event\player\PlayerRespawnAnchorUseEvent.
type PlayerRespawnAnchorUseEvent struct {
	PlayerEvent
	event.CancellableTrait

	block  Block
	action int
}

func NewPlayerRespawnAnchorUseEvent(player Player, block Block, action int) *PlayerRespawnAnchorUseEvent {
	return &PlayerRespawnAnchorUseEvent{PlayerEvent: PlayerEvent{player: player}, block: block, action: action}
}

func (e *PlayerRespawnAnchorUseEvent) GetBlock() Block { return e.block }

func (e *PlayerRespawnAnchorUseEvent) GetAction() int { return e.action }

func (e *PlayerRespawnAnchorUseEvent) SetAction(action int) { e.action = action }
