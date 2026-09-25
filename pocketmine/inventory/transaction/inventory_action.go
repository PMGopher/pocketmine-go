package transaction

import (
	"fmt"

	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/utils"
)

// Player is what transactions and their actions need from pocketmine\player\Player.
type Player interface {
	playerevent.Player
	HasFiniteResources() bool
	GetCreativeInventory() *inventory.CreativeInventory
	IsSpectator() bool
	DropItem(it item.Item)
}

// InventoryAction is a port of pocketmine\inventory\transaction\action\InventoryAction:
// represents an action involving a change that applies in some way to an inventory or other item
// source.
type InventoryAction interface {
	// GetSourceItem returns the item that was present before the action took place.
	GetSourceItem() item.Item
	// GetTargetItem returns the item that the action attempted to replace the source item with.
	GetTargetItem() item.Item
	// Validate returns an error if the action is not valid in the current state of the server.
	Validate(source Player) error
	// OnAddToTransaction is called when the action is added to the specified transaction.
	OnAddToTransaction(transaction *InventoryTransaction)
	// OnPreExecute is called by inventory transactions before any actions are processed. If this
	// returns false, the transaction will be cancelled.
	OnPreExecute(source Player) bool
	// Execute performs actions needed to complete the inventory-action server-side. This will
	// only be called if the transaction which it is part of is considered valid.
	Execute(source Player)
}

// InventoryActionBase holds the source and target items of an action.
type InventoryActionBase struct {
	sourceItem, targetItem item.Item
}

func newActionBase(sourceItem, targetItem item.Item) InventoryActionBase {
	return InventoryActionBase{sourceItem: sourceItem, targetItem: targetItem}
}

func (a *InventoryActionBase) GetSourceItem() item.Item { return a.sourceItem.Clone() }

func (a *InventoryActionBase) GetTargetItem() item.Item { return a.targetItem.Clone() }

func (a *InventoryActionBase) OnAddToTransaction(transaction *InventoryTransaction) {}

func (a *InventoryActionBase) OnPreExecute(source Player) bool { return true }

// SlotChangeAction is a port of SlotChangeAction: represents an action causing a change in an
// inventory slot.
type SlotChangeAction struct {
	InventoryActionBase

	inventory     inventory.Inventory
	inventorySlot int
}

func NewSlotChangeAction(inv inventory.Inventory, inventorySlot int, sourceItem, targetItem item.Item) *SlotChangeAction {
	return &SlotChangeAction{InventoryActionBase: newActionBase(sourceItem, targetItem), inventory: inv, inventorySlot: inventorySlot}
}

// GetInventory returns the inventory involved in this action.
func (a *SlotChangeAction) GetInventory() inventory.Inventory { return a.inventory }

// GetSlot returns the slot in the inventory which this action modified.
func (a *SlotChangeAction) GetSlot() int { return a.inventorySlot }

// slotValidated is SlotValidatedInventory.
type slotValidated interface {
	GetSlotValidators() *utils.ObjectSet[inventory.SlotValidator]
}

// Validate verifies that the slot exists, contains the expected original item, and that the
// target item fits.
func (a *SlotChangeAction) Validate(source Player) error {
	if !a.inventory.SlotExists(a.inventorySlot) {
		return validationError("Slot does not exist")
	}
	if !a.inventory.GetItem(a.inventorySlot).EqualsExact(a.sourceItem) {
		return validationError("Slot does not contain expected original item")
	}
	if a.targetItem.GetCount() > a.targetItem.GetMaxStackSize() {
		return validationError("Target item exceeds item type max stack size")
	}
	if a.targetItem.GetCount() > a.inventory.GetMaxStackSize() {
		return validationError("Target item exceeds inventory max stack size")
	}
	if v, ok := a.inventory.(slotValidated); ok && !a.targetItem.IsNull() {
		for validator := range v.GetSlotValidators().All() {
			if ret := validator.Validate(a.inventory, a.targetItem, a.inventorySlot); ret != nil {
				return validationError(fmt.Sprintf("Target item is not accepted by the inventory at slot #%d: %s", a.inventorySlot, ret.Message))
			}
		}
	}
	return nil
}

// Execute sets the item into the target inventory.
func (a *SlotChangeAction) Execute(source Player) {
	a.inventory.SetItem(a.inventorySlot, a.targetItem)
}

// CreateItemAction is a port of CreateItemAction: this action is used by creative players to
// balance transactions involving the creative inventory menu. The source item is the item being
// created ("taken" from the creative menu).
type CreateItemAction struct{ InventoryActionBase }

func NewCreateItemAction(sourceItem item.Item) *CreateItemAction {
	return &CreateItemAction{InventoryActionBase: newActionBase(sourceItem, inventory.Air())}
}

func (a *CreateItemAction) Validate(source Player) error {
	if source.HasFiniteResources() {
		return validationError("Player has finite resources, cannot create items")
	}
	if !source.GetCreativeInventory().Contains(a.sourceItem) {
		return validationError("Creative inventory does not contain requested item")
	}
	return nil
}

func (a *CreateItemAction) Execute(source Player) {
	//NOOP
}

// DestroyItemAction is a port of DestroyItemAction: this action type shows up when a creative
// player puts an item into the creative inventory menu to destroy it. The output is the item
// destroyed. You can think of this action type like setting an item into /dev/null.
type DestroyItemAction struct{ InventoryActionBase }

func NewDestroyItemAction(targetItem item.Item) *DestroyItemAction {
	return &DestroyItemAction{InventoryActionBase: newActionBase(inventory.Air(), targetItem)}
}

func (a *DestroyItemAction) Validate(source Player) error {
	if source.HasFiniteResources() {
		return validationError("Player has finite resources, cannot destroy items")
	}
	return nil
}

func (a *DestroyItemAction) Execute(source Player) {
	//NOOP
}

// DropItemAction is a port of DropItemAction: represents an action involving dropping an item
// into the world.
type DropItemAction struct{ InventoryActionBase }

func NewDropItemAction(targetItem item.Item) *DropItemAction {
	return &DropItemAction{InventoryActionBase: newActionBase(inventory.Air(), targetItem)}
}

func (a *DropItemAction) Validate(source Player) error {
	if a.targetItem.IsNull() {
		return validationError("Cannot drop an empty itemstack")
	}
	if a.targetItem.GetCount() > a.targetItem.GetMaxStackSize() {
		return validationError("Target item exceeds item type max stack size")
	}
	return nil
}

func (a *DropItemAction) OnPreExecute(source Player) bool {
	ev := playerevent.NewPlayerDropItemEvent(source, a.targetItem)
	if source.IsSpectator() {
		ev.Cancel()
	}
	event.Call(ev)
	return !ev.IsCancelled()
}

// Execute drops the target item in front of the player.
func (a *DropItemAction) Execute(source Player) {
	source.DropItem(a.targetItem)
}
