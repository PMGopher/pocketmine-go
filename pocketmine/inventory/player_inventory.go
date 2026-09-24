package inventory

import (
	"fmt"

	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/utils"
)

// HeldItemIndexChangeListener is called with the previous held slot whenever the held item index
// changes (PHP's heldItemIndexChangeListeners closures).
type HeldItemIndexChangeListener func(oldIndex int)

// PlayerInventory is a port of pocketmine\inventory\PlayerInventory. The holder is a Human.
type PlayerInventory struct {
	SimpleInventory

	holder                       entityevent.Entity
	itemInHandIndex              int
	heldItemIndexChangeListeners *utils.ObjectSet[*HeldItemIndexChangeListener]
}

// NewPlayerInventory is a port of PlayerInventory::__construct (36 slots).
func NewPlayerInventory(player entityevent.Entity) *PlayerInventory {
	p := &PlayerInventory{
		SimpleInventory:              SimpleInventory{slots: make([]item.Item, 36)},
		holder:                       player,
		heldItemIndexChangeListeners: utils.NewObjectSet[*HeldItemIndexChangeListener](),
	}
	p.Init(p)
	return p
}

func (p *PlayerInventory) IsHotbarSlot(slot int) bool {
	return slot >= 0 && slot < p.GetHotbarSize()
}

func (p *PlayerInventory) throwIfNotHotbarSlot(slot int) {
	if !p.IsHotbarSlot(slot) {
		panic(fmt.Sprintf("%d is not a valid hotbar slot index (expected 0 - %d)", slot, p.GetHotbarSize()-1))
	}
}

// GetHotbarSlotItem returns the item in the specified hotbar slot. Panics for a non-hotbar slot.
func (p *PlayerInventory) GetHotbarSlotItem(hotbarSlot int) item.Item {
	p.throwIfNotHotbarSlot(hotbarSlot)
	return p.GetItem(hotbarSlot)
}

// GetHeldItemIndex returns the hotbar slot number the holder is currently holding.
func (p *PlayerInventory) GetHeldItemIndex() int { return p.itemInHandIndex }

// SetHeldItemIndex is a port of PlayerInventory::setHeldItemIndex. Panics for a non-hotbar slot.
func (p *PlayerInventory) SetHeldItemIndex(hotbarSlot int) {
	p.throwIfNotHotbarSlot(hotbarSlot)

	oldIndex := p.itemInHandIndex
	p.itemInHandIndex = hotbarSlot

	for callback := range p.heldItemIndexChangeListeners.All() {
		(*callback)(oldIndex)
	}
}

func (p *PlayerInventory) GetHeldItemIndexChangeListeners() *utils.ObjectSet[*HeldItemIndexChangeListener] {
	return p.heldItemIndexChangeListeners
}

// GetItemInHand returns the currently-held item.
func (p *PlayerInventory) GetItemInHand() item.Item {
	return p.GetHotbarSlotItem(p.itemInHandIndex)
}

// SetItemInHand sets the item in the currently-held slot to the specified item.
func (p *PlayerInventory) SetItemInHand(it item.Item) {
	p.SetItem(p.GetHeldItemIndex(), it)
}

// GetHotbarSize returns the number of slots in the hotbar.
func (p *PlayerInventory) GetHotbarSize() int { return 9 }

func (p *PlayerInventory) GetHolder() entityevent.Entity { return p.holder }
