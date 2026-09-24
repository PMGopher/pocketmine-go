package inventory

import (
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
)

// PlayerOffHandInventory is a port of pocketmine\inventory\PlayerOffHandInventory. The holder is a
// Human.
type PlayerOffHandInventory struct {
	SimpleInventory

	holder entityevent.Entity
}

func NewPlayerOffHandInventory(player entityevent.Entity) *PlayerOffHandInventory {
	p := &PlayerOffHandInventory{SimpleInventory: SimpleInventory{slots: make([]item.Item, 1)}, holder: player}
	p.Init(p)
	return p
}

func (p *PlayerOffHandInventory) GetHolder() entityevent.Entity { return p.holder }
