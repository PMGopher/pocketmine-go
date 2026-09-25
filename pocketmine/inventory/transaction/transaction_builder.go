package transaction

import (
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

// TransactionBuilder is a port of pocketmine\inventory\transaction\TransactionBuilder: collects
// changes made to proxy inventories (TransactionBuilderInventory) and turns them into actions.
type TransactionBuilder struct {
	inventories  []*TransactionBuilderInventory
	extraActions []InventoryAction
}

func NewTransactionBuilder() *TransactionBuilder { return &TransactionBuilder{} }

func (b *TransactionBuilder) AddAction(action InventoryAction) {
	b.extraActions = append(b.extraActions, action)
}

// GetInventory returns the proxy for inv, creating it on first use.
func (b *TransactionBuilder) GetInventory(inv inventory.Inventory) *TransactionBuilderInventory {
	for _, existing := range b.inventories {
		if existing.actualInventory == inv {
			return existing
		}
	}
	proxy := NewTransactionBuilderInventory(inv)
	b.inventories = append(b.inventories, proxy)
	return proxy
}

// GenerateActions returns the actions for every change made through the builder.
func (b *TransactionBuilder) GenerateActions() []InventoryAction {
	actions := append([]InventoryAction(nil), b.extraActions...)
	for _, inv := range b.inventories {
		for _, action := range inv.GenerateActions() {
			actions = append(actions, action)
		}
	}
	return actions
}

// TransactionBuilderInventory is a port of TransactionBuilderInventory: this class facilitates
// generating SlotChangeActions to build an inventory transaction. It wraps around the inventory
// you want to modify under transaction, and generates a diff of changes. This allows plugins to
// modify inventories under transaction, and then submit the transaction without having to create
// the actions manually.
type TransactionBuilderInventory struct {
	inventory.BaseInventory

	actualInventory inventory.Inventory
	changedSlots    []item.Item
}

func NewTransactionBuilderInventory(actualInventory inventory.Inventory) *TransactionBuilderInventory {
	t := &TransactionBuilderInventory{actualInventory: actualInventory, changedSlots: make([]item.Item, actualInventory.GetSize())}
	t.Init(t)
	return t
}

func (t *TransactionBuilderInventory) GetActualInventory() inventory.Inventory {
	return t.actualInventory
}

func (t *TransactionBuilderInventory) InternalSetContents(items map[int]item.Item) {
	for i := 0; i < t.GetSize(); i++ {
		if it, ok := items[i]; ok {
			t.SetItem(i, it)
		} else {
			t.Clear(i)
		}
	}
}

func (t *TransactionBuilderInventory) InternalSetItem(index int, it item.Item) {
	if !it.EqualsExact(t.actualInventory.GetItem(index)) {
		if it.IsNull() {
			t.changedSlots[index] = inventory.Air()
		} else {
			t.changedSlots[index] = it.Clone()
		}
	}
}

func (t *TransactionBuilderInventory) GetSize() int { return t.actualInventory.GetSize() }

func (t *TransactionBuilderInventory) GetItem(index int) item.Item {
	if index >= 0 && index < len(t.changedSlots) && t.changedSlots[index] != nil {
		return t.changedSlots[index].Clone()
	}
	return t.actualInventory.GetItem(index)
}

func (t *TransactionBuilderInventory) GetContents(includeEmpty bool) map[int]item.Item {
	contents := t.actualInventory.GetContents(includeEmpty)
	for index, it := range t.changedSlots {
		if it == nil {
			continue
		}
		if includeEmpty || !it.IsNull() {
			contents[index] = it.Clone()
		} else {
			delete(contents, index)
		}
	}
	return contents
}

// GenerateActions returns the slot changes made through this proxy.
func (t *TransactionBuilderInventory) GenerateActions() []*SlotChangeAction {
	var result []*SlotChangeAction
	for index, newItem := range t.changedSlots {
		if newItem == nil {
			continue
		}
		oldItem := t.actualInventory.GetItem(index)
		if !newItem.EqualsExact(oldItem) {
			result = append(result, NewSlotChangeAction(t.actualInventory, index, oldItem, newItem))
		}
	}
	return result
}
