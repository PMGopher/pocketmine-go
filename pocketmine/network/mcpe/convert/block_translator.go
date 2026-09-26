// Package convert is a port of pocketmine\network\mcpe\convert: translating this port's own
// blocks and items into the Bedrock network forms clients expect.
package convert

import (
	"sync"

	"pocketmine-go/pocketmine/data/bedrock"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
	blockconvert "pocketmine-go/pocketmine/data/bedrock/block/convert"
	worldio "pocketmine-go/pocketmine/world/format/io"
)

// BlockTranslator is a port of pocketmine\network\mcpe\convert\BlockTranslator: internal block
// state IDs to network runtime IDs (positions in the vendored canonical_block_states.nbt), through
// the block state serializer (GlobalBlockStateHandlers). Safe for concurrent use.
type BlockTranslator struct {
	serializer *blockconvert.BlockObjectToStateSerializer

	mu             sync.RWMutex
	networkIDCache map[int]int32

	fallbackStateData bedrock.BlockStateData
	fallbackStateID   int32
}

// NewBlockTranslator is a port of BlockTranslator::__construct, with the dictionary being the
// vendored palette and the serializer GlobalBlockStateHandlers::getSerializer().
func NewBlockTranslator() *BlockTranslator {
	fallback := bedrock.BlockStateData{Name: ids.INFO_UPDATE, States: map[string]any{}, Version: blockconvert.CurrentBlockStateVersion}
	fallbackID, ok := bedrock.RuntimeIDFor(fallback.Name, fallback.States)
	if !ok {
		panic(ids.INFO_UPDATE + " should always exist")
	}
	return &BlockTranslator{
		serializer:        worldio.GetBlockStateSerializer(),
		networkIDCache:    map[int]int32{},
		fallbackStateData: fallback,
		fallbackStateID:   fallbackID,
	}
}

// InternalIDToNetworkID is a port of BlockTranslator::internalIdToNetworkId.
func (t *BlockTranslator) InternalIDToNetworkID(internalStateID int) int32 {
	t.mu.RLock()
	networkID, ok := t.networkIDCache[internalStateID]
	t.mu.RUnlock()
	if ok {
		return networkID
	}

	blockStateData, err := t.serializer.Serialize(internalStateID)
	if err == nil {
		var found bool
		networkID, found = bedrock.RuntimeIDFor(blockStateData.Name, blockStateData.States)
		if !found {
			panic("Unmapped blockstate returned by blockstate serializer: " + blockStateData.Name)
		}
	} else {
		//TODO: this will swallow any error caused by invalid block properties; this is not ideal, but it should be
		//covered by unit tests, so this is probably a safe assumption.
		networkID = t.fallbackStateID
	}

	t.mu.Lock()
	t.networkIDCache[internalStateID] = networkID
	t.mu.Unlock()
	return networkID
}

// NetworkIDForCachedState is InternalIDToNetworkID for the int32 state IDs chunks store (the
// interface world/sound and world/particle use).
func (t *BlockTranslator) NetworkIDForCachedState(internalStateID int32) int32 {
	return t.InternalIDToNetworkID(int(internalStateID))
}

// InternalIDToNetworkStateData is a port of BlockTranslator::internalIdToNetworkStateData.
func (t *BlockTranslator) InternalIDToNetworkStateData(internalStateID int) bedrock.BlockStateData {
	//we don't directly use the blockstate serializer here - we can't assume that the network blockstate NBT is the
	//same as the disk blockstate NBT, in case we decide to have different world version than network version (or in
	//case someone wants to implement multi version).
	return bedrock.BlockStates()[t.InternalIDToNetworkID(internalStateID)]
}

// GetFallbackStateData is a port of BlockTranslator::getFallbackStateData.
func (t *BlockTranslator) GetFallbackStateData() bedrock.BlockStateData { return t.fallbackStateData }

// FallbackStateID returns the network runtime ID of GetFallbackStateData.
func (t *BlockTranslator) FallbackStateID() int32 { return t.fallbackStateID }
