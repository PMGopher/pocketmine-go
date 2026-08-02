package convert

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
)

// BlockTranslator is a port of pocketmine\network\mcpe\convert\BlockTranslator, minus the reverse
// direction (InternalIdToNetworkStateData/network->internal) - nothing in this port needs to go
// from a network runtime ID back to a Behavior yet (that would need a registry of every internal
// state a given block type can have, the inverse of BlockStateSerializer, not built yet either).
type BlockTranslator struct {
	networkIDCache  map[int]int32
	fallbackStateID int32
}

// NewBlockTranslator is a port of BlockTranslator::__construct. Panics if
// "minecraft:info_update" (the vanilla fallback state for unrecognised blockstates) isn't in the
// vendored canonical_block_states.nbt - matching the PHP original's AssumptionFailedError, since
// that would mean the vendored asset itself is broken/mismatched.
func NewBlockTranslator() *BlockTranslator {
	fallbackID, ok := bedrock.RuntimeIDFor("minecraft:info_update", map[string]any{})
	if !ok {
		panic("convert: minecraft:info_update should always exist in canonical_block_states.nbt")
	}
	return &BlockTranslator{
		networkIDCache:  map[int]int32{},
		fallbackStateID: fallbackID,
	}
}

// InternalIDToNetworkID is a port of BlockTranslator::internalIdToNetworkId. Falls back to
// minecraft:info_update (matching the PHP original's BlockStateSerializeException handling)
// whenever the block type has no registered BlockStateSerializer yet, or the serializer's output
// doesn't match any canonical blockstate - both are the same "not supported over the network yet"
// case from the caller's perspective.
func (t *BlockTranslator) InternalIDToNetworkID(blk block.Behavior) int32 {
	internalStateID := blk.GetStateId()
	if networkID, ok := t.networkIDCache[internalStateID]; ok {
		return networkID
	}

	networkID := t.fallbackStateID
	if stateData, err := SerializeBlockState(blk); err == nil {
		if id, ok := bedrock.RuntimeIDFor(stateData.Name, stateData.States); ok {
			networkID = id
		}
	}

	t.networkIDCache[internalStateID] = networkID
	return networkID
}

// FallbackStateID is a port of BlockTranslator::getFallbackStateData, returning the runtime ID
// directly (via minecraft:info_update) instead of the full BlockStateData - nothing here needs the
// full state data, only the ID, unlike the PHP original's persistent (NBT-based) serialization
// path.
func (t *BlockTranslator) FallbackStateID() int32 { return t.fallbackStateID }

// NetworkIDForCachedState looks up a network runtime ID for a bare internal state ID that was
// already translated via InternalIDToNetworkID at least once. This has no direct PHP equivalent -
// PHP's internalIdToNetworkId always takes a bare int state ID and can reconstruct a Block from it
// via a global block factory; this port's Chunk/SubChunk/PalettedBlockArray (deliberately) only
// ever store compact int32 state IDs, not live block.Behavior instances, so chunk serialization
// - which only has those bare IDs on hand - can't call InternalIDToNetworkID directly. In
// practice this is never a cache miss: every ID a chunk can contain was placed there by generator
// code that had a real Behavior in hand and called InternalIDToNetworkID with it first (that's
// how the world/generator package's Flat generator populates a chunk), warming this exact cache
// entry as a side effect. Falls back to the same minecraft:info_update state as
// InternalIDToNetworkID for the one case that isn't true today: a chunk touched by a
// BlockTranslator instance different from the one serializing it.
func (t *BlockTranslator) NetworkIDForCachedState(internalStateID int32) int32 {
	if networkID, ok := t.networkIDCache[int(internalStateID)]; ok {
		return networkID
	}
	return t.fallbackStateID
}
