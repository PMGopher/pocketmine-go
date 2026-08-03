package format

import "pocketmine-go/pocketmine/block/tile"

// Dirty flag bits are a port of Chunk::DIRTY_FLAG_BLOCKS/DIRTY_FLAG_BIOMES/DIRTY_FLAGS_ALL/
// DIRTY_FLAGS_NONE - tracking which parts of a chunk have changed since the last time whatever
// resend/save mechanism cares about it last cleared them (see ClearTerrainDirtyFlags).
const (
	DirtyFlagBlocks = 1 << 0
	DirtyFlagBiomes = 1 << 3

	DirtyFlagsAll  = ^0
	DirtyFlagsNone = 0
)

const (
	MinSubChunkIndex = -4
	MaxSubChunkIndex = 19
	MaxSubChunks     = MaxSubChunkIndex - MinSubChunkIndex + 1 // 24
)

// Chunk is a port of pocketmine\world\format\Chunk: the block/biome/tile/heightmap storage this
// port's World, network chunk serializer, and (via tile.Tile) block/tile package all need.
type Chunk struct {
	subChunks        [MaxSubChunks]*SubChunk
	terrainPopulated bool
	lightPopulated   *bool // nil = PHP's `?bool` with no value forced yet; matches lightPopulated's own ?bool type (not just bool) since PHP distinguishes "never set" from "explicitly false" here, e.g. via setLightPopulated(null).

	heightMap *HeightArray

	// tiles is keyed by BlockHash(x,y,z) - the same packed single-int key Chunk::addTile/
	// removeTile/getTile use, imported from pocketmine/block/tile directly (see that package's own
	// doc comment on why depending on it here doesn't risk an import cycle: tile only imports
	// math/nbt, and nothing tile depends on ever depends on format).
	tiles map[int]tile.Tile

	terrainDirtyFlags int
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
	c := &Chunk{
		terrainPopulated:  terrainPopulated,
		tiles:             map[int]tile.Tile{},
		terrainDirtyFlags: DirtyFlagsAll,
	}
	falseVal := false
	c.lightPopulated = &falseVal

	for y := MinSubChunkIndex; y <= MaxSubChunkIndex; y++ {
		if sc, ok := subChunks[y]; ok {
			c.subChunks[y-MinSubChunkIndex] = sc
		} else {
			c.subChunks[y-MinSubChunkIndex] = NewSubChunk(defaultEmptyBlockID, nil, NewPalettedBlockArray(defaultBiomeID))
		}
	}

	// Matches `HeightArray::fill((self::MAX_SUBCHUNK_INDEX + 1) * SubChunk::EDGE_LENGTH)` -
	// initially "as if every column were solid all the way to the top", overwritten for real once
	// something (generation, the light engine) computes a real heightmap.
	c.heightMap = NewHeightArrayFilled((MaxSubChunkIndex + 1) * SubChunkEdgeLength)

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
	c.terrainDirtyFlags |= DirtyFlagBlocks
}

// GetSubChunks is a port of Chunk::getSubChunks.
func (c *Chunk) GetSubChunks() map[int]*SubChunk {
	result := make(map[int]*SubChunk, MaxSubChunks)
	for i, sc := range c.subChunks {
		result[i+MinSubChunkIndex] = sc
	}
	return result
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
	c.terrainDirtyFlags |= DirtyFlagBlocks
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
	c.terrainDirtyFlags |= DirtyFlagBiomes
}

// GetHeightMap is a port of Chunk::getHeightMap.
func (c *Chunk) GetHeightMap(x, z int) int { return c.heightMap.Get(x, z) }

// SetHeightMap is a port of Chunk::setHeightMap.
func (c *Chunk) SetHeightMap(x, z, value int) { c.heightMap.Set(x, z, value) }

// GetHeightMapArray is a port of Chunk::getHeightMapArray.
func (c *Chunk) GetHeightMapArray() [256]int { return c.heightMap.GetValues() }

// SetHeightMapArray is a port of Chunk::setHeightMapArray.
func (c *Chunk) SetHeightMapArray(values [256]int) { c.heightMap.SetValues(values) }

// IsLightPopulated is a port of Chunk::isLightPopulated. The bool return's own validity mirrors
// PHP's `?bool`: real PHP can hold null (light population state unknown/reset), which this
// represents as ok=false - callers should treat that the same way PHP callers treat null.
func (c *Chunk) IsLightPopulated() (populated bool, ok bool) {
	if c.lightPopulated == nil {
		return false, false
	}
	return *c.lightPopulated, true
}

// SetLightPopulated is a port of Chunk::setLightPopulated(?bool $value = true). Pass ok=false to
// match PHP's setLightPopulated(null).
func (c *Chunk) SetLightPopulated(value bool, ok bool) {
	if !ok {
		c.lightPopulated = nil
		return
	}
	v := value
	c.lightPopulated = &v
}

