package transaction

import (
	"fmt"
	"reflect"

	"pocketmine-go/pocketmine/event"
	inventoryevent "pocketmine-go/pocketmine/event/inventory"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

// transactionHooks are the InventoryTransaction methods subclasses (CraftingTransaction,
// EnchantingTransaction) override; Execute reaches them through self, like PHP's $this.
type transactionHooks interface {
	Validate() error
	CallExecuteEvent() bool
}

// InventoryTransaction is a port of pocketmine\inventory\transaction\InventoryTransaction: a set
// of actions that, taken together, move items around without creating or destroying any (unless
// a creative player is involved).
//
// This class does not care about the origin of the items (network, plugins, ...), only that the
// transaction balances and every action is valid.
type InventoryTransaction struct {
	self transactionHooks

	hasExecuted bool
	source      Player
	inventories []inventory.Inventory
	actions     []InventoryAction
}

// NewInventoryTransaction is a port of InventoryTransaction::__construct.
func NewInventoryTransaction(source Player, actions []InventoryAction) (*InventoryTransaction, error) {
	t := &InventoryTransaction{source: source}
	t.self = t
	for _, a := range actions {
		if err := t.AddAction(a); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// Init sets the transaction a subclass embeds this one in, for virtual dispatch.
func (t *InventoryTransaction) Init(self transactionHooks, source Player) {
	t.self = self
	t.source = source
}

func (t *InventoryTransaction) GetSource() Player { return t.source }

// GetSourcePlayer is GetSource for CraftItemEvent.
func (t *InventoryTransaction) GetSourcePlayer() inventoryevent.Player { return t.source }

// GetInventories returns the inventories involved in the transaction's slot changes.
func (t *InventoryTransaction) GetInventories() []inventory.Inventory {
	return append([]inventory.Inventory(nil), t.inventories...)
}

// GetActions returns an array of all the actions in the transaction. This is not ordered in any
// particular way.
func (t *InventoryTransaction) GetActions() []InventoryAction {
	return append([]InventoryAction(nil), t.actions...)
}

// AddAction is a port of InventoryTransaction::addAction.
func (t *InventoryTransaction) AddAction(action InventoryAction) error {
	for _, a := range t.actions {
		if a == action {
			return fmt.Errorf("Tried to add the same action to a transaction twice")
		}
	}
	t.actions = append(t.actions, action)
	action.OnAddToTransaction(t)
	if s, ok := action.(*SlotChangeAction); ok {
		known := false
		for _, inv := range t.inventories {
			if inv == s.GetInventory() {
				known = true
				break
			}
		}
		if !known {
			t.inventories = append(t.inventories, s.GetInventory())
		}
	}
	return nil
}

func (t *InventoryTransaction) removeAction(action InventoryAction) {
	for i, a := range t.actions {
		if a == action {
			t.actions = append(t.actions[:i], t.actions[i+1:]...)
			return
		}
	}
}

// MatchItems is a port of InventoryTransaction::matchItems: the items needed (targets) and had
// (sources) that don't cancel each other out.
func (t *InventoryTransaction) MatchItems() (needItems, haveItems []item.Item, err error) {
	for _, action := range t.actions {
		if targetItem := action.GetTargetItem(); !targetItem.IsNull() {
			needItems = append(needItems, targetItem)
		}
		if err := action.Validate(t.source); err != nil {
			return nil, nil, validationError(fmt.Sprintf("%s#%p: %s", reflect.TypeOf(action).Elem().Name(), action, err.Error()))
		}
		if sourceItem := action.GetSourceItem(); !sourceItem.IsNull() {
			haveItems = append(haveItems, sourceItem)
		}
	}

	for i := 0; i < len(needItems); i++ {
		needItem := needItems[i]
		for j := 0; j < len(haveItems); j++ {
			haveItem := haveItems[j]
			if needItem.CanStackWith(haveItem) {
				amount := min(needItem.GetCount(), haveItem.GetCount())
				needItem.SetCount(needItem.GetCount() - amount)
				haveItem.SetCount(haveItem.GetCount() - amount)
				if haveItem.GetCount() == 0 {
					haveItems = append(haveItems[:j], haveItems[j+1:]...)
					j--
				}
				if needItem.GetCount() == 0 {
					needItems = append(needItems[:i], needItems[i+1:]...)
					i--
					break
				}
			}
		}
	}
	return needItems, haveItems, nil
}

// squashDuplicateSlotChanges is a port of InventoryTransaction::squashDuplicateSlotChanges:
// processes transaction actions that change the same slot more than once, turning them into a
// single slot change from the original item to the final result.
func (t *InventoryTransaction) squashDuplicateSlotChanges() error {
	type slotKey struct {
		inv  inventory.Inventory
		slot int
	}
	var order []slotKey
	slotChanges := map[slotKey][]*SlotChangeAction{}
	for _, action := range t.actions {
		if s, ok := action.(*SlotChangeAction); ok {
			key := slotKey{s.GetInventory(), s.GetSlot()}
			if _, seen := slotChanges[key]; !seen {
				order = append(order, key)
			}
			slotChanges[key] = append(slotChanges[key], s)
		}
	}

	for _, key := range order {
		list := slotChanges[key]
		if len(list) == 1 { //No need to compact slot changes if there is only one on this slot
			continue
		}
		if !key.inv.SlotExists(key.slot) { //this can get hit for crafting tables because the validation happens after this compaction
			return validationError(fmt.Sprintf("Slot %d does not exist in inventory %T", key.slot, key.inv))
		}
		sourceItem := key.inv.GetItem(key.slot)

		targetItem := t.findResultItem(sourceItem, list)
		if targetItem == nil {
			return validationError(fmt.Sprintf("Failed to compact %d duplicate actions", len(list)))
		}

		for _, action := range list {
			t.removeAction(action)
		}

		if !targetItem.EqualsExact(sourceItem) {
			//sometimes we get actions on the crafting grid whose source and target items are the same, so dump them
			if err := t.AddAction(NewSlotChangeAction(key.inv, key.slot, sourceItem, targetItem)); err != nil {
				return err
			}
		}
	}
	return nil
}

// findResultItem is a port of InventoryTransaction::findResultItem.
func (t *InventoryTransaction) findResultItem(needOrigin item.Item, possibleActions []*SlotChangeAction) item.Item {
	var candidate *SlotChangeAction
	var newList []*SlotChangeAction
	for _, action := range possibleActions {
		if action.GetSourceItem().EqualsExact(needOrigin) {
			if candidate != nil {
				/*
				 * we found multiple possible actions that match the origin action
				 * this means that there are multiple ways that this chain could play out
				 * if we cared so much about this, we could build all the possible chains in parallel and see which
				 * variation managed to complete the chain, but this has an extremely high complexity which is not
				 * worth the trouble for this scenario (we don't usually expect to see chains longer than a couple
				 * of actions in here anyway), and might still result in multiple possible results.
				 */
				return nil
			}
			candidate = action
			continue
		}
		newList = append(newList, action)
	}

	if candidate == nil {
		//chaining is not possible with this origin, none of the actions are valid
		return nil
	}
	if len(newList) == 0 {
		return candidate.GetTargetItem()
	}
	return t.findResultItem(candidate.GetTargetItem(), newList)
}

// Validate is a port of InventoryTransaction::validate: verifies that the transaction is valid
// (it balances and every action is valid).
func (t *InventoryTransaction) Validate() error {
	if err := t.squashDuplicateSlotChanges(); err != nil {
		return err
	}
	needItems, haveItems, err := t.MatchItems()
	if err != nil {
		return err
	}
	if len(t.actions) == 0 {
		return validationError("Inventory transaction must have at least one action to be executable")
	}
	if len(haveItems) > 0 {
		return validationError("Transaction does not balance (tried to destroy some items)")
	}
	if len(needItems) > 0 {
		return validationError("Transaction does not balance (tried to create some items)")
	}
	return nil
}

// CallExecuteEvent is a port of InventoryTransaction::callExecuteEvent.
func (t *InventoryTransaction) CallExecuteEvent() bool {
	ev := inventoryevent.NewInventoryTransactionEvent(t.self)
	event.Call(ev)
	return !ev.IsCancelled()
}

// Execute is a port of InventoryTransaction::execute: executes the group of actions, returning a
// *TransactionValidationError if the transaction is invalid or a *TransactionCancelledError if a
// plugin cancelled it.
func (t *InventoryTransaction) Execute() error {
	if t.HasExecuted() {
		return validationError("Transaction has already been executed")
	}
	if err := t.self.Validate(); err != nil {
		return err
	}
	if !t.self.CallExecuteEvent() {
		return &TransactionCancelledError{Message: "Transaction event cancelled"}
	}
	for _, action := range t.actions {
		if !action.OnPreExecute(t.source) {
			return &TransactionCancelledError{Message: "One of the actions in this transaction was cancelled"}
		}
	}
	for _, action := range t.actions {
		action.Execute(t.source)
	}
	t.hasExecuted = true
	return nil
}

func (t *InventoryTransaction) HasExecuted() bool { return t.hasExecuted }
