// Package bedrock is a port of a small slice of pocketmine\data\bedrock: reference data shipped
// by the real PocketMine-MP as a separate Composer package (pocketmine/bedrock-data), not written
// in PHP at all - so there's no PocketMine-MP source to port here, just the data file itself
// (assets/canonical_block_states.nbt, vendored from pmmp/BedrockData tag 6.7.0+bedrock-1.26.30 -
// the version matching the Bedrock protocol github.com/sandertv/gophertunnel currently speaks)
// plus a loader/lookup matching pocketmine\network\mcpe\convert\BlockStateDictionary's shape.
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
// A block's position in this list IS its Bedrock network runtime ID (BlockStateDictionary's
// constructor takes a list<BlockStateDictionaryEntry> keyed by array index for exactly this
// reason) - there is no separate ID assignment step.
package bedrock

import (
	"bytes"
	_ "embed"
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

// loadBlockStates is a port of BlockStateDictionary::loadPaletteFromString, minus the metaMap
// machinery (idMetaToStateIdLookup/lookupStateIdFromIdMeta): nothing in this port needs legacy
// numeric meta-based lookups yet - only name+states lookups (RuntimeIDFor, below), matching
// BlockStateDictionary's stateDataToStateIdLookup fast path.
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
