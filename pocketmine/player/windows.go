package player

import (
	"fmt"
	"pocketmine-go/pocketmine/item"

	blockinventory "pocketmine-go/pocketmine/block/inventory"
	"pocketmine-go/pocketmine/event"
	inventoryevent "pocketmine-go/pocketmine/event/inventory"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/inventory/transaction"
)

// addDefaultWindows is a port of Player::addDefaultWindows.
func (p *Player) addDefaultWindows() {
	p.cursorInventory = inventory.NewPlayerCursorInventory(p)
	p.craftingGrid = blockinventory.NewPlayerCraftingInventory(p)

	p.addPermanentInventories(p.GetInventory(), p.GetArmorInventory(), p.cursorInventory, p.GetOffHandInventory(), p.craftingGrid)

	//TODO: more windows
}

// GetCursorInventory is a port of Player::getCursorInventory.
func (p *Player) GetCursorInventory() *inventory.PlayerCursorInventory { return p.cursorInventory }

// GetCraftingGrid is a port of Player::getCraftingGrid.
func (p *Player) GetCraftingGrid() inventory.TemporaryInventory { return p.craftingGrid }

// GetCreativeInventory returns the creative inventory shown to the player. Unless changed by a
// plugin, this is usually the same for all players.
func (p *Player) GetCreativeInventory() *inventory.CreativeInventory { return p.creativeInventory }

// SetCreativeInventory sets a custom creative inventory (make a Clone of an existing one).
func (p *Player) SetCreativeInventory(inv *inventory.CreativeInventory) {
	p.creativeInventory = inv
	if p.spawned && p.IsConnected() {
		if m := p.GetNetworkSession().GetInvManager(); m != nil {
			m.SyncCreative()
		}
	}
}

// doCloseInventory is a port of Player::doCloseInventory: called to clean up crafting grid and
// cursor inventory when it is detected that the player closed their inventory.
func (p *Player) doCloseInventory() {
	inventories := []inventory.Inventory{p.craftingGrid, p.cursorInventory}
	if temp, ok := p.currentWindow.(inventory.TemporaryInventory); ok {
		inventories = append(inventories, temp)
	}

	builder := transaction.NewTransactionBuilder()
	for _, inv := range inventories {
		contents := inv.GetContents(false)

		if len(contents) > 0 {
			var items = make([]item.Item, 0, len(contents))
			for slot := 0; slot < inv.GetSize(); slot++ {
				if it, ok := contents[slot]; ok {
					items = append(items, it)
				}
			}
			drops := builder.GetInventory(p.GetInventory()).AddItem(items...)
			for _, drop := range drops {
				builder.AddAction(transaction.NewDropItemAction(drop))
			}

			builder.GetInventory(inv).ClearAll()
		}
	}

	actions := builder.GenerateActions()
	if len(actions) != 0 {
		tx, err := transaction.NewInventoryTransaction(p, actions)
		if err == nil {
			err = tx.Execute()
		}
		switch err.(type) {
		case nil:
			p.logger.Debug("Successfully evacuated items from temporary inventories")
		case *transaction.TransactionCancelledError:
			p.logger.Debug("Plugin cancelled transaction evacuating items from temporary inventories; items will be destroyed")
			for _, inv := range inventories {
				inv.ClearAll()
			}
		default:
			panic(fmt.Sprintf("This server-generated transaction should never be invalid: %v", err))
		}
	}
}

// GetCurrentWindow returns the inventory the player is currently viewing. This might be a chest,
// furnace, or any other container.
func (p *Player) GetCurrentWindow() inventory.Inventory { return p.currentWindow }

// opener is the part of an inventory with its own open/close behaviour (block inventories
// animating their block).
type opener interface {
	OnOpen(who inventory.Player)
	OnClose(who inventory.Player)
}

// SetCurrentWindow is a port of Player::setCurrentWindow: opens an inventory window to the
// player. Returns whether it was successful.
func (p *Player) SetCurrentWindow(inv inventory.Inventory) bool {
	if inv == p.currentWindow {
		return true
	}
	ev := inventoryevent.NewInventoryOpenEvent(inv, p)
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}

	p.RemoveCurrentWindow()

	inventoryManager := p.GetNetworkSession().GetInvManager()
	if inventoryManager == nil {
		panic("Player cannot open inventories in this state")
	}
	p.logger.Debug(fmt.Sprintf("Opening inventory %T#%p", inv, inv))
	inventoryManager.OnCurrentWindowChange(inv)
	inv.OnOpen(p)
	p.currentWindow = inv
	return true
}

// RemoveCurrentWindow is a port of Player::removeCurrentWindow.
func (p *Player) RemoveCurrentWindow() {
	p.doCloseInventory()
	if p.currentWindow != nil {
		currentWindow := p.currentWindow
		p.logger.Debug(fmt.Sprintf("Closing inventory %T#%p", currentWindow, currentWindow))
		currentWindow.OnClose(p)
		if p.networkSession != nil {
			if inventoryManager := p.networkSession.GetInvManager(); inventoryManager != nil {
				inventoryManager.OnCurrentWindowRemove()
			}
		}
		p.currentWindow = nil
		event.Call(inventoryevent.NewInventoryCloseEvent(currentWindow, p))
	}
}

// addPermanentInventories is a port of Player::addPermanentInventories.
func (p *Player) addPermanentInventories(inventories ...inventory.Inventory) {
	for _, inv := range inventories {
		inv.OnOpen(p)
		p.permanentWindows = append(p.permanentWindows, inv)
	}
}

// removePermanentInventories is a port of Player::removePermanentInventories.
func (p *Player) removePermanentInventories() {
	for _, inv := range p.permanentWindows {
		inv.OnClose(p)
	}
	p.permanentWindows = nil
}

// GetPermanentWindows returns the inventories the player always has open (its own inventory,
// armor, cursor, off-hand and crafting grid).
func (p *Player) GetPermanentWindows() []inventory.Inventory {
	return append([]inventory.Inventory(nil), p.permanentWindows...)
}
