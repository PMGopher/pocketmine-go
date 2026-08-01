package inventory

import "pocketmine-go/pocketmine/item"

// SimpleInventory is a port of pocketmine\inventory\SimpleInventory: a complete, array-backed
// implementation of Inventory. Empty slots are stored as nil (matching the PHP original's
// SplFixedArray-of-nullable-Item internal representation), and GetItem synthesizes a fresh empty
// item (see newEmptyItem's doc comment) for nil slots the same way PHP synthesizes
// VanillaItems::AIR().
type SimpleInventory struct {
	BaseInventory

	slots []item.Item
}

func NewSimpleInventory(size int) *SimpleInventory {
	s := &SimpleInventory{slots: make([]item.Item, size)}
	s.Init(s)
	return s
}

func (s *SimpleInventory) GetSize() int { return len(s.slots) }

func (s *SimpleInventory) GetItem(index int) item.Item {
	if index < 0 || index >= len(s.slots) || s.slots[index] == nil {
		return newEmptyItem()
	}
	return s.slots[index].Clone()
}

func (s *SimpleInventory) internalSetItem(index int, it item.Item) {
	if index < 0 || index >= len(s.slots) {
		return
	}
	if it.IsNull() {
		s.slots[index] = nil
	} else {
		s.slots[index] = it
	}
}

func (s *SimpleInventory) GetContents(includeEmpty bool) map[int]item.Item {
	contents := map[int]item.Item{}
	for i, slot := range s.slots {
		if slot != nil {
			contents[i] = slot.Clone()
		} else if includeEmpty {
			contents[i] = newEmptyItem()
		}
	}
	return contents
}

func (s *SimpleInventory) internalSetContents(items map[int]item.Item) {
	for i := range s.slots {
		it, ok := items[i]
		if !ok || it.IsNull() {
			s.slots[i] = nil
		} else {
			s.slots[i] = it.Clone()
		}
	}
}

// getMatchingItemCount overrides BaseInventory's slow default, matching the PHP original's own
// override (avoids GetItem's clone-per-call cost).
func (s *SimpleInventory) getMatchingItemCount(slot int, test item.Item, checkTags bool) int {
	if slot < 0 || slot >= len(s.slots) || s.slots[slot] == nil {
		return 0
	}
	slotItem := s.slots[slot]
	if slotItem.Equals(test, checkTags) {
		return slotItem.GetCount()
	}
	return 0
}

// IsSlotEmpty overrides BaseInventory's slow default, matching the PHP original's own override.
func (s *SimpleInventory) IsSlotEmpty(index int) bool {
	return index < 0 || index >= len(s.slots) || s.slots[index] == nil || s.slots[index].IsNull()
}
