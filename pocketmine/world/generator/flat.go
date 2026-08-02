package generator

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/world/format"
)

// FlatLayer is a port of one entry in FlatGeneratorOptions' parsed structure - a Y-ordered
// (block, height) pair, matching one "N*name"/"name" segment of a flat preset string like
// "bedrock,59xstone,3xdirt,grass".
type FlatLayer struct {
	Block  block.Behavior
	Height int
}

// Flat is a port of pocketmine\world\generator\Flat.
//
// The PHP original also accepts an arbitrary preset string (e.g.
// "2;bedrock,59xstone,3xdirt,grass;1;decoration") parsed at construction time via
// FlatGeneratorOptions::parsePreset, which needs a general item-name parser
// (LegacyStringToItemParser) this port doesn't have - so NewFlat takes an explicit []FlatLayer
// instead of a preset string; see VanillaFlatLayers for the classic default preset built this way.
// The Ore populator (triggered by the PHP original's "decoration" extra option) isn't ported
// either - it needs a whole separate Populator/OreType vein-placement algorithm, unrelated to
// generating a base chunk at all.
type Flat struct {
	baseChunk *format.Chunk
}

// NewFlat is a port of Flat::generateBaseChunk (folded into the constructor, matching how the PHP
// original calls it from its own constructor). emptyStateID is the internal state ID Chunk uses
// above the topmost layer - this port's Chunk doesn't hardcode Block::EMPTY_STATE_ID the way PHP
// does (see format.Chunk's NewChunk doc comment for why), so it's passed explicitly; pass a real
// air block's GetStateId().
func NewFlat(layers []FlatLayer, biomeID int32, emptyStateID int32) *Flat {
	baseChunk := format.NewChunk(nil, false, emptyStateID, biomeID)

	y := 0
	for _, layer := range layers {
		stateID := int32(layer.Block.GetStateId())
		for i := 0; i < layer.Height; i++ {
			for z := 0; z < format.SubChunkEdgeLength; z++ {
				for x := 0; x < format.SubChunkEdgeLength; x++ {
					baseChunk.SetBlockStateID(x, y, z, stateID)
				}
			}
			y++
		}
	}

	return &Flat{baseChunk: baseChunk}
}

// GenerateChunk is a port of Flat::generateChunk. chunkX/chunkZ are unused (matching the PHP
// original: a flat world's generated chunk never depends on its coordinates), but are part of the
// Generator interface for generators that do depend on position (a future Normal generator).
func (f *Flat) GenerateChunk(chunkX, chunkZ int) *format.Chunk {
	return f.baseChunk.Clone()
}
