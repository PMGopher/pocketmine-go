package mcpe

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	blockinventory "pocketmine-go/pocketmine/block/inventory"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/network"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/player"
)

// Container IDs, BedrockProtocol's ContainerIds.
const (
	ContainerIDNone           = -1
	ContainerIDInventory      = 0
	ContainerIDFirst          = 1
	ContainerIDLast           = 100
	ContainerIDOffhand        = 119
	ContainerIDArmor          = 120
	ContainerIDHotbar         = 122
	ContainerIDFixedInventory = 123
	ContainerIDUI             = 124
)

// Window types, BedrockProtocol's WindowTypes (sent as a byte).
const (
	WindowTypeInventory     = -1
	WindowTypeContainer     = 0
	WindowTypeWorkbench     = 1
	WindowTypeFurnace       = 2
	WindowTypeEnchantment   = 3
	WindowTypeBrewingStand  = 4
	WindowTypeAnvil         = 5
	WindowTypeHopper        = 8
	WindowTypeLoom          = 24
	WindowTypeBlastFurnace  = 27
	WindowTypeSmoker        = 28
	WindowTypeStonecutter   = 29
	WindowTypeCartography   = 30
	WindowTypeSmithingTable = 33
)

// UI inventory slot offsets, BedrockProtocol's UIInventorySlotOffset: net slot => core slot maps
// for the inventories that live in the UI container.
var (
	UISlotCursor            = map[int]int{0: 0}
	UISlotAnvil             = map[int]int{1: 0, 2: 1}
	UISlotStoneCutterInput  = 3
	UISlotLoom              = map[int]int{9: 0, 10: 1, 11: 2}
	UISlotCartographyTable  = map[int]int{12: 0, 13: 1}
	UISlotEnchantingTable   = map[int]int{14: 0, 15: 1}
	UISlotGrindstone        = map[int]int{16: 0, 17: 1}
	UISlotCrafting2x2Input  = map[int]int{28: 0, 29: 1, 30: 2, 31: 3}
	UISlotCrafting3x3Input  = map[int]int{32: 0, 33: 1, 34: 2, 35: 3, 36: 4, 37: 5, 38: 6, 39: 7, 40: 8}
	UISlotCreatedItemOutput = 50
	UISlotSmithingTable     = map[int]int{51: 0, 52: 1, 53: 2}
)

// ItemStackInfo is a port of pocketmine\network\mcpe\ItemStackInfo: the network stack ID of an
// item in a slot, and the item stack request (if any) that caused it to be there.
type ItemStackInfo struct {
	requestID *int32
	stackID   int32
}

func NewItemStackInfo(requestID *int32, stackID int32) *ItemStackInfo {
	return &ItemStackInfo{requestID: requestID, stackID: stackID}
}

func (i *ItemStackInfo) GetRequestID() *int32 { return i.requestID }
func (i *ItemStackInfo) GetStackID() int32    { return i.stackID }

// ComplexInventoryMapEntry is a port of pocketmine\network\mcpe\ComplexInventoryMapEntry: an
// inventory whose slots live in the UI container at the given net slots.
type ComplexInventoryMapEntry struct {
	inventory inventory.Inventory
	// slotMap is net slot => core slot; reverseSlotMap core slot => net slot.
	slotMap        map[int]int
	reverseSlotMap map[int]int
}

func NewComplexInventoryMapEntry(inv inventory.Inventory, slotMap map[int]int) *ComplexInventoryMapEntry {
	e := &ComplexInventoryMapEntry{inventory: inv, slotMap: slotMap, reverseSlotMap: make(map[int]int, len(slotMap))}
	for net, core := range slotMap {
		e.reverseSlotMap[core] = net
	}
	return e
}

func (e *ComplexInventoryMapEntry) GetInventory() inventory.Inventory { return e.inventory }
func (e *ComplexInventoryMapEntry) GetSlotMap() map[int]int           { return e.slotMap }

// MapNetToCore is ComplexInventoryMapEntry::mapNetToCore.
func (e *ComplexInventoryMapEntry) MapNetToCore(slot int) (int, bool) {
	core, ok := e.slotMap[slot]
	return core, ok
}

// MapCoreToNet is ComplexInventoryMapEntry::mapCoreToNet.
func (e *ComplexInventoryMapEntry) MapCoreToNet(slot int) (int, bool) {
	net, ok := e.reverseSlotMap[slot]
	return net, ok
}

