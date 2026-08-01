// Package inventory is a port of pocketmine\inventory.
package inventory

import (
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/utils"
)

const MaxStack = 64

// Player is the minimal surface Inventory needs from a player — just enough identity to track
// viewers. Declared locally so the future player package satisfies it automatically with no
// import needed here, the same forward-compatible-local-interface pattern used for block.Item,
// block.Player, etc. elsewhere in this port.
type Player interface {
	// marker only — onOpen/onClose/getViewers just need identity (Go map/slice equality on the
	// concrete pointer type serves the same purpose as PHP's spl_object_id()).
}

// InventoryListener is a port of pocketmine\inventory\InventoryListener.
type InventoryListener interface {
	OnSlotChange(inv Inventory, index int, before item.Item)
	OnContentChange(inv Inventory, before map[int]item.Item)
}

// Inventory is a port of pocketmine\inventory\Inventory, using the same self-dispatch pattern as
// block.Behavior/item.Item: concrete inventory types embed BaseInventory, call Init(self) once
// their own storage is ready, and implement the handful of methods BaseInventory has no sensible
// default for (GetSize/GetItem/internalSetItem/internalSetContents/GetContents).
//
// Not ported: getSlotValidators (SlotValidatedInventory - needs the unported transaction/action/
// validator package). getViewers/onOpen/onClose keep working (just identity tracking), but
// onSlotChange/onContentChange's network-sync calls to viewer.getNetworkSession().getInvManager()
// aren't ported (needs the unported network/player packages) - listeners still fire correctly.
type Inventory interface {
	GetSize() int
	GetMaxStackSize() int
	SetMaxStackSize(size int)
	GetItem(index int) item.Item
	SetItem(index int, it item.Item)
	GetContents(includeEmpty bool) map[int]item.Item
	SetContents(items map[int]item.Item)
	AddItem(slots ...item.Item) []item.Item
	CanAddItem(it item.Item) bool
	GetAddableItemQuantity(it item.Item) int
	Contains(it item.Item) bool
	All(it item.Item) map[int]item.Item
	First(it item.Item, exact bool) int
	FirstEmpty() int
	IsSlotEmpty(index int) bool
	Remove(it item.Item)
	RemoveItem(slots ...item.Item) []item.Item
	Clear(index int)
	ClearAll()
	Swap(slot1, slot2 int)
	GetViewers() []Player
	OnOpen(who Player)
	OnClose(who Player)
	SlotExists(slot int) bool
	GetListeners() *utils.ObjectSet[InventoryListener]
}

// inventoryStorage is what a concrete type (e.g. SimpleInventory) must provide - the narrow
// self-dispatch surface BaseInventory's default method bodies reach through b.self, the same
// "*Shaper" pattern used throughout block/ and item/.
type inventoryStorage interface {
	GetSize() int
	GetItem(index int) item.Item
	GetContents(includeEmpty bool) map[int]item.Item
	internalSetItem(index int, it item.Item)
	internalSetContents(items map[int]item.Item)
}

// BaseInventory is a port of pocketmine\inventory\BaseInventory.
type BaseInventory struct {
	self inventoryStorage

	maxStackSize int
	viewers      []Player
	listeners    *utils.ObjectSet[InventoryListener]
}

// Init finishes constructing b, given self (the concrete inventory type embedding this
// BaseInventory). Must be called exactly once.
func (b *BaseInventory) Init(self inventoryStorage) {
	b.self = self
	b.maxStackSize = MaxStack
	b.listeners = utils.NewObjectSet[InventoryListener]()
}

// newEmptyItem returns a fresh always-empty Item, standing in for the PHP original's
// VanillaItems::AIR() sentinel (the item registry isn't ported, so a real Air item can't be
// constructed) - every place PHP returns/compares against Air uses this instead. Count is forced
// to 0 so IsNull() is true and Equals() against anything real is false, matching Air's role.
func newEmptyItem() item.Item {
	e := &emptyItem{}
	e.Init(e, item.NewItemIdentifier(0), "Air")
	e.SetCount(0)
	return e
}

type emptyItem struct {
	item.ItemBase
}

// Clone returns a fresh empty item rather than copying self - emptyItem carries no state beyond
// being empty, so there's nothing to preserve across the clone.
func (e *emptyItem) Clone() item.Item { return newEmptyItem() }

func (b *BaseInventory) GetMaxStackSize() int { return b.maxStackSize }

func (b *BaseInventory) SetMaxStackSize(size int) { b.maxStackSize = size }

// SetItem is a port of BaseInventory::setItem.
func (b *BaseInventory) SetItem(index int, it item.Item) {
	if it.IsNull() {
		it = newEmptyItem()
	} else {
		it = it.Clone()
	}

	oldItem := b.self.GetItem(index)
	b.self.internalSetItem(index, it)
	b.onSlotChange(index, oldItem)
}

