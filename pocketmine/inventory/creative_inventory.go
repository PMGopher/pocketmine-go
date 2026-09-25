package inventory

import (
	"sync"

	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/utils"
)

// CreativeCategory is a port of pocketmine\inventory\CreativeCategory.
type CreativeCategory int

const (
	CreativeCategoryConstruction CreativeCategory = iota
	CreativeCategoryNature
	CreativeCategoryEquipment
	CreativeCategoryItems
)

// CreativeGroup is a port of pocketmine\inventory\CreativeGroup: a named, collapsible group of
// items in the creative inventory, shown with an icon.
type CreativeGroup struct {
	name any // string or *lang.Translatable
	icon item.Item
}

// NewCreativeGroup is a port of CreativeGroup::__construct (panicking on an empty name, like PHP's
// InvalidArgumentException).
func NewCreativeGroup(name any, icon item.Item) *CreativeGroup {
	length := 0
	switch n := name.(type) {
	case *lang.Translatable:
		length = len(n.Text())
	case string:
		length = len(n)
	}
	if length == 0 {
		panic("Creative group name cannot be empty")
	}
	return &CreativeGroup{name: name, icon: icon}
}

// GetName returns the group name (a string or *lang.Translatable).
func (g *CreativeGroup) GetName() any { return g.name }

func (g *CreativeGroup) GetIcon() item.Item { return g.icon.Clone() }

// CreativeInventoryEntry is a port of pocketmine\inventory\CreativeInventoryEntry.
type CreativeInventoryEntry struct {
	item     item.Item
	category CreativeCategory
	group    *CreativeGroup
}

func NewCreativeInventoryEntry(it item.Item, category CreativeCategory, group *CreativeGroup) *CreativeInventoryEntry {
	return &CreativeInventoryEntry{item: it.Clone(), category: category, group: group}
}

func (e *CreativeInventoryEntry) GetItem() item.Item { return e.item.Clone() }

func (e *CreativeInventoryEntry) GetCategory() CreativeCategory { return e.category }

// GetGroup returns the entry's group, or nil.
func (e *CreativeInventoryEntry) GetGroup() *CreativeGroup { return e.group }

// MatchesItem is a port of CreativeInventoryEntry::matchesItem (damage checked, NBT not).
func (e *CreativeInventoryEntry) MatchesItem(it item.Item) bool {
	return it.Equals(e.item, false)
}

// CreativeInventory is a port of pocketmine\inventory\CreativeInventory: the items offered in
// the creative inventory menu, shared by every player unless a plugin gives one a custom copy.
type CreativeInventory struct {
	creative []*CreativeInventoryEntry
	// removed marks entries removed with Remove: PHP unsets them, keeping the other entries'
	// indexes (network creative item IDs are these indexes).
	removed                 map[int]bool
	contentChangedCallbacks *utils.ObjectSet[*func()]
}

// NewCreativeInventory creates an empty creative inventory.
func NewCreativeInventory() *CreativeInventory {
	return &CreativeInventory{removed: map[int]bool{}, contentChangedCallbacks: utils.NewObjectSet[*func()]()}
}

// CreativeInventoryLoader fills the default creative inventory with the vanilla creative items
// (PHP's constructor, reading BedrockData's creative item lists). The item data lives in packages
// above this one, so the network/mcpe/convert package installs it.
var CreativeInventoryLoader func(inv *CreativeInventory)

var (
	creativeInventoryOnce     sync.Once
	creativeInventoryInstance *CreativeInventory
)

// GetCreativeInventory is CreativeInventory::getInstance.
func GetCreativeInventory() *CreativeInventory {
	creativeInventoryOnce.Do(func() {
		creativeInventoryInstance = NewCreativeInventory()
		if CreativeInventoryLoader != nil {
			CreativeInventoryLoader(creativeInventoryInstance)
		}
	})
	return creativeInventoryInstance
}

// Clone returns a copy that can be customised for one player (Player::setCreativeInventory).
func (c *CreativeInventory) Clone() *CreativeInventory {
	cp := NewCreativeInventory()
	cp.creative = append(cp.creative, c.creative...)
	for k := range c.removed {
		cp.removed[k] = true
	}
	return cp
}

// Clear removes all the creative items.
func (c *CreativeInventory) Clear() {
	c.creative = nil
	c.removed = map[int]bool{}
	c.onContentChange()
}

// GetAll is a port of CreativeInventory::getAll: index => item.
func (c *CreativeInventory) GetAll() map[int]item.Item {
	result := make(map[int]item.Item, len(c.creative))
	for i, e := range c.creative {
		if !c.removed[i] {
			result[i] = e.GetItem()
		}
	}
	return result
}

// GetAllEntries is a port of CreativeInventory::getAllEntries, in index order; the second result
// holds each entry's index.
func (c *CreativeInventory) GetAllEntries() ([]*CreativeInventoryEntry, []int) {
	var entries []*CreativeInventoryEntry
	var indexes []int
	for i, e := range c.creative {
		if !c.removed[i] {
			entries = append(entries, e)
			indexes = append(indexes, i)
		}
	}
	return entries, indexes
}

// GetItem returns the item at the given index, or nil.
func (c *CreativeInventory) GetItem(index int) item.Item {
	if e := c.GetEntry(index); e != nil {
		return e.GetItem()
	}
	return nil
}

// GetEntry returns the entry at the given index, or nil.
func (c *CreativeInventory) GetEntry(index int) *CreativeInventoryEntry {
	if index < 0 || index >= len(c.creative) || c.removed[index] {
		return nil
	}
	return c.creative[index]
}

// GetItemIndex returns the index of the entry matching it, or -1.
func (c *CreativeInventory) GetItemIndex(it item.Item) int {
	for i, e := range c.creative {
		if !c.removed[i] && e.MatchesItem(it) {
			return i
		}
	}
	return -1
}

// Add adds an item to the creative menu. Note: Players will not see this until they respawn or
// rejoin.
func (c *CreativeInventory) Add(it item.Item, category CreativeCategory, group *CreativeGroup) {
	c.creative = append(c.creative, NewCreativeInventoryEntry(it, category, group))
	c.onContentChange()
}

// Remove removes an item from the creative menu.
func (c *CreativeInventory) Remove(it item.Item) {
	if index := c.GetItemIndex(it); index != -1 {
		c.removed[index] = true
		c.onContentChange()
	}
}

func (c *CreativeInventory) Contains(it item.Item) bool { return c.GetItemIndex(it) != -1 }

// GetContentChangedCallbacks returns the callbacks called whenever the contents change.
func (c *CreativeInventory) GetContentChangedCallbacks() *utils.ObjectSet[*func()] {
	return c.contentChangedCallbacks
}

func (c *CreativeInventory) onContentChange() {
	for cb := range c.contentChangedCallbacks.All() {
		(*cb)()
	}
}
