package blockinventory

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/world/sound"
)

// EnderChestInventory is a port of pocketmine\block\inventory\EnderChestInventory: a view of a
// player's ender inventory through an ender chest block.
type EnderChestInventory struct {
	*inventory.DelegateInventory
	BlockInventoryTrait

	inventory *inventory.PlayerEnderInventory
}

func NewEnderChestInventory(holder block.Position, enderInventory *inventory.PlayerEnderInventory) *EnderChestInventory {
	e := &EnderChestInventory{
		DelegateInventory:   inventory.NewDelegateInventory(enderInventory),
		BlockInventoryTrait: BlockInventoryTrait{Holder: holder},
		inventory:           enderInventory,
	}
	// Dispatch BaseInventory's $this (listeners, viewers' sync) to the outer inventory.
	e.Init(e)
	return e
}

func (e *EnderChestInventory) GetEnderInventory() *inventory.PlayerEnderInventory { return e.inventory }

// enderChestTile is the EnderChest tile at the holder, if any.
func (e *EnderChestInventory) enderChestTile() *tile.EnderChest {
	world, err := e.Holder.GetWorld()
	if err != nil {
		return nil
	}
	t, ok := world.GetTile(e.Holder)
	if !ok {
		return nil
	}
	enderChest, _ := t.(*tile.EnderChest)
	return enderChest
}

// GetViewerCount is a port of EnderChestInventory::getViewerCount: the ender chest tile counts
// the viewers of every player's EnderChestInventory at that block.
func (e *EnderChestInventory) GetViewerCount() int {
	enderChest := e.enderChestTile()
	if enderChest == nil {
		return 0
	}
	return enderChest.GetViewerCount()
}

// OnOpen is a port of AnimatedBlockInventoryTrait::onOpen.
func (e *EnderChestInventory) OnOpen(who inventory.Player) {
	e.DelegateInventory.OnOpen(who)
	if e.Holder.IsValid() && e.GetViewerCount() == 1 {
		//TODO: this crap really shouldn't be managed by the inventory
		e.animateBlock(true)
		if world, err := e.Holder.GetWorld(); err == nil {
			world.AddSound(e.Holder.Add(0.5, 0.5, 0.5), sound.EnderChestOpenSound{})
		}
	}
}

// OnClose is a port of EnderChestInventory::onClose (AnimatedBlockInventoryTrait::onClose, then
// one viewer less on the tile).
func (e *EnderChestInventory) OnClose(who inventory.Player) {
	if e.Holder.IsValid() && e.GetViewerCount() == 1 {
		//TODO: this crap really shouldn't be managed by the inventory
		e.animateBlock(false)
		if world, err := e.Holder.GetWorld(); err == nil {
			world.AddSound(e.Holder.Add(0.5, 0.5, 0.5), sound.EnderChestCloseSound{})
		}
	}
	e.DelegateInventory.OnClose(who)

	if enderChest := e.enderChestTile(); enderChest != nil {
		enderChest.SetViewerCount(enderChest.GetViewerCount() - 1)
	}
}

// animateBlock is a port of EnderChestInventory::animateBlock.
func (e *EnderChestInventory) animateBlock(isOpen bool) { broadcastChestEvent(e.Holder, isOpen) }
