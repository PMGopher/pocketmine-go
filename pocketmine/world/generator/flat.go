package generator

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/generator/populator"
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
// (LegacyStringToItemParser) this port doesn't have - so NewFlat takes an explicit []FlatLayer and
// []populator.Populator instead of a preset string; see VanillaFlatLayers for the classic default
// preset built this way, and VanillaFlatOreTypes for the "decoration" option's Ore populator setup.
type Flat struct {
	baseChunk *format.Chunk

	seed       int
	random     *utils.Random
	populators []populator.Populator
}

// NewFlat is a port of Flat's constructor plus Flat::generateBaseChunk (folded together, matching
// how the PHP original calls generateBaseChunk from its own constructor). emptyStateID is the
// internal state ID Chunk uses above the topmost layer - this port's Chunk doesn't hardcode
// Block::EMPTY_STATE_ID the way PHP does (see format.Chunk's NewChunk doc comment for why), so
// it's passed explicitly; pass a real air block's GetStateId(). populators may be nil/empty,
// matching the PHP constructor's default (no "decoration" option set).
func NewFlat(seed int, layers []FlatLayer, biomeID int32, emptyStateID int32, populators []populator.Populator) *Flat {
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

	return &Flat{baseChunk: baseChunk, seed: seed, random: utils.NewRandom(seed), populators: populators}
}

// GenerateChunk is a port of Flat::generateChunk. chunkX/chunkZ are unused (matching the PHP
// original: a flat world's generated chunk never depends on its coordinates), but are part of the
// Generator interface for generators that do depend on position (a future Normal generator).
func (f *Flat) GenerateChunk(chunkX, chunkZ int) *format.Chunk {
	return f.baseChunk.Clone()
}

// PopulateChunk is a port of Flat::populateChunk: reseeds the generator's shared Random
// deterministically from the chunk's coordinates and the world seed, then runs every configured
// populator against it, in order.
func (f *Flat) PopulateChunk(world block.World, chunkX, chunkZ int) {
	f.random.SetSeed(0xdeadbeef ^ (chunkX << 8) ^ chunkZ ^ f.seed)
	for _, p := range f.populators {
		p.Populate(world, chunkX, chunkZ, f.random)
	}
}
