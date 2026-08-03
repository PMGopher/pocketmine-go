package light

import (
	"fmt"

	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/utils"
)

// BlockLightUpdate is a port of pocketmine\world\light\BlockLightUpdate.
type BlockLightUpdate struct {
	*Base
	// LightEmitters maps an internal block state ID to the light level it emits (0-15) - built by
	// World from every registered block's real GetLightLevel().
	LightEmitters map[int32]int
}

// NewBlockLightUpdate is a port of BlockLightUpdate::__construct.
func NewBlockLightUpdate(explorer *utils.SubChunkExplorer, lightFilters map[int32]int, lightEmitters map[int32]int) *BlockLightUpdate {
	base := newBase(explorer, lightFilters)
	b := &BlockLightUpdate{Base: base, LightEmitters: lightEmitters}
	base.getCurrentLightArray = func() *format.LightArray {
		return explorer.CurrentSubChunk.GetBlockLightArray()
	}
	return b
}

// RecalculateNode is a port of BlockLightUpdate::recalculateNode.
func (b *BlockLightUpdate) RecalculateNode(x, y, z int) {
	if b.SubChunkExplorer.MoveTo(x, y, z) == utils.StatusInvalid {
		return
	}
	lx, ly, lz := x&format.SubChunkCoordMask, y&format.SubChunkCoordMask, z&format.SubChunkCoordMask
	blockState := b.SubChunkExplorer.CurrentSubChunk.GetBlockStateID(lx, ly, lz)

	level := max(b.LightEmitters[blockState], b.getHighestAdjacentLight(x, y, z)-b.lightFilterFor(blockState))
	b.SetAndUpdateLight(x, y, z, level)
}

// RecalculateChunk is a port of BlockLightUpdate::recalculateChunk.
func (b *BlockLightUpdate) RecalculateChunk(chunkX, chunkZ int) int {
	if b.SubChunkExplorer.MoveToChunk(chunkX, 0, chunkZ) == utils.StatusInvalid {
		panic(fmt.Sprintf("light: chunk (%d,%d) does not exist", chunkX, chunkZ))
	}
	chunk := b.SubChunkExplorer.CurrentChunk

	lightSources := 0
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		subChunk := chunk.GetSubChunk(y)
		subChunk.SetBlockLightArray(format.NewLightArrayFilled(0))

		hasEmitter := false
	layerLoop:
		for _, layer := range subChunk.GetBlockLayers() {
			for _, state := range layer.GetPalette() {
				if b.LightEmitters[state] > 0 {
					hasEmitter = true
					break layerLoop
				}
			}
		}
		if hasEmitter {
			lightSources += b.scanForLightEmittingBlocks(subChunk, chunkX<<format.SubChunkCoordBitSize, y<<format.SubChunkCoordBitSize, chunkZ<<format.SubChunkCoordBitSize)
		}
	}

	return lightSources
}

// scanForLightEmittingBlocks is a port of BlockLightUpdate::scanForLightEmittingBlocks.
func (b *BlockLightUpdate) scanForLightEmittingBlocks(subChunk *format.SubChunk, baseX, baseY, baseZ int) int {
	lightSources := 0
	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			for y := 0; y < format.SubChunkEdgeLength; y++ {
				light := b.LightEmitters[subChunk.GetBlockStateID(x, y, z)]
				if light > 0 {
					b.SetAndUpdateLight(baseX+x, baseY+y, baseZ+z, light)
					lightSources++
				}
			}
		}
	}
	return lightSources
}