// InventoryManagerEntry is a port of pocketmine\network\mcpe\InventoryManagerEntry.
type InventoryManagerEntry struct {
	inventory      inventory.Inventory
	complexSlotMap *ComplexInventoryMapEntry
	// predictions is slot => the item the client predicted it to hold.
	predictions map[int]protocol.ItemStack
	// itemStackInfos is slot => the stack ID info last sent for that slot.
	itemStackInfos map[int]*ItemStackInfo
	// pendingSyncs is slot => the item to send at the end of the tick.
	pendingSyncs map[int]protocol.ItemStack
}

func newInventoryManagerEntry(inv inventory.Inventory, complexSlotMap *ComplexInventoryMapEntry) *InventoryManagerEntry {
	return &InventoryManagerEntry{
		inventory:      inv,
		complexSlotMap: complexSlotMap,
		predictions:    map[int]protocol.ItemStack{},
		itemStackInfos: map[int]*ItemStackInfo{},
		pendingSyncs:   map[int]protocol.ItemStack{},
	}
}

// ContainerOpenFunc is InventoryManager's ContainerOpenClosure: it returns the packets opening
// inv as window id, or nil if it doesn't know inv.
type ContainerOpenFunc func(id int, inv inventory.Inventory) []packet.Packet

// InventoryManager is a port of pocketmine\network\mcpe\InventoryManager: tracks which inventories
// the client can see, under which window IDs, the network stack IDs of their items and the
// client's predictions, and keeps the client in sync.
//
// Not ported: furnace, brewing stand, anvil and hopper window types (those block inventories
// aren't ported), and syncEnchantingTableOptions (EnchantingHelper isn't ported).
type InventoryManager struct {
	player  *player.Player
	session *NetworkSession

	// inventories is PHP's spl_object_id(Inventory) => entry map; order keeps PHP's insertion
	// order for syncAll.
	inventories map[inventory.Inventory]*InventoryManagerEntry
	order       []inventory.Inventory

	networkIDToInventoryMap   map[int]inventory.Inventory
	complexSlotToInventoryMap map[int]*ComplexInventoryMapEntry

	lastInventoryNetworkID int
	currentWindowType      int

	clientSelectedHotbarSlot int

	containerOpenCallbacks []ContainerOpenFunc

	pendingCloseWindowID      *int
	pendingOpenWindowCallback func()

	nextItemStackID           int32
	currentItemStackRequestID *int32

	fullSyncRequested bool

	heldItemIndexListener *inventory.HeldItemIndexChangeListener
}

// NewInventoryManager is a port of InventoryManager::__construct.
func NewInventoryManager(p *player.Player, session *NetworkSession) *InventoryManager {
	m := &InventoryManager{
		player:                    p,
		session:                   session,
		inventories:               map[inventory.Inventory]*InventoryManagerEntry{},
		networkIDToInventoryMap:   map[int]inventory.Inventory{},
		complexSlotToInventoryMap: map[int]*ComplexInventoryMapEntry{},
		lastInventoryNetworkID:    ContainerIDFirst,
		currentWindowType:         WindowTypeContainer,
		clientSelectedHotbarSlot:  -1,
		nextItemStackID:           1,
	}
	m.containerOpenCallbacks = []ContainerOpenFunc{createContainerOpen}

	m.add(ContainerIDInventory, p.GetInventory())
	m.add(ContainerIDOffhand, p.GetOffHandInventory())
	m.add(ContainerIDArmor, p.GetArmorInventory())
	m.addComplex(UISlotCursor, p.GetCursorInventory())
	m.addComplex(UISlotCrafting2x2Input, p.GetCraftingGrid())

	listener := inventory.HeldItemIndexChangeListener(func(int) { m.SyncSelectedHotbarSlot() })
	m.heldItemIndexListener = &listener
	p.GetInventory().GetHeldItemIndexChangeListeners().Add(m.heldItemIndexListener)
	return m
}

// dispose drops the held item listener NewInventoryManager added (PHP relies on the GC for this).
func (m *InventoryManager) dispose() {
	m.player.GetInventory().GetHeldItemIndexChangeListeners().Remove(m.heldItemIndexListener)
}

