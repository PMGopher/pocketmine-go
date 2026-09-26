package handler

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	blockinventory "pocketmine-go/pocketmine/block/inventory"
	"pocketmine-go/pocketmine/crafting"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/inventory/transaction"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/network/mcpe"
	"pocketmine-go/pocketmine/player"
)

// ItemStackRequestProcessError is a port of
// pocketmine\network\mcpe\handler\ItemStackRequestProcessException.
type ItemStackRequestProcessError struct {
	Message string
	Cause   error
}

func (e *ItemStackRequestProcessError) Error() string { return e.Message }
func (e *ItemStackRequestProcessError) Unwrap() error { return e.Cause }

func processError(format string, args ...any) *ItemStackRequestProcessError {
	return &ItemStackRequestProcessError{Message: fmt.Sprintf(format, args...)}
}

// ItemStackRequestExecutor is a port of pocketmine\network\mcpe\handler\ItemStackRequestExecutor:
// turns an ItemStackRequest into an InventoryTransaction.
type ItemStackRequestExecutor struct {
	player           *player.Player
	inventoryManager *mcpe.InventoryManager
	request          protocol.ItemStackRequest

	builder          *transaction.TransactionBuilder
	requestSlotInfos []protocol.StackRequestSlotInfo

	// specialTransaction is the CraftingTransaction or EnchantingTransaction the request builds
	// (nil for a plain InventoryTransaction).
	specialTransaction specialTransaction

	craftingResults []item.Item

	nextCreatedItem                  item.Item
	createdItemFromCreativeInventory bool
	createdItemsTakenCount           int
}

func NewItemStackRequestExecutor(p *player.Player, inventoryManager *mcpe.InventoryManager, request protocol.ItemStackRequest) *ItemStackRequestExecutor {
	return &ItemStackRequestExecutor{player: p, inventoryManager: inventoryManager, request: request, builder: transaction.NewTransactionBuilder()}
}

func prettyInventoryAndSlot(inv inventory.Inventory, slot int) string {
	if b, ok := inv.(*transaction.TransactionBuilderInventory); ok {
		inv = b.GetActualInventory()
	}
	return fmt.Sprintf("%T#%p, slot: %d", inv, inv, slot)
}

// matchItemStack is a port of ItemStackRequestExecutor::matchItemStack.
func (e *ItemStackRequestExecutor) matchItemStack(inv inventory.Inventory, slotID int, clientItemStackID int32) error {
	info := e.inventoryManager.GetItemStackInfo(inv, slotID)
	if info == nil {
		panic("The inventory is tracked and the slot is valid, so this should not be null")
	}
	var matches bool
	if clientItemStackID < 0 {
		matches = info.GetRequestID() != nil && *info.GetRequestID() == clientItemStackID
	} else {
		matches = info.GetStackID() == clientItemStackID
	}
	if !matches {
		lastRequest := "none"
		if id := info.GetRequestID(); id != nil {
			lastRequest = fmt.Sprint(*id)
		}
		return processError("%s: Mismatched expected itemstack, client expected: %d, server actual: %d, last modified by request: %s",
			prettyInventoryAndSlot(inv, slotID), clientItemStackID, info.GetStackID(), lastRequest)
	}
	return nil
}

// getBuilderInventoryAndSlot is a port of ItemStackRequestExecutor::getBuilderInventoryAndSlot.
func (e *ItemStackRequestExecutor) getBuilderInventoryAndSlot(info protocol.StackRequestSlotInfo) (*transaction.TransactionBuilderInventory, int, error) {
	windowID, slotID, err := TranslateItemStackContainerID(info.Container.ContainerID, e.inventoryManager.GetCurrentWindowID(), int(info.Slot))
	if err != nil {
		return nil, 0, err
	}
	inv, slot, ok := e.inventoryManager.LocateWindowAndSlot(windowID, slotID)
	if !ok {
		return nil, 0, processError("No open inventory matches container UI ID: %d, slot ID: %d", info.Container.ContainerID, info.Slot)
	}
	if !inv.SlotExists(slot) {
		return nil, 0, processError("No such inventory slot :%s", prettyInventoryAndSlot(inv, slot))
	}

	if info.StackNetworkID != e.request.RequestID { //the itemstack may have been modified by the current request
		if err := e.matchItemStack(inv, slot, info.StackNetworkID); err != nil {
			return nil, 0, err
		}
	}
	return e.builder.GetInventory(inv), slot, nil
}

// transferItems is a port of ItemStackRequestExecutor::transferItems.
func (e *ItemStackRequestExecutor) transferItems(source, destination protocol.StackRequestSlotInfo, count int) error {
	removed, err := e.removeItemFromSlot(source, count)
	if err != nil {
		return err
	}
	return e.addItemToSlot(destination, removed, count)
}