// SetContents is a port of BaseInventory::setContents.
func (b *BaseInventory) SetContents(items map[int]item.Item) {
	if len(items) > b.self.GetSize() {
		trimmed := make(map[int]item.Item, b.self.GetSize())
		for i, it := range items {
			if i >= 0 && i < b.self.GetSize() {
				trimmed[i] = it
			}
		}
		items = trimmed
	}

	oldContents := b.self.GetContents(true)
	b.self.internalSetContents(items)
	b.onContentChange(oldContents)
}

// getMatchingItemCount is a port of BaseInventory::getMatchingItemCount - the slow-but-correct
// default; SimpleInventory overrides it to avoid GetItem's clone-per-call cost, matching the PHP
// original's own override for the same reason.
func (b *BaseInventory) getMatchingItemCount(slot int, test item.Item, checkTags bool) int {
	it := b.self.GetItem(slot)
	if it.Equals(test, checkTags) {
		return it.GetCount()
	}
	return 0
}

func (b *BaseInventory) Contains(it item.Item) bool {
	count := max(1, it.GetCount())
	checkTags := it.HasNamedTag()
	for i, size := 0, b.self.GetSize(); i < size; i++ {
		slotCount := b.matchingItemCount(i, it, checkTags)
		if slotCount > 0 {
			count -= slotCount
			if count <= 0 {
				return true
			}
		}
	}
	return false
}

func (b *BaseInventory) All(it item.Item) map[int]item.Item {
	slots := map[int]item.Item{}
	checkTags := it.HasNamedTag()
	for i, size := 0, b.self.GetSize(); i < size; i++ {
		if b.matchingItemCount(i, it, checkTags) > 0 {
			slots[i] = b.self.GetItem(i)
		}
	}
	return slots
}

func (b *BaseInventory) First(it item.Item, exact bool) int {
	count := it.GetCount()
	if !exact {
		count = max(1, count)
	}
	checkTags := exact || it.HasNamedTag()

	for i, size := 0, b.self.GetSize(); i < size; i++ {
		slotCount := b.matchingItemCount(i, it, checkTags)
		if slotCount > 0 && (slotCount == count || (!exact && slotCount > count)) {
			return i
		}
	}
	return -1
}

func (b *BaseInventory) FirstEmpty() int {
	for i, size := 0, b.self.GetSize(); i < size; i++ {
		if b.isSlotEmpty(i) {
			return i
		}
	}
	return -1
}

// IsSlotEmpty is a port of BaseInventory::isSlotEmpty - the slow-but-correct default;
// SimpleInventory overrides it for the same reason as getMatchingItemCount.
func (b *BaseInventory) IsSlotEmpty(index int) bool {
	return b.self.GetItem(index).IsNull()
}

func (b *BaseInventory) CanAddItem(it item.Item) bool {
	return b.GetAddableItemQuantity(it) == it.GetCount()
}

func (b *BaseInventory) GetAddableItemQuantity(it item.Item) int {
	count := it.GetCount()
	maxStackSize := min(b.maxStackSize, it.GetMaxStackSize())

	for i, size := 0, b.self.GetSize(); i < size; i++ {
		if b.isSlotEmpty(i) {
			count -= maxStackSize
		} else {
			slotCount := b.matchingItemCount(i, it, true)
			if diff := maxStackSize - slotCount; slotCount > 0 && diff > 0 {
				count -= diff
			}
		}
		if count <= 0 {
			return it.GetCount()
		}
	}
	return it.GetCount() - count
}

// AddItem is a port of BaseInventory::addItem.
func (b *BaseInventory) AddItem(slots ...item.Item) []item.Item {
	var itemSlots []item.Item
	for _, slot := range slots {
		if !slot.IsNull() {
			itemSlots = append(itemSlots, slot.Clone())
		}
	}

	var returnSlots []item.Item
	for _, it := range itemSlots {
		leftover := b.internalAddItem(it)
		if !leftover.IsNull() {
			returnSlots = append(returnSlots, leftover)
		}
	}
	return returnSlots
}

func (b *BaseInventory) internalAddItem(newItem item.Item) item.Item {
	var emptySlots []int
	maxStackSize := min(b.maxStackSize, newItem.GetMaxStackSize())

	for i, size := 0, b.self.GetSize(); i < size; i++ {
		if b.isSlotEmpty(i) {
			emptySlots = append(emptySlots, i)
			continue
		}
		slotCount := b.matchingItemCount(i, newItem, true)
		if slotCount == 0 {
			continue
		}
		if slotCount < maxStackSize {
			amount := min(maxStackSize-slotCount, newItem.GetCount())
			if amount > 0 {
				newItem.SetCount(newItem.GetCount() - amount)
				slotItem := b.self.GetItem(i)
				slotItem.SetCount(slotItem.GetCount() + amount)
				b.SetItem(i, slotItem)
				if newItem.GetCount() <= 0 {
					return newItem
				}
			}
		}
	}

	for _, slotIndex := range emptySlots {
		amount := min(maxStackSize, newItem.GetCount())
		newItem.SetCount(newItem.GetCount() - amount)
		slotItem := newItem.Clone()
		slotItem.SetCount(amount)
		b.SetItem(slotIndex, slotItem)
		if newItem.GetCount() <= 0 {
			return newItem
		}
	}

	return newItem
}

