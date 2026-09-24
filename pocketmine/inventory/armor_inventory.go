package inventory

import (
	"pocketmine-go/pocketmine/block"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
)

// ArmorInventory slot constants, a port of ArmorInventory::SLOT_*.
const (
	ArmorSlotHead  = 0
	ArmorSlotChest = 1
	ArmorSlotLegs  = 2
	ArmorSlotFeet  = 3
)

// ArmorInventory is a port of pocketmine\inventory\ArmorInventory. The holder is a Living (the
// local entityevent.Entity interface here - pocketmine/entity imports this package).
type ArmorInventory struct {
	SimpleInventory

	holder entityevent.Entity
}

// NewArmorInventory is a port of ArmorInventory::__construct.
func NewArmorInventory(holder entityevent.Entity) *ArmorInventory {
	a := &ArmorInventory{SimpleInventory: SimpleInventory{slots: make([]item.Item, 4)}, holder: holder}
	a.Init(a)
	a.GetSlotValidators().Add(NewCallbackSlotValidator(validateArmorSlot))
	return a
}

func (a *ArmorInventory) GetHolder() entityevent.Entity { return a.holder }

func (a *ArmorInventory) GetHelmet() item.Item { return a.GetItem(ArmorSlotHead) }

func (a *ArmorInventory) GetChestplate() item.Item { return a.GetItem(ArmorSlotChest) }

func (a *ArmorInventory) GetLeggings() item.Item { return a.GetItem(ArmorSlotLegs) }

func (a *ArmorInventory) GetBoots() item.Item { return a.GetItem(ArmorSlotFeet) }

func (a *ArmorInventory) SetHelmet(helmet item.Item) { a.SetItem(ArmorSlotHead, helmet) }

func (a *ArmorInventory) SetChestplate(chestplate item.Item) { a.SetItem(ArmorSlotChest, chestplate) }

func (a *ArmorInventory) SetLeggings(leggings item.Item) { a.SetItem(ArmorSlotLegs, leggings) }

func (a *ArmorInventory) SetBoots(boots item.Item) { a.SetItem(ArmorSlotFeet, boots) }

// armorSlotted is the surface validateArmorSlot needs from item.Armor.
type armorSlotted interface {
	GetArmorSlot() int
}

// validateArmorSlot is a port of ArmorInventory::validate: armor must go in its own slot; the only
// non-armor items accepted are carved pumpkins and mob heads, in the head slot.
func validateArmorSlot(_ Inventory, it item.Item, slot int) *TransactionValidationError {
	if armor, ok := it.(armorSlotted); ok {
		if armor.GetArmorSlot() != slot {
			return &TransactionValidationError{Message: "Armor item is in wrong slot"}
		}
		return nil
	}
	if ib, ok := it.(*item.ItemBlock); ok && slot == ArmorSlotHead {
		if t := ib.GetBlock().GetTypeId(); t == block.CARVED_PUMPKIN || t == block.MOB_HEAD {
			return nil
		}
	}
	return &TransactionValidationError{Message: "Item is not accepted in an armor slot"}
}
