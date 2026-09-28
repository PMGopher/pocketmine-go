package blockconvert

import (
	"sync"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
)

// Pass-through vanilla block states. This isn't in PocketMine-MP, which turns every state it has
// no block for (moss, kelp, dripstone, ...) into the "update!" block and logs an error. Like
// Dragonfly, this port keeps them instead: every state of the vendored palette that the vanilla
// mappings can't deserialize gets an UnknownBlock state ID of its own, which saves back to the
// same state data and is sent to clients as the same runtime ID. The blocks look and save exactly
// as they were; they have no behaviour (UnknownBlock: breaks instantly, no drops).
//
// The state ID of palette entry runtimeID is (passthroughTypeIDBase + runtimeID>>11)<<11 |
// runtimeID&InternalStateDataMask, so it only depends on the palette, and stays in int32 range.
const passthroughTypeIDBase = 1 << 19

var (
	passthroughMu      sync.RWMutex
	passthroughByState map[int32]int // palette runtime ID -> internal state ID
	passthroughByID    map[int]int32 // internal state ID -> palette runtime ID
)

func passthroughStateID(runtimeID int32) int {
	return (passthroughTypeIDBase+int(runtimeID)>>block.InternalStateDataBits)<<block.InternalStateDataBits |
		int(runtimeID)&block.InternalStateDataMask
}

// IsPassthroughState reports whether stateID is a pass-through vanilla state.
func IsPassthroughState(stateID int) bool {
	return stateID>>block.InternalStateDataBits >= passthroughTypeIDBase
}

// RegisterPassthroughStates registers every palette state r's deserializer has no block for as a
// pass-through state (in the RuntimeBlockStateRegistry and r's serializer). It runs once, when the
// global handlers are created, before any world exists.
func RegisterPassthroughStates(r *BlockSerializerDeserializerRegistrar) {
	registry := block.GetRuntimeBlockStateRegistry()
	byState := map[int32]int{}
	byID := map[int]int32{}
	mappedTypes := map[int]bool{}
	for runtimeID, stateData := range bedrock.BlockStates() {
		if _, err := r.Deserializer.Deserialize(stateData); err == nil {
			continue
		}
		stateID := passthroughStateID(int32(runtimeID))
		registry.RegisterUnknownState(stateID)
		byState[int32(runtimeID)] = stateID
		byID[stateID] = int32(runtimeID)

		blk := registry.FromStateId(stateID)
		if !mappedTypes[blk.GetTypeId()] {
			mappedTypes[blk.GetTypeId()] = true
			r.Serializer.Map(blk, serializePassthrough)
		}
	}
	passthroughMu.Lock()
	passthroughByState, passthroughByID = byState, byID
	passthroughMu.Unlock()
}

// serializePassthrough is the serializer of the pass-through types: the palette state the state ID
// was made from. Other state IDs of those types were never registered.
func serializePassthrough(b block.Behavior) bedrock.BlockStateData {
	passthroughMu.RLock()
	runtimeID, ok := passthroughByID[b.GetStateId()]
	passthroughMu.RUnlock()
	if !ok {
		panic(&BlockStateSerializeError{Message: "State ID " + itoa(b.GetStateId()) + " isn't a pass-through vanilla state"})
	}
	return bedrock.BlockStates()[runtimeID]
}

// DeserializeOrPassthrough is Deserialize, falling back to the pass-through state of a vanilla
// state that has no block here. The error is Deserialize's, returned only if stateData isn't a
// palette state either.
func (d *BlockStateToObjectDeserializer) DeserializeOrPassthrough(stateData bedrock.BlockStateData) (int, error) {
	stateID, err := d.Deserialize(stateData)
	if err == nil {
		return stateID, nil
	}
	if runtimeID, ok := bedrock.RuntimeIDFor(stateData.Name, stateData.States); ok {
		passthroughMu.RLock()
		passthroughID, found := passthroughByState[runtimeID]
		passthroughMu.RUnlock()
		if found {
			return passthroughID, nil
		}
	}
	return 0, err
}