// IsPopulated is a port of Chunk::isPopulated.
func (c *Chunk) IsPopulated() bool { return c.terrainPopulated }

// SetPopulated is a port of Chunk::setPopulated(bool $value = true).
func (c *Chunk) SetPopulated(value bool) {
	c.terrainPopulated = value
	c.terrainDirtyFlags |= DirtyFlagBlocks
}

// IsTerrainDirty is a port of Chunk::isTerrainDirty.
func (c *Chunk) IsTerrainDirty() bool { return c.terrainDirtyFlags != DirtyFlagsNone }

// GetTerrainDirtyFlag is a port of Chunk::getTerrainDirtyFlag.
func (c *Chunk) GetTerrainDirtyFlag(flag int) bool { return c.terrainDirtyFlags&flag != 0 }

// GetTerrainDirtyFlags is a port of Chunk::getTerrainDirtyFlags.
func (c *Chunk) GetTerrainDirtyFlags() int { return c.terrainDirtyFlags }

// SetTerrainDirtyFlag is a port of Chunk::setTerrainDirtyFlag.
func (c *Chunk) SetTerrainDirtyFlag(flag int, value bool) {
	if value {
		c.terrainDirtyFlags |= flag
	} else {
		c.terrainDirtyFlags &^= flag
	}
}

// SetTerrainDirty is a port of Chunk::setTerrainDirty.
func (c *Chunk) SetTerrainDirty() { c.terrainDirtyFlags = DirtyFlagsAll }

// ClearTerrainDirtyFlags is a port of Chunk::clearTerrainDirtyFlags.
func (c *Chunk) ClearTerrainDirtyFlags() { c.terrainDirtyFlags = DirtyFlagsNone }

// BlockHash is a port of Chunk::blockHash: packs chunk-local block coordinates into a single int
// key, used for this Chunk's own tile map (matching real Chunk::addTile/removeTile/getTile).
func BlockHash(x, y, z int) int {
	return (y << (2 * SubChunkCoordBitSize)) | ((z & SubChunkCoordMask) << SubChunkCoordBitSize) | (x & SubChunkCoordMask)
}

// AddTile is a port of Chunk::addTile. Panics on a closed tile or a location collision with a
// different tile, matching the PHP original's InvalidArgumentException (both are programmer
// errors at the call site, not conditions this port should quietly recover from).
func (c *Chunk) AddTile(t tile.Tile) {
	if t.IsClosed() {
		panic("format: attempted to add a closed tile to a chunk")
	}
	pos := t.GetPosition()
	index := BlockHash(pos.FloorX(), pos.FloorY(), pos.FloorZ())
	if existing, ok := c.tiles[index]; ok && existing != t {
		panic("format: another tile is already at this location")
	}
	c.tiles[index] = t
}

// RemoveTile is a port of Chunk::removeTile.
func (c *Chunk) RemoveTile(t tile.Tile) {
	pos := t.GetPosition()
	delete(c.tiles, BlockHash(pos.FloorX(), pos.FloorY(), pos.FloorZ()))
}

// GetTiles is a port of Chunk::getTiles.
func (c *Chunk) GetTiles() map[int]tile.Tile { return c.tiles }

// GetTile is a port of Chunk::getTile. x/y/z are chunk-local (x/z 0-15, y a world Y coordinate),
// matching the PHP original.
func (c *Chunk) GetTile(x, y, z int) (tile.Tile, bool) {
	t, ok := c.tiles[BlockHash(x, y, z)]
	return t, ok
}

// OnUnload is a port of Chunk::onUnload: closes every tile still in this chunk.
func (c *Chunk) OnUnload() {
	for _, t := range c.tiles {
		t.Close()
	}
}

// CollectGarbage is a port of Chunk::collectGarbage.
func (c *Chunk) CollectGarbage() {
	for _, sc := range c.subChunks {
		sc.CollectGarbage()
	}
}

// Clone is a port of Chunk's implicit __clone (deep-copies every subchunk - PHP's own comment on
// why entities/tiles aren't cloned too, "impractical to do so, too many dependencies", doesn't
// apply here since neither is part of this port's Chunk yet at all).
func (c *Chunk) Clone() *Chunk {
	clone := &Chunk{
		terrainPopulated:  c.terrainPopulated,
		terrainDirtyFlags: c.terrainDirtyFlags,
		tiles:             map[int]tile.Tile{}, // deliberately not cloned - see PHP __clone's own comment
	}
	for i, sc := range c.subChunks {
		clone.subChunks[i] = sc.Clone()
	}
	if c.lightPopulated != nil {
		v := *c.lightPopulated
		clone.lightPopulated = &v
	}
	clone.heightMap = c.heightMap.Clone()
	return clone
}
