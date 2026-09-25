package handler

import (
	"slices"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/network/mcpe"
)

// ItemStackResponseBuilder is a port of pocketmine\network\mcpe\handler\ItemStackResponseBuilder.
type ItemStackResponseBuilder struct {
	requestID        int32
	inventoryManager *mcpe.InventoryManager
	// changedSlots is container UI ID => set of slot IDs, in first-seen order.
	containerOrder []byte
	changedSlots   map[byte][]int
}

func NewItemStackResponseBuilder(requestID int32, inventoryManager *mcpe.InventoryManager) *ItemStackResponseBuilder {
	return &ItemStackResponseBuilder{requestID: requestID, inventoryManager: inventoryManager, changedSlots: map[byte][]int{}}
}

// AddSlot is a port of ItemStackResponseBuilder::addSlot.
func (b *ItemStackResponseBuilder) AddSlot(containerInterfaceID byte, slotID int) {
	slots, ok := b.changedSlots[containerInterfaceID]
	if !ok {
		b.containerOrder = append(b.containerOrder, containerInterfaceID)
	}
	if !slices.Contains(slots, slotID) {
		b.changedSlots[containerInterfaceID] = append(slots, slotID)
	}
}

func (b *ItemStackResponseBuilder) getInventoryAndSlot(containerInterfaceID byte, slotID int) (inventory.Inventory, int, bool) {
	windowID, slotID, err := TranslateItemStackContainerID(containerInterfaceID, b.inventoryManager.GetCurrentWindowID(), slotID)
	if err != nil {
		return nil, 0, false
	}
	inv, slot, ok := b.inventoryManager.LocateWindowAndSlot(windowID, slotID)
	if !ok || !inv.SlotExists(slot) {
		return nil, 0, false
	}
	return inv, slot, true
}

// durableItem is pocketmine\item\Durable as the response builder needs it.
type durableItem interface {
	GetDamage() int
}

// Build is a port of ItemStackResponseBuilder::build.
func (b *ItemStackResponseBuilder) Build() protocol.ItemStackResponse {
	var containerInfos []protocol.StackResponseContainerInfo
	for _, containerInterfaceID := range b.containerOrder {
		if containerInterfaceID == protocol.ContainerCreatedOutput {
			continue
		}
		var slotInfos []protocol.StackResponseSlotInfo
		for _, slotID := range b.changedSlots[containerInterfaceID] {
			inv, slot, ok := b.getInventoryAndSlot(containerInterfaceID, slotID)
			if !ok {
				//a plugin may have closed the inventory during an event, or the slot may have been invalid
				continue
			}
			info := b.inventoryManager.GetItemStackInfo(inv, slot)
			if info == nil {
				panic("ItemStackInfo should never be null for an open inventory")
			}
			it := inv.GetItem(slot)
			damage := 0
			if d, ok := it.(durableItem); ok {
				damage = d.GetDamage()
			}
			slotInfos = append(slotInfos, protocol.StackResponseSlotInfo{
				Slot:                 byte(slotID),
				HotbarSlot:           byte(slotID),
				Count:                byte(it.GetCount()),
				StackNetworkID:       info.GetStackID(),
				CustomName:           it.GetCustomName(),
				FilteredCustomName:   protocol.Option(it.GetCustomName()),
				DurabilityCorrection: int32(damage),
			})
		}
		if len(slotInfos) > 0 {
			containerInfos = append(containerInfos, protocol.StackResponseContainerInfo{
				Container: protocol.FullContainerName{ContainerID: containerInterfaceID},
				SlotInfo:  slotInfos,
			})
		}
	}
	return protocol.ItemStackResponse{Status: protocol.ItemStackResponseStatusOK, RequestID: b.requestID, ContainerInfo: containerInfos}
}
