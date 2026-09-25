// Package player is a port of pocketmine\event\player: events fired by (or about) players.
//
// PlayerExhaustEvent and PlayerExperienceChangeEvent extend EntityEvent in PHP (they're fired for
// any Human, not just players), so they embed entityevent.EntityEvent here too; every other event
// embeds PlayerEvent.
//
// This package sits below pocketmine/entity and pocketmine/player in the import graph (both fire
// its events), so the PHP-typed payloads are small local interfaces that *player.Player,
// *entity.Skin, block.Behavior, item.Item and so on satisfy. Listeners type-assert to the concrete
// type they need, exactly where PHP code would narrow with instanceof. Translatable|string values
// are `any` holding a string or *lang.Translatable, as everywhere else in this port.
//
// Importers conventionally alias this package as playerevent.
package player

import (
	blockevent "pocketmine-go/pocketmine/event/block"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// Player is the surface these events need from pocketmine\player\Player.
type Player interface {
	entityevent.Entity
	GetName() string
	GetDisplayName() string
}

// Block is the surface these events need from pocketmine\block\Block.
type Block = blockevent.Block

// Item is the surface these events need from pocketmine\item\Item.
type Item = entityevent.Item

// Entity is the surface these events need from pocketmine\entity\Entity.
type Entity = entityevent.Entity

// Position is pocketmine\world\Position.
type Position = entityevent.Position

// NetworkSession is the surface these events need from pocketmine\network\mcpe\NetworkSession.
type NetworkSession interface {
	GetIp() string
	GetPort() int
}

// PlayerInfo is the surface these events need from pocketmine\player\PlayerInfo.
type PlayerInfo interface {
	GetUsername() string
	GetUUID() string
	GetLocale() string
}

// PlayerEvent is a port of pocketmine\event\player\PlayerEvent.
type PlayerEvent struct {
	player Player
}

// NewPlayerEventBase builds the embedded PlayerEvent (PHP's `$this->player = $player`).
func NewPlayerEventBase(player Player) PlayerEvent { return PlayerEvent{player: player} }

// GetPlayer is a port of PlayerEvent::getPlayer.
func (e *PlayerEvent) GetPlayer() Player { return e.player }

// cloneItems is Utils::cloneObjectArray for items.
func cloneItems(items []Item) []Item {
	result := make([]Item, len(items))
	for i, it := range items {
		result[i] = entityevent.CloneItem(it)
	}
	return result
}
