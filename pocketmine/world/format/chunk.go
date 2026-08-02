package format

const (
	MinSubChunkIndex = -4
	MaxSubChunkIndex = 19
	MaxSubChunks     = MaxSubChunkIndex - MinSubChunkIndex + 1 // 24
)

// Chunk is a port of a slice of pocketmine\world\format\Chunk: the block/biome storage this
// port's World and network chunk serializer both need. HeightArray (heightmap), tiles, the
// terrain-dirty-flag bitmask, and lightPopulated aren't ported - none of them are read by
// pocketmine\network\mcpe\serializer\ChunkSerializer::serializeFullChunk (the only consumer this
// port has so far), and each depends on machinery not built yet either (a real Tile-in-World
// system for tiles; incremental per-region resend tracking for the dirty flags). Adding them is
// purely additive whenever something needs them - Chunk's shape here doesn't preclude it.
type Chunk struct {
	subChunks        [MaxSubChunks]*SubChunk
	terrainPopulated bool
}

// NewChunk is a port of `new Chunk($subChunks, $terrainPopulated)`. subChunks is keyed by real
// subchunk Y index (MinSubChunkIndex..MaxSubChunkIndex, matching Chunk::getSubChunk's own
// indexing) - any index not present is filled with an empty subchunk using defaultEmptyBlockID/
// defaultBiomeID, matching the PHP constructor's `new SubChunk(Block::EMPTY_STATE_ID, [], new
// PalettedBlockArray(BiomeIds::OCEAN))` default. Both defaults are taken as parameters rather than
// hard-coded: unlike PHP, this package doesn't import pocketmine-go's block package (avoiding an
// unnecessary dependency for a package that only needs a bare int32 state ID), so it can't compute
// air's real state ID itself - the caller (World/generator code, which already has a real
// block.Behavior for air) provides it.
func NewChunk(subChunks map[int]*SubChunk, terrainPopulated bool, defaultEmptyBlockID int32, defaultBiomeID int32) *Chunk {
	c := &Chunk{terrainPopulated: terrainPopulated}
	for y := MinSubChunkIndex; y <= MaxSubChunkIndex; y++ {
		if sc, ok := subChunks[y]; ok {
			c.subChunks[y-MinSubChunkIndex] = sc
		} else {
			c.subChunks[y-MinSubChunkIndex] = NewSubChunk(defaultEmptyBlockID, nil, NewPalettedBlockArray(defaultBiomeID))
		}
	}
	return c
}

// GetSubChunk is a port of Chunk::getSubChunk. Panics for an out-of-range y, matching the PHP
// original's InvalidArgumentException (a programmer error at the call site).
func (c *Chunk) GetSubChunk(y int) *SubChunk {
	if y < MinSubChunkIndex || y > MaxSubChunkIndex {
		panic("format: invalid subchunk Y coordinate")
	}
	return c.subChunks[y-MinSubChunkIndex]
}

// SetSubChunk is a port of Chunk::setSubChunk (minus the "nil replaces with an empty subchunk"
// convenience - callers here always pass a real *SubChunk, so there's no nullable case to handle).
func (c *Chunk) SetSubChunk(y int, subChunk *SubChunk) {
	if y < MinSubChunkIndex || y > MaxSubChunkIndex {
		panic("format: invalid subchunk Y coordinate")
	}
	c.subChunks[y-MinSubChunkIndex] = subChunk
}

// GetBlockStateID is a port of Chunk::getBlockStateId. y is a world Y coordinate, not a
// subchunk-relative one (0-15) - matching the PHP original's `$y >> SubChunk::COORD_BIT_SIZE`
// subchunk lookup.
func (c *Chunk) GetBlockStateID(x, y, z int) int32 {
	return c.GetSubChunk(y>>SubChunkCoordBitSize).GetBlockStateID(x, y&SubChunkCoordMask, z)
}

// SetBlockStateID is a port of Chunk::setBlockStateId.
func (c *Chunk) SetBlockStateID(x, y, z int, block int32) {
	c.GetSubChunk(y>>SubChunkCoordBitSize).SetBlockStateID(x, y&SubChunkCoordMask, z, block)
}

// GetHighestBlockAt is a port of Chunk::getHighestBlockAt. The bool return reports whether any
// block was found in the whole column (PHP returns ?int).
func (c *Chunk) GetHighestBlockAt(x, z int) (int, bool) {
	for y := MaxSubChunkIndex; y >= MinSubChunkIndex; y-- {
		if height, ok := c.GetSubChunk(y).GetHighestBlockAt(x, z); ok {
			return height | (y << SubChunkCoordBitSize), true
		}
	}
	return 0, false
}

// GetBiomeID is a port of Chunk::getBiomeId.
func (c *Chunk) GetBiomeID(x, y, z int) int32 {
	return c.GetSubChunk(y>>SubChunkCoordBitSize).GetBiomeArray().Get(x, y&SubChunkCoordMask, z)
}

// SetBiomeID is a port of Chunk::setBiomeId.
func (c *Chunk) SetBiomeID(x, y, z int, biomeID int32) {
	c.GetSubChunk(y>>SubChunkCoordBitSize).GetBiomeArray().Set(x, y&SubChunkCoordMask, z, biomeID)
}

// IsPopulated is a port of Chunk::isPopulated.
func (c *Chunk) IsPopulated() bool { return c.terrainPopulated }

// SetPopulated is a port of Chunk::setPopulated.
func (c *Chunk) SetPopulated(value bool) { c.terrainPopulated = value }

// CollectGarbage is a port of Chunk::collectGarbage.
func (c *Chunk) CollectGarbage() {
	for _, sc := range c.subChunks {
		sc.CollectGarbage()
	}
}