func (b *BaseInventory) Remove(it item.Item) {
	checkTags := it.HasNamedTag()
	for i, size := 0, b.self.GetSize(); i < size; i++ {
		if b.matchingItemCount(i, it, checkTags) > 0 {
			b.Clear(i)
		}
	}
}

// RemoveItem is a port of BaseInventory::removeItem.
func (b *BaseInventory) RemoveItem(slots ...item.Item) []item.Item {
	type indexed struct {
		idx  int
		item item.Item
	}
	var searchItems []indexed
	for _, slot := range slots {
		if !slot.IsNull() {
			searchItems = append(searchItems, indexed{len(searchItems), slot.Clone()})
		}
	}

	for i, size := 0, b.self.GetSize(); i < size && len(searchItems) > 0; i++ {
		if b.isSlotEmpty(i) {
			continue
		}
		remaining := searchItems[:0]
		for _, search := range searchItems {
			slotCount := b.matchingItemCount(i, search.item, search.item.HasNamedTag())
			if slotCount > 0 {
				amount := min(slotCount, search.item.GetCount())
				search.item.SetCount(search.item.GetCount() - amount)

				slotItem := b.self.GetItem(i)
				slotItem.SetCount(slotItem.GetCount() - amount)
				b.SetItem(i, slotItem)
			}
			if search.item.GetCount() > 0 {
				remaining = append(remaining, search)
			}
		}
		searchItems = remaining
	}

	result := make([]item.Item, len(searchItems))
	for i, s := range searchItems {
		result[i] = s.item
	}
	return result
}

func (b *BaseInventory) Clear(index int) { b.SetItem(index, newEmptyItem()) }

func (b *BaseInventory) ClearAll() { b.SetContents(nil) }

func (b *BaseInventory) Swap(slot1, slot2 int) {
	i1 := b.self.GetItem(slot1)
	i2 := b.self.GetItem(slot2)
	b.SetItem(slot1, i2)
	b.SetItem(slot2, i1)
}

func (b *BaseInventory) GetViewers() []Player { return b.viewers }

// RemoveAllViewers is a port of BaseInventory::removeAllViewers, minus the
// viewer.getCurrentWindow()/removeCurrentWindow() calls - Player is a marker-only interface here
// (see its doc comment), so it doesn't expose a "current window" concept to check against yet.
func (b *BaseInventory) RemoveAllViewers() { b.viewers = nil }

func (b *BaseInventory) OnOpen(who Player) {
	for _, v := range b.viewers {
		if v == who {
			return
		}
	}
	b.viewers = append(b.viewers, who)
}

func (b *BaseInventory) OnClose(who Player) {
	for i, v := range b.viewers {
		if v == who {
			b.viewers = append(b.viewers[:i], b.viewers[i+1:]...)
			return
		}
	}
}

// onSlotChange is a port of BaseInventory::onSlotChange, minus the viewer network-sync loop (see
// the Inventory interface's doc comment).
func (b *BaseInventory) onSlotChange(index int, before item.Item) {
	for l := range b.listeners.All() {
		l.OnSlotChange(b.self.(Inventory), index, before)
	}
}

// onContentChange is a port of BaseInventory::onContentChange, minus the viewer network-sync loop.
func (b *BaseInventory) onContentChange(itemsBefore map[int]item.Item) {
	for l := range b.listeners.All() {
		l.OnContentChange(b.self.(Inventory), itemsBefore)
	}
}

func (b *BaseInventory) SlotExists(slot int) bool { return slot >= 0 && slot < b.self.GetSize() }

func (b *BaseInventory) GetListeners() *utils.ObjectSet[InventoryListener] { return b.listeners }

// matchingItemCount and isSlotEmpty reach getMatchingItemCount/IsSlotEmpty through self rather
// than calling BaseInventory's own methods directly - PHP's $this-> calls resolve to the
// most-derived override automatically, but Go's embedding doesn't, so every internal caller in
// this file goes through these two helpers instead of b.getMatchingItemCount/b.IsSlotEmpty
// directly. SimpleInventory overrides both for performance (avoiding GetItem's clone-per-call
// cost), matching the PHP original's own overrides for the same reason.
func (b *BaseInventory) matchingItemCount(slot int, test item.Item, checkTags bool) int {
	if m, ok := b.self.(interface {
		getMatchingItemCount(slot int, test item.Item, checkTags bool) int
	}); ok {
		return m.getMatchingItemCount(slot, test, checkTags)
	}
	return b.getMatchingItemCount(slot, test, checkTags)
}

func (b *BaseInventory) isSlotEmpty(index int) bool {
	if s, ok := b.self.(interface{ IsSlotEmpty(index int) bool }); ok {
		return s.IsSlotEmpty(index)
	}
	return b.IsSlotEmpty(index)
}
