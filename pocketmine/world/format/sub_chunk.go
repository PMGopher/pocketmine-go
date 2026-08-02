package format

const (
	SubChunkCoordBitSize = 4
	SubChunkCoordMask    = ^(^0 << SubChunkCoordBitSize)
	SubChunkEdgeLength   = 1 << SubChunkCoordBitSize
)

// SubChunk is a port of pocketmine\world\format\SubChunk. LightArray (sky/block light) isn't
// ported: pocketmine\network\mcpe\serializer\ChunkSerializer::serializeFullChunk never reads it -
// light isn't part of LevelChunkPacket's payload at all in the current protocol (the client
// computes/receives it separately) - so it would be dead weight with no caller.
type SubChunk struct {
	emptyBlockID int32
	blockLayers  []*PalettedBlockArray
	biomes       *PalettedBlockArray
}

// NewSubChunk is a port of `new SubChunk($emptyBlockId, $blockLayers, $biomes)`.
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

// CollectGarbage is a port of SubChunk::collectGarbage.
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
}
