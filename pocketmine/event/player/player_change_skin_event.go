package player

import "pocketmine-go/pocketmine/event"

// PlayerChangeSkinEvent is a port of pocketmine\event\player\PlayerChangeSkinEvent: called when a
// player changes their skin in-game. Skins are *entity.Skin.
type PlayerChangeSkinEvent struct {
	PlayerEvent
	event.CancellableTrait

	oldSkin, newSkin any
}

func NewPlayerChangeSkinEvent(player Player, oldSkin, newSkin any) *PlayerChangeSkinEvent {
	return &PlayerChangeSkinEvent{PlayerEvent: PlayerEvent{player: player}, oldSkin: oldSkin, newSkin: newSkin}
}

func (e *PlayerChangeSkinEvent) GetOldSkin() any { return e.oldSkin }

func (e *PlayerChangeSkinEvent) GetNewSkin() any { return e.newSkin }

func (e *PlayerChangeSkinEvent) SetNewSkin(skin any) { e.newSkin = skin }