// removeItemFromSlot is a port of ItemStackRequestExecutor::removeItemFromSlot: deducts items
// from an inventory slot, returning a stack containing the removed items.
func (e *ItemStackRequestExecutor) removeItemFromSlot(slotInfo protocol.StackRequestSlotInfo, count int) (item.Item, error) {
	if slotInfo.Container.ContainerID == protocol.ContainerCreatedOutput && int(slotInfo.Slot) == mcpe.UISlotCreatedItemOutput {
		//special case for the "created item" output slot
		//TODO: do we need to send a response for this slot info?
		return e.takeCreatedItem(count)
	}
	e.requestSlotInfos = append(e.requestSlotInfos, slotInfo)
	inv, slot, err := e.getBuilderInventoryAndSlot(slotInfo)
	if err != nil {
		return nil, err
	}
	if count < 1 {
		//this should be impossible at the protocol level, but in case of buggy core code this will prevent exploits
		return nil, processError("%s: Cannot take less than 1 items from a stack", prettyInventoryAndSlot(inv, slot))
	}

	existingItem := inv.GetItem(slot)
	if existingItem.GetCount() < count {
		return nil, processError("%s: Cannot take %d items from a stack of %d", prettyInventoryAndSlot(inv, slot), count, existingItem.GetCount())
	}

	removed := existingItem.PopCount(count)
	inv.SetItem(slot, existingItem)
	return removed, nil
}

// addItemToSlot is a port of ItemStackRequestExecutor::addItemToSlot: adds items to the target
// slot, if they are stackable.
func (e *ItemStackRequestExecutor) addItemToSlot(slotInfo protocol.StackRequestSlotInfo, it item.Item, count int) error {
	e.requestSlotInfos = append(e.requestSlotInfos, slotInfo)
	inv, slot, err := e.getBuilderInventoryAndSlot(slotInfo)
	if err != nil {
		return err
	}
	if count < 1 {
		//this should be impossible at the protocol level, but in case of buggy core code this will prevent exploits
		return processError("%s: Cannot take less than 1 items from a stack", prettyInventoryAndSlot(inv, slot))
	}

	existingItem := inv.GetItem(slot)
	if !existingItem.IsNull() && !existingItem.CanStackWith(it) {
		return processError("%s: Can only add items to an empty slot, or a slot containing the same item", prettyInventoryAndSlot(inv, slot))
	}

	//we can't use the existing item here; it may be an empty stack
	newItem := it.Clone()
	newItem.SetCount(existingItem.GetCount() + count)
	inv.SetItem(slot, newItem)
	return nil
}

// setNextCreatedItem is a port of ItemStackRequestExecutor::setNextCreatedItem.
func (e *ItemStackRequestExecutor) setNextCreatedItem(it item.Item, creative bool) error {
	if it != nil && it.IsNull() {
		it = nil
	}
	if e.nextCreatedItem != nil {
		//while this is more complicated than simply adding the action when the item is taken, this ensures that
		//plugins can tell the difference between 1 item that got split into 2 slots, vs 2 separate items.
		if e.createdItemFromCreativeInventory && e.createdItemsTakenCount > 0 {
			e.nextCreatedItem.SetCount(e.createdItemsTakenCount)
			e.builder.AddAction(transaction.NewCreateItemAction(e.nextCreatedItem))
		} else if e.createdItemsTakenCount < e.nextCreatedItem.GetCount() {
			return processError("Not all of the previous created item was taken")
		}
	}
	e.nextCreatedItem = it
	e.createdItemFromCreativeInventory = creative
	e.createdItemsTakenCount = 0
	return nil
}

// specialTransaction is the part of CraftingTransaction/EnchantingTransaction the executor uses.
type specialTransaction interface {
	AddAction(action transaction.InventoryAction) error
	Execute() error
	GetActions() []transaction.InventoryAction
}

// craftingManagerOwner is Server::getCraftingManager.
type craftingManagerOwner interface {
	GetCraftingManager() *crafting.CraftingManager
}