func (m *InventoryManager) associateIDWithInventory(id int, inv inventory.Inventory) {
	m.networkIDToInventoryMap[id] = inv
}

func (m *InventoryManager) getNewWindowID() int {
	m.lastInventoryNetworkID = max(ContainerIDFirst, (m.lastInventoryNetworkID+1)%ContainerIDLast)
	return m.lastInventoryNetworkID
}

func (m *InventoryManager) track(inv inventory.Inventory, entry *InventoryManagerEntry) {
	if _, ok := m.inventories[inv]; ok {
		panic(fmt.Sprintf("Inventory %T is already tracked", inv))
	}
	m.inventories[inv] = entry
	m.order = append(m.order, inv)
}

func (m *InventoryManager) add(id int, inv inventory.Inventory) {
	m.track(inv, newInventoryManagerEntry(inv, nil))
	m.associateIDWithInventory(id, inv)
}

func (m *InventoryManager) addDynamic(inv inventory.Inventory) int {
	id := m.getNewWindowID()
	m.add(id, inv)
	return id
}

func (m *InventoryManager) addComplex(slotMap map[int]int, inv inventory.Inventory) {
	complexSlotMap := NewComplexInventoryMapEntry(inv, slotMap)
	m.track(inv, newInventoryManagerEntry(inv, complexSlotMap))
	for netSlot := range complexSlotMap.GetSlotMap() {
		m.complexSlotToInventoryMap[netSlot] = complexSlotMap
	}
}

func (m *InventoryManager) addComplexDynamic(slotMap map[int]int, inv inventory.Inventory) int {
	m.addComplex(slotMap, inv)
	id := m.getNewWindowID()
	m.associateIDWithInventory(id, inv)
	return id
}

func (m *InventoryManager) remove(id int) {
	inv := m.networkIDToInventoryMap[id]
	delete(m.networkIDToInventoryMap, id)
	if _, ok := m.GetWindowID(inv); !ok {
		delete(m.inventories, inv)
		m.order = slices.DeleteFunc(m.order, func(i inventory.Inventory) bool { return i == inv })
		for netSlot, entry := range m.complexSlotToInventoryMap {
			if entry.GetInventory() == inv {
				delete(m.complexSlotToInventoryMap, netSlot)
			}
		}
	}
}

// GetWindowID is a port of InventoryManager::getWindowId.
func (m *InventoryManager) GetWindowID(inv inventory.Inventory) (int, bool) {
	for id, other := range m.networkIDToInventoryMap {
		if other == inv {
			return id, true
		}
	}
	return 0, false
}

// GetCurrentWindowID is a port of InventoryManager::getCurrentWindowId.
func (m *InventoryManager) GetCurrentWindowID() int { return m.lastInventoryNetworkID }

// LocateWindowAndSlot is a port of InventoryManager::locateWindowAndSlot.
func (m *InventoryManager) LocateWindowAndSlot(windowID, netSlotID int) (inventory.Inventory, int, bool) {
	if windowID == ContainerIDUI {
		entry, ok := m.complexSlotToInventoryMap[netSlotID]
		if !ok {
			return nil, 0, false
		}
		inv := entry.GetInventory()
		coreSlotID, ok := entry.MapNetToCore(netSlotID)
		if ok && inv.SlotExists(coreSlotID) {
			return inv, coreSlotID, true
		}
		return nil, 0, false
	}
	if inv, ok := m.networkIDToInventoryMap[windowID]; ok && inv.SlotExists(netSlotID) {
		return inv, netSlotID, true
	}
	return nil, 0, false
}

func (m *InventoryManager) addPredictedSlotChangeInternal(inv inventory.Inventory, slot int, stack protocol.ItemStack) {
	if entry, ok := m.inventories[inv]; ok {
		entry.predictions[slot] = stack
	}
}

// AddPredictedSlotChange is a port of InventoryManager::addPredictedSlotChange.
func (m *InventoryManager) AddPredictedSlotChange(inv inventory.Inventory, slot int, it item.Item) {
	m.addPredictedSlotChangeInternal(inv, slot, convert.CoreItemStackToNet(it))
}

