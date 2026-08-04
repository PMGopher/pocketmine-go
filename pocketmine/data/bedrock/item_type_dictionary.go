package bedrock

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"sort"
	"sync"

	gtnbt "github.com/sandertv/gophertunnel/minecraft/nbt"
)

//go:embed assets/required_item_list.json
var requiredItemListData []byte

// ItemTypeEntry is a port of pocketmine\network\mcpe\protocol\types\ItemTypeEntry - one row of the
// item table sent to the client via the ItemRegistry packet (see ItemTypeDictionaryFromDataHelper,
// this file's real PHP counterpart). Data holds the item's component NBT (see
// required_item_list.json's own "component_nbt" field, base64 LittleEndian NBT - decoded once here
// instead of every time an ItemRegistry packet is built) - nil for the ~96% of items that aren't
// component_based.
type ItemTypeEntry struct {
	Name           string
	RuntimeID      int32
	ComponentBased bool
	Version        int32
	Data           map[string]any
}

var (
	itemTypesOnce        sync.Once
	itemTypes            []ItemTypeEntry
	itemTypesByName      map[string]int32
	itemTypesByRuntimeID map[int32]string
)

type rawItemTypeEntry struct {
	RuntimeID      int32  `json:"runtime_id"`
	ComponentBased bool   `json:"component_based"`
	Version        int32  `json:"version"`
	ComponentNBT   string `json:"component_nbt"`
}

// loadItemTypes is a port of ItemTypeDictionaryFromDataHelper::loadFromString, reading the vendored
// required_item_list.json (from pmmp/BedrockData, the same tag as canonical_block_states.nbt - see
// that file's own doc comment) instead of PHP's json_decode + manual field validation.
func loadItemTypes() {
	itemTypesOnce.Do(func() {
		var raw map[string]rawItemTypeEntry
		if err := json.Unmarshal(requiredItemListData, &raw); err != nil {
			panic("bedrock: invalid required_item_list.json: " + err.Error())
		}

		names := make([]string, 0, len(raw))
		for name := range raw {
			names = append(names, name)
		}
		sort.Strings(names) // deterministic iteration order - JSON object key order isn't guaranteed

		itemTypesByName = make(map[string]int32, len(raw))
		itemTypesByRuntimeID = make(map[int32]string, len(raw))
		for _, name := range names {
			entry := raw[name]

			var data map[string]any
			if entry.ComponentNBT != "" {
				nbtBytes, err := base64.StdEncoding.DecodeString(entry.ComponentNBT)
				if err != nil {
					panic("bedrock: invalid component_nbt base64 for " + name + ": " + err.Error())
				}
				dec := gtnbt.NewDecoderWithEncoding(bytes.NewReader(nbtBytes), gtnbt.LittleEndian)
				if err := dec.Decode(&data); err != nil {
					panic("bedrock: invalid component_nbt for " + name + ": " + err.Error())
				}
			}

			itemTypes = append(itemTypes, ItemTypeEntry{
				Name:           name,
				RuntimeID:      entry.RuntimeID,
				ComponentBased: entry.ComponentBased,
				Version:        entry.Version,
				Data:           data,
			})
			itemTypesByName[name] = entry.RuntimeID
			itemTypesByRuntimeID[entry.RuntimeID] = name
		}
	})
}

// ItemTypes returns the full vendored item table - a port of ItemTypeDictionary::getEntries.
func ItemTypes() []ItemTypeEntry {
	loadItemTypes()
	return itemTypes
}

// ItemRuntimeIDFor is a port of ItemTypeDictionary::fromStringId.
func ItemRuntimeIDFor(name string) (int32, bool) {
	loadItemTypes()
	id, ok := itemTypesByName[name]
	return id, ok
}

// ItemNameForRuntimeID is a port of ItemTypeDictionary::fromIntId.
func ItemNameForRuntimeID(runtimeID int32) (string, bool) {
	loadItemTypes()
	name, ok := itemTypesByRuntimeID[runtimeID]
	return name, ok
}