// beginCrafting is a port of ItemStackRequestExecutor::beginCrafting.
func (e *ItemStackRequestExecutor) beginCrafting(recipeID uint32, repetitions int) error {
	if e.specialTransaction != nil {
		return processError("Another special transaction is already in progress")
	}
	if repetitions < 1 {
		return processError("Cannot craft a recipe less than 1 time")
	}
	if repetitions > 256 {
		//TODO: we can probably lower this limit to 64, but I'm unsure if there are cases where the client may
		//request more than 64 repetitions of a recipe.
		//It's already hard-limited to 256 repetitions in the protocol, so this is just a sanity check.
		return processError("Cannot craft a recipe more than 256 times")
	}
	owner, ok := e.player.GetServer().(craftingManagerOwner)
	if !ok {
		return processError("No crafting manager")
	}
	craftingManager := owner.GetCraftingManager()
	recipeIndex := int(recipeID) - mcpe.RecipeIDOffset
	recipe := craftingManager.GetCraftingRecipeFromIndex(recipeIndex)
	if recipe == nil {
		return processError("No such crafting recipe index: %d", recipeIndex)
	}

	tx, err := transaction.NewCraftingTransaction(e.player, craftingManager, nil, recipe, repetitions)
	if err != nil {
		return &ItemStackRequestProcessError{Message: err.Error(), Cause: err}
	}
	e.specialTransaction = tx

	//TODO: Since the system assumes that crafting can only be done in the crafting grid, we have to give it a
	//crafting grid to make the API happy. No implementation of getResultsFor() actually uses the crafting grid
	//right now, so this will work, but this will become a problem in the future for things like shulker boxes and
	//custom crafting recipes.
	var grid *crafting.CraftingGrid
	if g, ok := e.player.GetCraftingGrid().(interface{ Grid() *crafting.CraftingGrid }); ok {
		grid = g.Grid()
	}
	for _, craftingResult := range recipe.GetResultsFor(grid) {
		craftingResult.SetCount(craftingResult.GetCount() * repetitions)
		e.craftingResults = append(e.craftingResults, craftingResult)
	}
	if len(e.craftingResults) == 1 {
		//for multi-output recipes, later actions will tell us which result to create and when
		return e.setNextCreatedItem(e.craftingResults[0], false)
	}
	return nil
}

// takeCreatedItem is a port of ItemStackRequestExecutor::takeCreatedItem.
func (e *ItemStackRequestExecutor) takeCreatedItem(count int) (item.Item, error) {
	if count < 1 {
		//this should be impossible at the protocol level, but in case of buggy core code this will prevent exploits
		return nil, processError("Cannot take less than 1 created item")
	}
	createdItem := e.nextCreatedItem
	if createdItem == nil {
		return nil, processError("No created item is waiting to be taken")
	}

	if !e.createdItemFromCreativeInventory {
		availableCount := createdItem.GetCount() - e.createdItemsTakenCount
		if count > availableCount {
			return nil, processError("Not enough created items available to be taken (have %d, tried to take %d)", availableCount, count)
		}
	}

	e.createdItemsTakenCount += count
	takenItem := createdItem.Clone()
	takenItem.SetCount(count)
	if !e.createdItemFromCreativeInventory && e.createdItemsTakenCount >= createdItem.GetCount() {
		if err := e.setNextCreatedItem(nil, false); err != nil {
			return nil, err
		}
	}
	return takenItem, nil
}

// assertDoingCrafting is a port of ItemStackRequestExecutor::assertDoingCrafting.
func (e *ItemStackRequestExecutor) assertDoingCrafting() error {
	switch e.specialTransaction.(type) {
	case *transaction.CraftingTransaction, *transaction.EnchantingTransaction:
		return nil
	case nil:
		return processError("Expected CraftRecipe or CraftRecipeAuto action to precede this action")
	}
	return processError("A different special transaction is already in progress")
}

// durableForPrediction is pocketmine\item\Durable as MineBlockStackRequestAction needs it.
type durableForPrediction interface {
	GetMaxDurability() int
	SetDamage(damage int)
}

