package bedrock

import (
	_ "embed"
	"encoding/json"
	"sort"
	"sync"
)

//go:embed assets/item_tags.json
var itemTagsJSON []byte

// ItemTagToIdMap is a port of pocketmine\data\bedrock\ItemTagToIdMap: the vanilla item tags (from
// pmmp BedrockData's item_tags.json) and the item IDs each contains.
type ItemTagToIdMap struct {
	mu          sync.RWMutex
	tagToIdsMap map[string]map[string]bool
}

// NewItemTagToIdMap is a port of ItemTagToIdMap::__construct.
func NewItemTagToIdMap(tagToIds map[string][]string) *ItemTagToIdMap {
	m := &ItemTagToIdMap{tagToIdsMap: map[string]map[string]bool{}}
	for tag, ids := range tagToIds {
		for _, id := range ids {
			m.addIdToTagLocked(tag, id)
		}
	}
	return m
}

var (
	itemTagToIdMapOnce     sync.Once
	itemTagToIdMapInstance *ItemTagToIdMap
)

// GetItemTagToIdMap is ItemTagToIdMap::getInstance (ItemTagToIdMap::make).
func GetItemTagToIdMap() *ItemTagToIdMap {
	itemTagToIdMapOnce.Do(func() {
		var tagToIds map[string][]string
		if err := json.Unmarshal(itemTagsJSON, &tagToIds); err != nil {
			panic("Invalid item tag map: " + err.Error())
		}
		itemTagToIdMapInstance = NewItemTagToIdMap(tagToIds)
	})
	return itemTagToIdMapInstance
}

// GetIdsForTag is a port of ItemTagToIdMap::getIdsForTag.
func (m *ItemTagToIdMap) GetIdsForTag(tag string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.tagToIdsMap[tag]))
	for id := range m.tagToIdsMap[tag] {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// TagContainsId is a port of ItemTagToIdMap::tagContainsId.
func (m *ItemTagToIdMap) TagContainsId(tag, id string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tagToIdsMap[tag][id]
}

// AddIdToTag is a port of ItemTagToIdMap::addIdToTag.
func (m *ItemTagToIdMap) AddIdToTag(tag, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addIdToTagLocked(tag, id)
}

func (m *ItemTagToIdMap) addIdToTagLocked(tag, id string) {
	if m.tagToIdsMap[tag] == nil {
		m.tagToIdsMap[tag] = map[string]bool{}
	}
	m.tagToIdsMap[tag][id] = true
}
