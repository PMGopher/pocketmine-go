package inventory

import (
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
)

// PlayerEnderInventory is a port of pocketmine\inventory\PlayerEnderInventory. The holder is a
// Human.
type PlayerEnderInventory struct {
	SimpleInventory

	holder entityevent.Entity
}

// NewPlayerEnderInventory is a port of PlayerEnderInventory::__construct (PHP's default size is 27).
func NewPlayerEnderInventory(holder entityevent.Entity, size int) *PlayerEnderInventory {
	p := &PlayerEnderInventory{SimpleInventory: SimpleInventory{slots: make([]item.Item, size)}, holder: holder}
	p.Init(p)
	return p
}

func (p *PlayerEnderInventory) GetHolder() entityevent.Entity { return p.holder }