// AddRawPredictedSlotChanges is a port of InventoryManager::addRawPredictedSlotChanges.
func (m *InventoryManager) AddRawPredictedSlotChanges(networkInventoryActions []protocol.InventoryAction) error {
	for _, action := range networkInventoryActions {
		if action.SourceType != protocol.InventoryActionSourceContainer {
			continue
		}

		//legacy transactions should not modify or predict anything other than these inventories, since these are
		//the only ones accessible when not in-game (ItemStackRequest is used for everything else)
		windowID8, ok := action.WindowID.Value()
		if !ok {
			return &network.PacketHandlingError{Message: "Window ID should always be set for SOURCE_CONTAINER"}
		}
		windowID := int(uint8(windowID8))
		switch windowID {
		case ContainerIDInventory, ContainerIDOffhand, ContainerIDArmor:
		default:
			return &network.PacketHandlingError{Message: fmt.Sprintf("Legacy transactions cannot predict changes to inventory with ID %d", windowID)}
		}
		inv, slot, ok := m.LocateWindowAndSlot(windowID, int(action.InventorySlot))
		if !ok {
			continue
		}
		m.addPredictedSlotChangeInternal(inv, slot, action.NewItem.Stack)
	}
	return nil
}

// SetCurrentItemStackRequestID is a port of InventoryManager::setCurrentItemStackRequestId.
func (m *InventoryManager) SetCurrentItemStackRequestID(id *int32) { m.currentItemStackRequestID = id }

// openWindowDeferred is a port of InventoryManager::openWindowDeferred.
//
// When the server initiates a window close, it does so by sending a ContainerClose to the client, which causes the
// client to behave as if it initiated the close itself. It responds by sending a ContainerClose back to the server,
// which the server is then expected to respond to.
//
// Sending the client a new window before sending this final response creates buggy behaviour on the client, which
// is problematic when switching windows. Therefore, we defer sending any new windows until after the client
// responds to our window close instruction, so that we can complete the window handshake correctly.
func (m *InventoryManager) openWindowDeferred(fn func()) {
	if m.pendingCloseWindowID != nil {
		m.session.GetLogger().Debug(fmt.Sprintf("Deferring opening of new window, waiting for close ack of window %d", *m.pendingCloseWindowID))
		m.pendingOpenWindowCallback = fn
	} else {
		fn()
	}
}

// createComplexSlotMapping is a port of InventoryManager::createComplexSlotMapping.
func createComplexSlotMapping(inv inventory.Inventory) map[int]int {
	//TODO: make this dynamic so plugins can add mappings for stuff not implemented by PM
	switch inv.(type) {
	case *blockinventory.EnchantInventory:
		return UISlotEnchantingTable
	case *blockinventory.LoomInventory:
		return UISlotLoom
	case *blockinventory.StonecutterInventory:
		return map[int]int{UISlotStoneCutterInput: blockinventory.StonecutterSlotInput}
	case *blockinventory.CraftingTableInventory:
		return UISlotCrafting3x3Input
	case *blockinventory.CartographyTableInventory:
		return UISlotCartographyTable
	case *blockinventory.SmithingTableInventory:
		return UISlotSmithingTable
	}
	return nil
}

// OnCurrentWindowChange is a port of InventoryManager::onCurrentWindowChange.
func (m *InventoryManager) OnCurrentWindowChange(inv inventory.Inventory) {
	m.OnCurrentWindowRemove()

	m.openWindowDeferred(func() {
		var windowID int
		if slotMap := createComplexSlotMapping(inv); slotMap != nil {
			windowID = m.addComplexDynamic(slotMap, inv)
		} else {
			windowID = m.addDynamic(inv)
		}

		for _, callback := range m.containerOpenCallbacks {
			pks := callback(windowID, inv)
			if pks == nil {
				continue
			}
			windowType := WindowTypeContainer
			for _, pk := range pks {
				if open, ok := pk.(*packet.ContainerOpen); ok {
					//workaround useless bullshit in 1.21 - ContainerClose requires a type now for some reason
					windowType = int(int8(open.ContainerType))
				}
				m.session.SendDataPacket(pk)
			}
			m.currentWindowType = windowType
			m.SyncContents(inv)
			return
		}
		panic("Unsupported inventory type")
	})
}

// AddContainerOpenCallback is InventoryManager::getContainerOpenCallbacks()->add().
func (m *InventoryManager) AddContainerOpenCallback(callback ContainerOpenFunc) {
	m.containerOpenCallbacks = append(m.containerOpenCallbacks, callback)
}

