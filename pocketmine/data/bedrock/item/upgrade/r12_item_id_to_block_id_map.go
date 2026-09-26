package itemupgrade

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// R12ItemIdToBlockIdMap is a port of pocketmine\data\bedrock\item\upgrade\R12ItemIdToBlockIdMap:
// the 1.12 blockitem IDs and the block IDs they place.
type R12ItemIdToBlockIdMap struct {
	itemToBlock map[string]string
	blockToItem map[string]string
}

// NewR12ItemIdToBlockIdMap is a port of R12ItemIdToBlockIdMap::__construct.
func NewR12ItemIdToBlockIdMap(itemToBlock map[string]string) *R12ItemIdToBlockIdMap {
	m := &R12ItemIdToBlockIdMap{itemToBlock: map[string]string{}, blockToItem: map[string]string{}}
	for itemID, blockID := range itemToBlock {
		m.itemToBlock[strings.ToLower(itemID)] = blockID
		m.blockToItem[strings.ToLower(blockID)] = itemID
	}
	return m
}

var (
	r12MapOnce sync.Once
	r12Map     *R12ItemIdToBlockIdMap
)

// GetR12ItemIdToBlockIdMap is a port of R12ItemIdToBlockIdMap::getInstance():
// 1.12.0_item_id_to_block_id_map.json.
func GetR12ItemIdToBlockIdMap() *R12ItemIdToBlockIdMap {
	r12MapOnce.Do(func() {
		raw, err := SchemaFS.ReadFile("schema/1.12.0_item_id_to_block_id_map.json")
		if err != nil {
			panic(err)
		}
		var m map[string]string
		if err := json.Unmarshal(raw, &m); err != nil {
			panic(fmt.Sprintf("Invalid blockitem ID mapping table: %v", err))
		}
		r12Map = NewR12ItemIdToBlockIdMap(m)
	})
	return r12Map
}

// ItemIdToBlockId is a port of R12ItemIdToBlockIdMap::itemIdToBlockId.
func (m *R12ItemIdToBlockIdMap) ItemIdToBlockId(itemID string) (string, bool) {
	blockID, ok := m.itemToBlock[strings.ToLower(itemID)]
	return blockID, ok
}

// BlockIdToItemId is a port of R12ItemIdToBlockIdMap::blockIdToItemId.
func (m *R12ItemIdToBlockIdMap) BlockIdToItemId(blockID string) (string, bool) {
	itemID, ok := m.blockToItem[strings.ToLower(blockID)]
	return itemID, ok
}
