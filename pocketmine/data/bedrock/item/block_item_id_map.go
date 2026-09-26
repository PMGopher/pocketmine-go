package bedrockitem

import (
	"strings"
	"sync"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/item"
)

// BlockItemIdMap is a port of pocketmine\data\bedrock\item\BlockItemIdMap: the item ID of each
// block whose item ID differs from its block ID.
//
// PocketMine-MP loads BedrockData's block_id_to_item_id_map.json (every block ID with an item), which
// isn't vendored. The same table is derived from the vendored palette and item list: a block's item
// has the block's ID, except when a non-block item already uses it (e.g. "minecraft:bed"), then the
// block item is "minecraft:item.<name>". Block items have runtime IDs below 256 in the item list,
// and IDs ItemSerializerDeserializerRegistrar maps (e.g. hanging signs) are dedicated items.
type BlockItemIdMap struct {
	blockToItemId map[string]string
	itemToBlockId map[string]string
}

var (
	blockItemIdMap     *BlockItemIdMap
	blockItemIdMapOnce sync.Once
)

// GetBlockItemIdMap is BlockItemIdMap::getInstance().
func GetBlockItemIdMap() *BlockItemIdMap {
	blockItemIdMapOnce.Do(func() {
		m := &BlockItemIdMap{blockToItemId: map[string]string{}, itemToBlockId: map[string]string{}}
		itemIDs := map[string]int32{}
		for _, entry := range bedrock.ItemTypes() {
			itemIDs[entry.Name] = entry.RuntimeID
		}
		//item IDs that ItemSerializerDeserializerRegistrar maps are dedicated (non-block) items
		dedicated := &ItemDeserializer{deserializers: map[string]func(SavedItemData) item.Item{}}
		NewItemSerializerDeserializerRegistrar(dedicated, nil)
		for _, state := range bedrock.BlockStates() {
			blockID := state.Name
			if _, done := m.blockToItemId[blockID]; done {
				continue
			}
			itemID := "minecraft:item." + strings.TrimPrefix(blockID, "minecraft:")
			if _, ok := itemIDs[itemID]; !ok {
				itemID = blockID
				runtimeID, ok := itemIDs[itemID]
				//block items have runtime IDs below 256 (legacy block IDs, or negative for newer blocks)
				if !ok || runtimeID >= 256 {
					continue //no item for this block
				}
				if _, isDedicated := dedicated.deserializers[itemID]; isDedicated {
					continue
				}
			}
			m.blockToItemId[blockID] = itemID
			m.itemToBlockId[itemID] = blockID
		}
		blockItemIdMap = m
	})
	return blockItemIdMap
}

// LookupItemID is a port of BlockItemIdMap::lookupItemId.
func (m *BlockItemIdMap) LookupItemID(blockID string) (string, bool) {
	id, ok := m.blockToItemId[blockID]
	return id, ok
}

// LookupBlockID is a port of BlockItemIdMap::lookupBlockId.
func (m *BlockItemIdMap) LookupBlockID(itemID string) (string, bool) {
	id, ok := m.itemToBlockId[itemID]
	return id, ok
}