// createContainerOpen is a port of InventoryManager::createContainerOpen.
func createContainerOpen(id int, inv inventory.Inventory) []packet.Packet {
	//TODO: we should be using some kind of tagging system to identify the types. Instanceof is flaky especially
	//if the class isn't final, not to mention being inflexible.
	blockInv, ok := inv.(blockinventory.BlockInventory)
	if !ok {
		return nil
	}
	holder := blockInv.GetHolder()
	windowType := WindowTypeContainer
	switch inv.(type) {
	case *blockinventory.LoomInventory:
		windowType = WindowTypeLoom
	case *blockinventory.EnchantInventory:
		windowType = WindowTypeEnchantment
	case *blockinventory.CraftingTableInventory:
		windowType = WindowTypeWorkbench
	case *blockinventory.StonecutterInventory:
		windowType = WindowTypeStonecutter
	case *blockinventory.CartographyTableInventory:
		windowType = WindowTypeCartography
	case *blockinventory.SmithingTableInventory:
		windowType = WindowTypeSmithingTable
	}
	return []packet.Packet{&packet.ContainerOpen{
		WindowID:                byte(id),
		ContainerType:           byte(windowType),
		ContainerPosition:       protocol.BlockPos{int32(holder.FloorX()), int32(holder.FloorY()), int32(holder.FloorZ())},
		ContainerEntityUniqueID: -1,
	}}
}

// OnClientOpenMainInventory is a port of InventoryManager::onClientOpenMainInventory.
func (m *InventoryManager) OnClientOpenMainInventory() {
	m.OnCurrentWindowRemove()

	m.openWindowDeferred(func() {
		windowID := m.getNewWindowID()
		m.associateIDWithInventory(windowID, m.player.GetInventory())
		m.currentWindowType = WindowTypeInventory

		m.session.SendDataPacket(&packet.ContainerOpen{
			WindowID:                byte(windowID),
			ContainerType:           byte(int8(m.currentWindowType)),
			ContainerEntityUniqueID: int64(m.player.GetID()),
		})
	})
}

// OnCurrentWindowRemove is a port of InventoryManager::onCurrentWindowRemove.
func (m *InventoryManager) OnCurrentWindowRemove() {
	if _, ok := m.networkIDToInventoryMap[m.lastInventoryNetworkID]; ok {
		m.remove(m.lastInventoryNetworkID)
		m.session.SendDataPacket(&packet.ContainerClose{WindowID: byte(m.lastInventoryNetworkID), ContainerType: byte(int8(m.currentWindowType)), ServerSide: true})
		if m.pendingCloseWindowID != nil {
			panic("We should not have opened a new window while a window was waiting to be closed")
		}
		id := m.lastInventoryNetworkID
		m.pendingCloseWindowID = &id
	}
}

// OnClientRemoveWindow is a port of InventoryManager::onClientRemoveWindow.
func (m *InventoryManager) OnClientRemoveWindow(id int) {
	if int8(id) == ContainerIDNone {
		//TODO: HACK! Since 1.21.100 (and probably earlier), the client will send -1 to close windows that it can't
		//view for some reason, e.g. if the chat window was already open. This is pretty awkward, since it means
		//that we can only assume it refers to the most recently sent window, and if we don't handle it,
		//InventoryManager will never get the green light to send subsequent windows, which breaks inventory UIs.
		//Fortunately, we already wait for close acks anyway, so the window ID is technically useless...?
		m.session.GetLogger().Debug(fmt.Sprintf("Client rejected opening of a window, assuming it was %d", m.lastInventoryNetworkID))
		id = m.lastInventoryNetworkID
	}
	if id == m.lastInventoryNetworkID {
		if _, ok := m.networkIDToInventoryMap[id]; ok && (m.pendingCloseWindowID == nil || id != *m.pendingCloseWindowID) {
			m.remove(id)
			m.player.RemoveCurrentWindow()
		}
	} else {
		m.session.GetLogger().Debug(fmt.Sprintf("Attempted to close inventory with network ID %d, but current is %d", id, m.lastInventoryNetworkID))
	}

	//Always send this, even if no window matches. If we told the client to close a window, it will behave as if it
	//initiated the close and expect an ack.
	m.session.SendDataPacket(&packet.ContainerClose{WindowID: byte(id), ContainerType: byte(int8(m.currentWindowType)), ServerSide: false})

	if m.pendingCloseWindowID != nil && *m.pendingCloseWindowID == id {
		m.pendingCloseWindowID = nil
		if m.pendingOpenWindowCallback != nil {
			m.session.GetLogger().Debug(fmt.Sprintf("Opening deferred window after close ack of window %d", id))
			callback := m.pendingOpenWindowCallback
			m.pendingOpenWindowCallback = nil
			callback()
		}
	}
}

