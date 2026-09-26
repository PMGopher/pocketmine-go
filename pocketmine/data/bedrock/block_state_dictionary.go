// Package bedrock is a port of a small slice of pocketmine\data\bedrock: reference data shipped
// by the real PocketMine-MP as a separate Composer package (pocketmine/bedrock-data), not written
// in PHP at all - so there's no PocketMine-MP source to port here, just the data file itself
// (assets/canonical_block_states.nbt for Bedrock 1.26.50 - the version the pinned
// github.com/sandertv/gophertunnel speaks) plus a loader/lookup matching pocketmine\network\mcpe\convert\BlockStateDictionary's shape.
//
// canonical_block_states.nbt lists every vanilla block state Bedrock recognises, in "network NBT"
// encoding (NBT's tag structure, but with zigzag-varint-encoded integers/lengths instead of fixed-
// width ones - a different codec from this port's own pocketmine/nbt package, which only
// implements the two world-save variants (big-endian Java, little-endian Bedrock disk format), not
// this network-specific one). Rather than porting a third NBT codec by hand, this reuses
// github.com/sandertv/gophertunnel/minecraft/nbt's NetworkLittleEndian encoding, since it's the
// same protocol library already relied on for the Bedrock connection itself - this is vendored
// reference data plus the wire codec needed to read it, not "game logic" reimplemented via a
// second library the way BlockTransactionImpl/ChunkSerializer are hand-written.
//
// Source: pmmp/BedrockData has no 1.26.50 release yet (its newest tag is 6.7.0+bedrock-1.26.30),
// so the 1.26.50 assets are vendored from df-mc/dragonfly (MIT, assets/LICENSE-dragonfly), which
// dumps the same files from the client. The format is identical to BedrockData's (states sorted by
// the FNV-1 64 hash of the block name, like the client). Switch back to BedrockData once pmmp
// publishes 1.26.50 data.
//
// A block's position in this list IS its Bedrock network runtime ID (BlockStateDictionary's
// constructor takes a list<BlockStateDictionaryEntry> keyed by array index for exactly this
// reason) - there is no separate ID assignment step.
package bedrock

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	gtnbt "github.com/sandertv/gophertunnel/minecraft/nbt"
)

//go:embed assets/canonical_block_states.nbt
var canonicalBlockStatesData []byte

// BlockStateData is a port of pocketmine\data\bedrock\block\BlockStateData - the common
// information found in a serialized Bedrock blockstate. Unlike the PHP original (which stores
// States as NBT Tag objects), this stores decoded Go values (string, int32, byte, ...) directly,
// since nothing here needs to re-encode a state back to NBT - only to look one up by value.
type BlockStateData struct {
	Name    string
	States  map[string]any
	Version int32
}

var (
	blockStatesOnce   sync.Once
	blockStates       []BlockStateData
	blockStatesByName map[string][]int32
)

// loadBlockStates is a port of BlockStateDictionary::loadPaletteFromString (the palette half; the
// meta map is loaded by loadIdMetaLookup).
func loadBlockStates() {
	blockStatesOnce.Do(func() {
		dec := gtnbt.NewDecoderWithEncoding(bytes.NewReader(canonicalBlockStatesData), gtnbt.NetworkLittleEndian)
		blockStatesByName = make(map[string][]int32)
		for {
			var raw struct {
				Name    string         `nbt:"name"`
				States  map[string]any `nbt:"states"`
				Version int32          `nbt:"version"`
			}
			if err := dec.Decode(&raw); err != nil {
				break
			}
			runtimeID := int32(len(blockStates))
			blockStates = append(blockStates, BlockStateData{Name: raw.Name, States: raw.States, Version: raw.Version})
			blockStatesByName[raw.Name] = append(blockStatesByName[raw.Name], runtimeID)
		}
	})
}

// BlockStates returns the full canonical block state list. A state's index in this slice is its
// Bedrock network runtime ID.
func BlockStates() []BlockStateData {
	loadBlockStates()
	return blockStates
}

// RuntimeIDFor is a port of BlockStateDictionary::lookupStateIdFromData, searching by exact
// name+states match. Returns false if no canonical block state has that exact name and property
// set - e.g. because a property is missing/misspelled, or (rarely) because vanilla renamed the
// block between bedrock-data versions.
func RuntimeIDFor(name string, states map[string]any) (int32, bool) {
	loadBlockStates()
	for _, runtimeID := range blockStatesByName[name] {
		if reflect.DeepEqual(blockStates[runtimeID].States, states) {
			return runtimeID, true
		}
	}
	return 0, false
}

// blockStateMetaMapData is the legacy meta of every palette state (BedrockData's
// block_state_meta_map.json). pmmp has no 1.26.50 BedrockData yet, so this file is derived from
// pmmp/BedrockData 6.7.0+bedrock-1.26.30 by matching states (ignoring the 1.26.50-only
// minecraft:corner and minecraft:connection_* properties); states of blocks added since 1.26.30
// get the index of their distinct state among the block's states.
//
//go:embed assets/block_state_meta_map.json
var blockStateMetaMapData []byte

var (
	idMetaLookupOnce sync.Once
	blockStateMetas  []int
	// idMetaLookup is BlockStateDictionary::getIdMetaToStateIdLookup: name => meta => runtime ID.
	idMetaLookup map[string]map[int]int32
)

func loadIdMetaLookup() {
	idMetaLookupOnce.Do(func() {
		loadBlockStates()
		if err := json.Unmarshal(blockStateMetaMapData, &blockStateMetas); err != nil {
			panic(fmt.Sprintf("bedrock: invalid block state meta map: %v", err))
		}
		if len(blockStateMetas) != len(blockStates) {
			panic(fmt.Sprintf("bedrock: block state meta map has %d entries, the palette has %d states", len(blockStateMetas), len(blockStates)))
		}
		idMetaLookup = map[string]map[int]int32{}
		for i, state := range blockStates {
			if idMetaLookup[state.Name] == nil {
				idMetaLookup[state.Name] = map[int]int32{}
			}
			idMetaLookup[state.Name][blockStateMetas[i]] = int32(i)
		}
	})
}

// GetMetaFromStateId is a port of BlockStateDictionary::getMetaFromStateId.
func GetMetaFromStateId(runtimeID int32) (int, bool) {
	loadIdMetaLookup()
	if runtimeID < 0 || int(runtimeID) >= len(blockStateMetas) {
		return 0, false
	}
	return blockStateMetas[runtimeID], true
}

// LookupStateIdFromIdMeta is a port of BlockStateDictionary::lookupStateIdFromIdMeta. Like PHP, a
// block with only one meta value returns that state for any meta.
func LookupStateIdFromIdMeta(id string, meta int) (int32, bool) {
	loadIdMetaLookup()
	metas, ok := idMetaLookup[id]
	if !ok {
		return 0, false
	}
	if len(metas) == 1 {
		for _, runtimeID := range metas {
			return runtimeID, true
		}
	}
	runtimeID, ok := metas[meta]
	return runtimeID, ok
}

// GenerateDataFromStateId is a port of BlockStateDictionary::generateDataFromStateId.
func GenerateDataFromStateId(runtimeID int32) (BlockStateData, bool) {
	loadBlockStates()
	if runtimeID < 0 || int(runtimeID) >= len(blockStates) {
		return BlockStateData{}, false
	}
	state := blockStates[runtimeID]
	states := make(map[string]any, len(state.States))
	for k, v := range state.States {
		states[k] = v
	}
	return BlockStateData{Name: state.Name, States: states, Version: state.Version}, true
}
