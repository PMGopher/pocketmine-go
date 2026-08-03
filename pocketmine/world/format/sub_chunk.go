package format

const (
	SubChunkCoordBitSize = 4
	SubChunkCoordMask    = ^(^0 << SubChunkCoordBitSize)
	SubChunkEdgeLength   = 1 << SubChunkCoordBitSize
)

// SubChunk is a port of pocketmine\world\format\SubChunk, including its sky/block LightArray
// fields - not read by pocketmine\network\mcpe\serializer\ChunkSerializer::serializeFullChunk
// (light isn't part of LevelChunkPacket's payload; the client computes/receives it separately),
// but needed by the world/light engine, which works purely server-side.
type SubChunk struct {
	emptyBlockID int32
	blockLayers  []*PalettedBlockArray
	biomes       *PalettedBlockArray

	skyLight   *LightArray
	blockLight *LightArray
}

// NewSubChunk is a port of `new SubChunk($emptyBlockId, $blockLayers, $biomes)` - the 2-argument
// form; skyLight/blockLight default to nil (PHP's own constructor defaults), lazily materialized
// on first access via GetBlockSkyLightArray/GetBlockLightArray, matching the PHP original's `??=`.
func NewSubChunk(emptyBlockID int32, blockLayers []*PalettedBlockArray, biomes *PalettedBlockArray) *SubChunk {
	return &SubChunk{emptyBlockID: emptyBlockID, blockLayers: blockLayers, biomes: biomes}
}

// IsEmptyFast is a port of SubChunk::isEmptyFast.
func (s *SubChunk) IsEmptyFast() bool { return len(s.blockLayers) == 0 }

// GetEmptyBlockID is a port of SubChunk::getEmptyBlockId.
func (s *SubChunk) GetEmptyBlockID() int32 { return s.emptyBlockID }

// GetBlockStateID is a port of SubChunk::getBlockStateId.
func (s *SubChunk) GetBlockStateID(x, y, z int) int32 {
	if len(s.blockLayers) == 0 {
		return s.emptyBlockID
	}
	return s.blockLayers[0].Get(x, y, z)
}

// SetBlockStateID is a port of SubChunk::setBlockStateId.
func (s *SubChunk) SetBlockStateID(x, y, z int, block int32) {
	if len(s.blockLayers) == 0 {
		s.blockLayers = append(s.blockLayers, NewPalettedBlockArray(s.emptyBlockID))
	}
	s.blockLayers[0].Set(x, y, z, block)
}

// GetBlockLayers is a port of SubChunk::getBlockLayers.
func (s *SubChunk) GetBlockLayers() []*PalettedBlockArray { return s.blockLayers }

// GetHighestBlockAt is a port of SubChunk::getHighestBlockAt. The bool return reports whether any
// non-empty block was found (PHP returns ?int; Go has no nullable int).
func (s *SubChunk) GetHighestBlockAt(x, z int) (int, bool) {
	if len(s.blockLayers) == 0 {
		return 0, false
	}
	for y := SubChunkEdgeLength - 1; y >= 0; y-- {
		if s.blockLayers[0].Get(x, y, z) != s.emptyBlockID {
			return y, true
		}
	}
	return 0, false
}

// GetBiomeArray is a port of SubChunk::getBiomeArray.
func (s *SubChunk) GetBiomeArray() *PalettedBlockArray { return s.biomes }

// GetBlockSkyLightArray is a port of SubChunk::getBlockSkyLightArray (lazily fills with 0, matching
// the PHP original's `$this->skyLight ??= LightArray::fill(0)`).
func (s *SubChunk) GetBlockSkyLightArray() *LightArray {
	if s.skyLight == nil {
		s.skyLight = NewLightArrayFilled(0)
	}
	return s.skyLight
}

// SetBlockSkyLightArray is a port of SubChunk::setBlockSkyLightArray.
func (s *SubChunk) SetBlockSkyLightArray(data *LightArray) { s.skyLight = data }

// GetBlockLightArray is a port of SubChunk::getBlockLightArray.
func (s *SubChunk) GetBlockLightArray() *LightArray {
	if s.blockLight == nil {
		s.blockLight = NewLightArrayFilled(0)
	}
	return s.blockLight
}

// SetBlockLightArray is a port of SubChunk::setBlockLightArray.
func (s *SubChunk) SetBlockLightArray(data *LightArray) { s.blockLight = data }

// Clone is a port of SubChunk's implicit __clone (deep-copies every block layer, the biome array,
// and the sky/block light arrays if present, matching the PHP original's `array_map(fn($array) =>
// clone $array, ...)` plus its explicit skyLight/blockLight clone checks).
func (s *SubChunk) Clone() *SubChunk {
	layers := make([]*PalettedBlockArray, len(s.blockLayers))
	for i, layer := range s.blockLayers {
		layers[i] = layer.Clone()
	}
	c := &SubChunk{
		emptyBlockID: s.emptyBlockID,
		blockLayers:  layers,
		biomes:       s.biomes.Clone(),
	}
	if s.skyLight != nil {
		c.skyLight = s.skyLight.Clone()
	}
	if s.blockLight != nil {
		c.blockLight = s.blockLight.Clone()
	}
	return c
}

// CollectGarbage is a port of SubChunk::collectGarbage, including its skyLight/blockLight ==
// uniform(0) => nil collapse.
func (s *SubChunk) CollectGarbage() {
	cleaned := make([]*PalettedBlockArray, 0, len(s.blockLayers))
	for _, layer := range s.blockLayers {
		layer.CollectGarbage()
		if layer.GetBitsPerBlock() != 0 || layer.Get(0, 0, 0) != s.emptyBlockID {
			cleaned = append(cleaned, layer)
		}
	}
	s.blockLayers = cleaned
	s.biomes.CollectGarbage()

	if s.skyLight != nil && s.skyLight.IsUniform(0) {
		s.skyLight = nil
	}
	if s.blockLight != nil && s.blockLight.IsUniform(0) {
		s.blockLight = nil
	}
}