// itemStacksEqual is a port of InventoryManager::itemStacksEqual (+ itemStackExtraDataEqual, which
// compares the decoded NBT and can-place-on/can-destroy lists; gophertunnel has already decoded
// them).
func itemStacksEqual(left, right protocol.ItemStack) bool {
	return left.NetworkID == right.NetworkID &&
		left.MetadataValue == right.MetadataValue &&
		left.BlockRuntimeID == right.BlockRuntimeID &&
		left.Count == right.Count &&
		slices.Equal(left.CanBePlacedOn, right.CanBePlacedOn) &&
		slices.Equal(left.CanBreak, right.CanBreak) &&
		(len(left.NBTData) == 0 && len(right.NBTData) == 0 || reflect.DeepEqual(left.NBTData, right.NBTData))
}

// OnSlotChange is a port of InventoryManager::onSlotChange.
func (m *InventoryManager) OnSlotChange(inv inventory.Inventory, slot int) {
	entry, ok := m.inventories[inv]
	if !ok {
		//this can happen when an inventory changed during InventoryCloseEvent, or when a temporary inventory
		//is cleared before removal.
		return
	}
	currentItem := convert.CoreItemStackToNet(inv.GetItem(slot))
	clientSideItem, predicted := entry.predictions[slot]
	if !predicted || !itemStacksEqual(currentItem, clientSideItem) {
		//no prediction or incorrect - do not associate this with the currently active itemstack request
		m.trackItemStack(entry, slot, currentItem, nil)
		entry.pendingSyncs[slot] = currentItem
	} else {
		//correctly predicted - associate the change with the currently active itemstack request
		m.trackItemStack(entry, slot, currentItem, m.currentItemStackRequestID)
	}

	delete(entry.predictions, slot)
}

// sendInventorySlotPackets is a port of InventoryManager::sendInventorySlotPackets.
func (m *InventoryManager) sendInventorySlotPackets(windowID, netSlot int, wrapper protocol.ItemInstance) {
	/*
	 * TODO: HACK!
	 * As of 1.20.12, the client ignores change of itemstackID in some cases when the old item == the new item.
	 * Notably, this happens with armor, offhand and enchanting tables, but not with main inventory.
	 * While we could track the items previously sent to the client, that's a waste of memory and would
	 * cost performance. Instead, clear the slot(s) first, then send the new item(s).
	 * The network cost of doing this is fortunately minimal, as an air itemstack is only 1 byte.
	 */
	if wrapper.StackNetworkID != 0 {
		m.session.SendDataPacket(&packet.InventorySlot{WindowID: uint32(windowID), Slot: uint32(netSlot)})
	}
	//now send the real contents
	m.session.SendDataPacket(&packet.InventorySlot{WindowID: uint32(windowID), Slot: uint32(netSlot), NewItem: wrapper})
}

// sendInventoryContentPackets is a port of InventoryManager::sendInventoryContentPackets.
func (m *InventoryManager) sendInventoryContentPackets(windowID int, wrappers []protocol.ItemInstance) {
	// See sendInventorySlotPackets: clear first, then send the real contents.
	m.session.SendDataPacket(&packet.InventoryContent{
		WindowID:  uint32(windowID),
		Content:   make([]protocol.ItemInstance, len(wrappers)),
		Container: protocol.FullContainerName{ContainerID: 0},
	})
	m.session.SendDataPacket(&packet.InventoryContent{
		WindowID:  uint32(windowID),
		Content:   wrappers,
		Container: protocol.FullContainerName{ContainerID: 0},
	})
}