// processItemStackRequestAction is a port of ItemStackRequestExecutor::processItemStackRequestAction.
func (e *ItemStackRequestExecutor) processItemStackRequestAction(action protocol.StackRequestAction) error {
	switch a := action.(type) {
	case *protocol.TakeStackRequestAction:
		return e.transferItems(a.Source, a.Destination, int(a.Count))
	case *protocol.PlaceStackRequestAction:
		return e.transferItems(a.Source, a.Destination, int(a.Count))
	case *protocol.SwapStackRequestAction:
		e.requestSlotInfos = append(e.requestSlotInfos, a.Source, a.Destination)

		inventory1, slot1, err := e.getBuilderInventoryAndSlot(a.Source)
		if err != nil {
			return err
		}
		inventory2, slot2, err := e.getBuilderInventoryAndSlot(a.Destination)
		if err != nil {
			return err
		}
		item1 := inventory1.GetItem(slot1)
		item2 := inventory2.GetItem(slot2)
		inventory1.SetItem(slot1, item2)
		inventory2.SetItem(slot2, item1)
	case *protocol.DropStackRequestAction:
		//TODO: this action has a "randomly" field, I have no idea what it's used for
		dropped, err := e.removeItemFromSlot(a.Source, int(a.Count))
		if err != nil {
			return err
		}
		e.builder.AddAction(transaction.NewDropItemAction(dropped))
	case *protocol.ConsumeStackRequestAction:
		// CraftingConsumeInputStackRequestAction
		if err := e.assertDoingCrafting(); err != nil {
			return err
		}
		_, err := e.removeItemFromSlot(a.Source, int(a.Count)) //output discarded - we allow CraftingTransaction to verify the balance
		return err
	case *protocol.DestroyStackRequestAction:
		destroyed, err := e.removeItemFromSlot(a.Source, int(a.Count))
		if err != nil {
			return err
		}
		e.builder.AddAction(transaction.NewDestroyItemAction(destroyed))
	case *protocol.CraftCreativeStackRequestAction:
		it := e.player.GetCreativeInventory().GetItem(int(a.CreativeItemNetworkID))
		if it == nil {
			return processError("No such creative item index: %d", a.CreativeItemNetworkID)
		}
		return e.setNextCreatedItem(it, true)
	case *protocol.CraftRecipeStackRequestAction:
		if window, ok := e.player.GetCurrentWindow().(*blockinventory.EnchantInventory); ok {
			optionID, found := e.inventoryManager.GetEnchantingTableOptionIndex(int(a.RecipeNetworkID))
			if found {
				if option := window.GetOption(optionID); option != nil {
					e.specialTransaction = transaction.NewEnchantingTransaction(e.player, option, optionID+1)
					return e.setNextCreatedItem(window.GetOutput(optionID), false)
				}
			}
			return nil
		}
		return e.beginCrafting(a.RecipeNetworkID, int(a.NumberOfCrafts))
	case *protocol.AutoCraftRecipeStackRequestAction:
		return e.beginCrafting(a.RecipeNetworkID, int(a.NumberOfCrafts))
	case *protocol.CreateStackRequestAction:
		// CraftingCreateSpecificResultStackRequestAction
		if err := e.assertDoingCrafting(); err != nil {
			return err
		}
		index := int(a.ResultsSlot)
		if index >= len(e.craftingResults) {
			return processError("No such crafting result index: %d", a.ResultsSlot)
		}
		return e.setNextCreatedItem(e.craftingResults[index], false)
	case *protocol.CraftResultsDeprecatedStackRequestAction:
		//no obvious use
	case *protocol.MineBlockStackRequestAction:
		slot := int(a.HotbarSlot)
		e.requestSlotInfos = append(e.requestSlotInfos, protocol.StackRequestSlotInfo{
			Container:      protocol.FullContainerName{ContainerID: protocol.ContainerHotBar},
			Slot:           byte(slot),
			StackNetworkID: a.StackNetworkID,
		})
		inv := e.player.GetInventory()
		if inv.SlotExists(slot) {
			usedItem := inv.GetItem(slot)
			predictedDamage := int(a.PredictedDurability)
			if durable, ok := usedItem.(durableForPrediction); ok && predictedDamage >= 0 && predictedDamage <= durable.GetMaxDurability() {
				durable.SetDamage(predictedDamage)
				e.inventoryManager.AddPredictedSlotChange(inv, slot, usedItem)
			}
		}
	default:
		return processError("Unhandled item stack request action")
	}
	return nil
}

// GenerateInventoryTransaction is a port of ItemStackRequestExecutor::generateInventoryTransaction:
// nil when the request only carried predictions.
func (e *ItemStackRequestExecutor) GenerateInventoryTransaction() (specialTransaction, error) {
	for k, action := range e.request.Actions {
		if err := e.processItemStackRequestAction(action); err != nil {
			return nil, &ItemStackRequestProcessError{Message: fmt.Sprintf("Error processing action %d (%T): %s", k, action, err.Error()), Cause: err}
		}
	}
	if err := e.setNextCreatedItem(nil, false); err != nil {
		return nil, err
	}
	inventoryActions := e.builder.GenerateActions()
	if len(inventoryActions) == 0 {
		return nil, nil
	}

	var tx specialTransaction = e.specialTransaction
	if tx == nil {
		plain, err := transaction.NewInventoryTransaction(e.player, nil)
		if err != nil {
			return nil, err
		}
		tx = plain
	}
	for _, action := range inventoryActions {
		if err := tx.AddAction(action); err != nil {
			return nil, &ItemStackRequestProcessError{Message: err.Error(), Cause: err}
		}
	}
	return tx, nil
}

// GetItemStackResponseBuilder is a port of ItemStackRequestExecutor::getItemStackResponseBuilder.
func (e *ItemStackRequestExecutor) GetItemStackResponseBuilder() *ItemStackResponseBuilder {
	builder := NewItemStackResponseBuilder(e.request.RequestID, e.inventoryManager)
	for _, requestInfo := range e.requestSlotInfos {
		builder.AddSlot(requestInfo.Container.ContainerID, int(requestInfo.Slot))
	}
	return builder
}

// BuildItemStackResponse is a port of ItemStackRequestExecutor::buildItemStackResponse.
func (e *ItemStackRequestExecutor) BuildItemStackResponse() protocol.ItemStackResponse {
	return e.GetItemStackResponseBuilder().Build()
}
