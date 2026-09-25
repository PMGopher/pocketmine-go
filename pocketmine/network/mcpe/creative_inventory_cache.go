package mcpe

import (
	"fmt"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network/mcpe/convert"
)

// CreativeInventoryCacheEntry is a port of pocketmine\network\mcpe\cache\CreativeInventoryCacheEntry.
type CreativeInventoryCacheEntry struct {
	categories []inventory.CreativeCategory
	// groups holds nil for an anonymous group.
	groups []*inventory.CreativeGroup
	items  []protocol.CreativeItem
}

// CreativeInventoryCache is a port of pocketmine\network\mcpe\cache\CreativeInventoryCache.
type CreativeInventoryCache struct {
	mu     sync.Mutex
	caches map[*inventory.CreativeInventory]*CreativeInventoryCacheEntry
}

var creativeInventoryCache = &CreativeInventoryCache{caches: map[*inventory.CreativeInventory]*CreativeInventoryCacheEntry{}}

// GetCreativeInventoryCache is CreativeInventoryCache::getInstance.
func GetCreativeInventoryCache() *CreativeInventoryCache { return creativeInventoryCache }

func (c *CreativeInventoryCache) getCacheEntry(inv *inventory.CreativeInventory) *CreativeInventoryCacheEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.caches[inv]; ok {
		return entry
	}
	// PHP also drops the entry from the inventory's destructor callbacks; Go has no destructors,
	// and per-player creative inventories are clones that stop changing once the player leaves.
	invalidate := func() {
		c.mu.Lock()
		delete(c.caches, inv)
		c.mu.Unlock()
	}
	inv.GetContentChangedCallbacks().Add(&invalidate)
	entry := c.buildCacheEntry(inv)
	c.caches[inv] = entry
	return entry
}

// buildCacheEntry is a port of CreativeInventoryCache::buildCacheEntry.
func (c *CreativeInventoryCache) buildCacheEntry(inv *inventory.CreativeInventory) *CreativeInventoryCacheEntry {
	entry := &CreativeInventoryCacheEntry{}

	type groupKey struct {
		category inventory.CreativeCategory
		group    *inventory.CreativeGroup // nil is PHP's PHP_INT_MIN anonymous group
	}
	nextIndex := 0
	groupIndexes := map[groupKey]int{}

	entries, indexes := inv.GetAllEntries()
	itemGroupIndexes := make([]int, len(entries))
	for k, e := range entries {
		group := e.GetGroup()
		category := e.GetCategory()
		if group != nil {
			delete(groupIndexes, groupKey{category, nil}) //start a new anonymous group for this category
		}

		//group object may be reused by multiple categories
		key := groupKey{category, group}
		if _, ok := groupIndexes[key]; !ok {
			groupIndexes[key] = nextIndex
			nextIndex++
			entry.categories = append(entry.categories, category)
			entry.groups = append(entry.groups, group)
		}
		itemGroupIndexes[k] = groupIndexes[key]
	}

	//creative inventory may have holes if items were unregistered - ensure network IDs used are always consistent
	for k, e := range entries {
		entry.items = append(entry.items, protocol.CreativeItem{
			CreativeItemNetworkID: uint32(indexes[k]),
			Item:                  convert.CoreItemStackToNet(e.GetItem()),
			GroupIndex:            uint32(itemGroupIndexes[k]),
		})
	}
	return entry
}

// BuildPacket is a port of CreativeInventoryCache::buildPacket.
func (c *CreativeInventoryCache) BuildPacket(inv *inventory.CreativeInventory, session *NetworkSession) *packet.CreativeContent {
	p := session.GetPlayer()
	if p == nil {
		panic("Cannot prepare creative data for a session without a player")
	}
	language := p.GetLanguage()
	forceLanguage := session.server.IsLanguageForced()
	cachedEntry := c.getCacheEntry(inv)
	translate := func(name any) string {
		switch t := name.(type) {
		case *lang.Translatable:
			if !forceLanguage {
				message, _ := session.PrepareClientTranslatableMessage(t)
				return message
			}
			return language.Translate(t)
		default:
			return fmt.Sprint(name)
		}
	}

	groupEntries := make([]protocol.CreativeGroup, 0, len(cachedEntry.categories))
	for index, category := range cachedEntry.categories {
		group := cachedEntry.groups[index]
		categoryID := convert.ProtocolCreativeCategory(category)
		if group == nil {
			groupEntries = append(groupEntries, protocol.CreativeGroup{Category: categoryID})
			continue
		}
		groupIcon := group.GetIcon()
		//TODO: HACK! In 1.21.60, Workaround glitchy behaviour when an item is used as an icon for a group it
		//doesn't belong to. Without this hack, both instances of the item will show a +, but neither of them
		//will actually expand the group work correctly.
		groupIcon.GetNamedTag().SetInt("___GroupBugWorkaround___", nbt.IntTag(index))
		groupEntries = append(groupEntries, protocol.CreativeGroup{
			Category: categoryID,
			Name:     translate(group.GetName()),
			Icon:     convert.CoreItemStackToNet(groupIcon),
		})
	}
	return &packet.CreativeContent{Groups: groupEntries, Items: cachedEntry.items}
}