// SyncSlot is a port of InventoryManager::syncSlot.
func (m *InventoryManager) SyncSlot(inv inventory.Inventory, slot int, stack protocol.ItemStack) {
	entry, ok := m.inventories[inv]
	if !ok {
		panic("Cannot sync an untracked inventory")
	}
	info, ok := entry.itemStackInfos[slot]
	if !ok {
		panic("Cannot sync an untracked inventory slot")
	}
	var windowID, netSlot int
	if entry.complexSlotMap != nil {
		windowID = ContainerIDUI
		netSlot, ok = entry.complexSlotMap.MapCoreToNet(slot)
		if !ok {
			panic("We already have an ItemStackInfo, so this should not be null")
		}
	} else {
		windowID, ok = m.GetWindowID(inv)
		if !ok {
			panic("We already have an ItemStackInfo, so this should not be null")
		}
		netSlot = slot
	}

	wrapper := protocol.ItemInstance{StackNetworkID: info.GetStackID(), Stack: stack}
	if windowID == ContainerIDOffhand {
		//TODO: HACK!
		//The client may sometimes ignore the InventorySlotPacket for the offhand slot.
		//This can cause a lot of problems (totems, arrows, and more...).
		//The workaround is to send an InventoryContentPacket instead
		//BDS (Bedrock Dedicated Server) also seems to work this way.
		m.sendInventoryContentPackets(windowID, []protocol.ItemInstance{wrapper})
	} else {
		m.sendInventorySlotPackets(windowID, netSlot, wrapper)
	}
	delete(entry.predictions, slot)
	delete(entry.pendingSyncs, slot)
}

// SyncContents is a port of InventoryManager::syncContents.
func (m *InventoryManager) SyncContents(inv inventory.Inventory) {
	entry, ok := m.inventories[inv]
	if !ok {
		//this can happen when an inventory changed during InventoryCloseEvent, or when a temporary inventory
		//is cleared before removal.
		return
	}
	var windowID int
	if entry.complexSlotMap != nil {
		windowID = ContainerIDUI
	} else if windowID, ok = m.GetWindowID(inv); !ok {
		return
	}
	entry.predictions = map[int]protocol.ItemStack{}
	entry.pendingSyncs = map[int]protocol.ItemStack{}
	contents := inv.GetContents(true)
	wrappers := make([]protocol.ItemInstance, inv.GetSize())
	for slot := 0; slot < inv.GetSize(); slot++ {
		stack := convert.CoreItemStackToNet(contents[slot])
		info := m.trackItemStack(entry, slot, stack, nil)
		wrappers[slot] = protocol.ItemInstance{StackNetworkID: info.GetStackID(), Stack: stack}
	}
	if entry.complexSlotMap != nil {
		for slot, wrapper := range wrappers {
			packetSlot, ok := entry.complexSlotMap.MapCoreToNet(slot)
			if !ok {
				continue
			}
			m.sendInventorySlotPackets(windowID, packetSlot, wrapper)
		}
	} else {
		m.sendInventoryContentPackets(windowID, wrappers)
	}
}

// SyncAll is a port of InventoryManager::syncAll.
func (m *InventoryManager) SyncAll() {
	for _, inv := range append([]inventory.Inventory(nil), m.order...) {
		m.SyncContents(inv)
	}
}

// RequestSyncAll is a port of InventoryManager::requestSyncAll.
func (m *InventoryManager) RequestSyncAll() { m.fullSyncRequested = true }

// SyncMismatchedPredictedSlotChanges is a port of InventoryManager::syncMismatchedPredictedSlotChanges.
func (m *InventoryManager) SyncMismatchedPredictedSlotChanges() {
	for _, inv := range m.order {
		entry := m.inventories[inv]
		for slot := range entry.predictions {
			if _, tracked := entry.itemStackInfos[slot]; !inv.SlotExists(slot) || !tracked {
				continue //TODO: size desync ???
			}

			//any prediction that still exists at this point is a slot that was predicted to change but didn't
			m.session.GetLogger().Debug(fmt.Sprintf("Detected prediction mismatch in inventory %T#%p slot %d", inv, inv, slot))
			entry.pendingSyncs[slot] = convert.CoreItemStackToNet(inv.GetItem(slot))
		}

		entry.predictions = map[int]protocol.ItemStack{}
	}
}

// FlushPendingUpdates is a port of InventoryManager::flushPendingUpdates.
func (m *InventoryManager) FlushPendingUpdates() {
	if m.fullSyncRequested {
		m.fullSyncRequested = false
		m.session.GetLogger().Debug(fmt.Sprintf("Full inventory sync requested, sending contents of %d inventories", len(m.inventories)))
		m.SyncAll()
		return
	}
	for _, inv := range append([]inventory.Inventory(nil), m.order...) {
		entry, ok := m.inventories[inv]
		if !ok || len(entry.pendingSyncs) == 0 {
			continue
		}
		slots := make([]int, 0, len(entry.pendingSyncs))
		for slot := range entry.pendingSyncs {
			slots = append(slots, slot)
		}
		slices.Sort(slots)
		names := make([]string, len(slots))
		for i, slot := range slots {
			names[i] = fmt.Sprint(slot)
		}
		m.session.GetLogger().Debug(fmt.Sprintf("Syncing slots %s in inventory %T#%p", strings.Join(names, ", "), inv, inv))
		for _, slot := range slots {
			m.SyncSlot(inv, slot, entry.pendingSyncs[slot])
		}
		entry.pendingSyncs = map[int]protocol.ItemStack{}
	}
}

// SyncData is a port of InventoryManager::syncData.
func (m *InventoryManager) SyncData(inv inventory.Inventory, propertyID, value int) {
	if windowID, ok := m.GetWindowID(inv); ok {
		m.session.SendDataPacket(&packet.ContainerSetData{WindowID: byte(windowID), Key: int32(propertyID), Value: int32(value)})
	}
}

// OnClientSelectHotbarSlot is a port of InventoryManager::onClientSelectHotbarSlot.
func (m *InventoryManager) OnClientSelectHotbarSlot(slot int) { m.clientSelectedHotbarSlot = slot }

// SyncSelectedHotbarSlot is a port of InventoryManager::syncSelectedHotbarSlot.
func (m *InventoryManager) SyncSelectedHotbarSlot() {
	playerInventory := m.player.GetInventory()
	selected := playerInventory.GetHeldItemIndex()
	if selected == m.clientSelectedHotbarSlot {
		return
	}
	entry, ok := m.inventories[playerInventory]
	if !ok {
		panic("Player inventory should always be tracked")
	}
	info, ok := entry.itemStackInfos[selected]
	if !ok {
		panic(fmt.Sprintf("Untracked player inventory slot %d", selected))
	}

	m.session.SendDataPacket(&packet.MobEquipment{
		EntityRuntimeID: uint64(m.player.GetID()),
		NewItem:         protocol.ItemInstance{StackNetworkID: info.GetStackID(), Stack: convert.CoreItemStackToNet(playerInventory.GetItemInHand())},
		InventorySlot:   byte(selected),
		HotBarSlot:      byte(selected),
		WindowID:        ContainerIDInventory,
	})
	m.clientSelectedHotbarSlot = selected
}

// SyncCreative is a port of InventoryManager::syncCreative.
func (m *InventoryManager) SyncCreative() {
	m.session.SendDataPacket(GetCreativeInventoryCache().BuildPacket(m.player.GetCreativeInventory(), m.session))
}

// GetEnchantingTableOptionIndex is a port of InventoryManager::getEnchantingTableOptionIndex.
// Enchanting options are never sent (EnchantingHelper isn't ported), so there are none.
func (m *InventoryManager) GetEnchantingTableOptionIndex(recipeID int) (int, bool) { return 0, false }

func (m *InventoryManager) newItemStackID() int32 {
	id := m.nextItemStackID
	m.nextItemStackID++
	return id
}

// GetItemStackInfo is a port of InventoryManager::getItemStackInfo.
func (m *InventoryManager) GetItemStackInfo(inv inventory.Inventory, slot int) *ItemStackInfo {
	if entry, ok := m.inventories[inv]; ok {
		return entry.itemStackInfos[slot]
	}
	return nil
}

func (m *InventoryManager) trackItemStack(entry *InventoryManagerEntry, slotID int, stack protocol.ItemStack, requestID *int32) *ItemStackInfo {
	//TODO: ItemStack->isNull() would be nice to have here
	stackID := int32(0)
	if stack.NetworkID != 0 {
		stackID = m.newItemStackID()
	}
	info := NewItemStackInfo(requestID, stackID)
	entry.itemStackInfos[slotID] = info
	return info
}

var _ player.InventoryManager = (*InventoryManager)(nil)
