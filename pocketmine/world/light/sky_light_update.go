package light

import (
	"fmt"

	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/utils"
)

// worldYMin/worldYMax mirror world.YMin/YMax (format.MinSubChunkIndex/MaxSubChunkIndex times
// format.SubChunkEdgeLength) - computed directly from format's own constants rather than taking a
// parameter from World, since world/light must not import world (world imports world/light) and
// these are really just format.Chunk's own vertical range, not truly World-specific configuration.
const (
	worldYMin = format.MinSubChunkIndex * format.SubChunkEdgeLength
	worldYMax = (format.MaxSubChunkIndex + 1) * format.SubChunkEdgeLength
)

// SkyLightUpdate is a port of pocketmine\world\light\SkyLightUpdate.
type SkyLightUpdate struct {
	*Base
	// DirectSkyLightBlockers is the set of internal state IDs that fully block direct sky light
	// (Block::blocksDirectSkyLight() === true for every state registered with World) - used to
	// compute each column's heightmap.
	DirectSkyLightBlockers map[int32]bool
}

// NewSkyLightUpdate is a port of SkyLightUpdate::__construct.
func NewSkyLightUpdate(explorer *utils.SubChunkExplorer, lightFilters map[int32]int, directSkyLightBlockers map[int32]bool) *SkyLightUpdate {
	base := newBase(explorer, lightFilters)
	s := &SkyLightUpdate{Base: base, DirectSkyLightBlockers: directSkyLightBlockers}

	base.getCurrentLightArray = func() *format.LightArray {
		return explorer.CurrentSubChunk.GetBlockSkyLightArray()
	}
	base.effectiveLightOverride = func(x, y, z int) (int, bool) {
		if y >= worldYMax {
			explorer.Invalidate()
			return 15, true
		}
		return 0, false
	}
	return s
}

// RecalculateNode is a port of SkyLightUpdate::recalculateNode.
func (s *SkyLightUpdate) RecalculateNode(x, y, z int) {
	if s.SubChunkExplorer.MoveTo(x, y, z) == utils.StatusInvalid {
		return
	}
	chunk := s.SubChunkExplorer.CurrentChunk

	lx, lz := x&format.SubChunkCoordMask, z&format.SubChunkCoordMask
	oldHeightMap := chunk.GetHeightMap(lx, lz)
	source := s.SubChunkExplorer.CurrentSubChunk.GetBlockStateID(lx, y&format.SubChunkCoordMask, lz)

	yPlusOne := y + 1

	var newHeightMap int
	switch {
	case yPlusOne == oldHeightMap:
		// Block changed directly beneath the heightmap - check if a block was removed or changed
		// to a different light filter.
		newHeightMap = recalculateHeightMapColumn(chunk, lx, lz, s.DirectSkyLightBlockers)
		chunk.SetHeightMap(lx, lz, newHeightMap)
	case yPlusOne > oldHeightMap:
		// Block changed above the heightmap.
		if !s.DirectSkyLightBlockers[source] {
			// No effect on direct sky light, e.g. placing/removing glass.
			return
		}
		chunk.SetHeightMap(lx, lz, yPlusOne)
		newHeightMap = yPlusOne
	default:
		// Block changed below the heightmap.
		newHeightMap = oldHeightMap
	}

	if newHeightMap >= oldHeightMap {
		for i := y - 1; i >= oldHeightMap; i-- {
			s.SetAndUpdateLight(x, i, z, 0) // Remove all light beneath; adjacent recalculation handles the rest.
		}
		// Recalculate light for the placed block from its surroundings - avoids re-checking
		// effective light during propagation.
		level := s.getHighestAdjacentLight(x, y, z) - s.lightFilterFor(source)
		s.SetAndUpdateLight(x, y, z, max(0, level))
	} else {
		// Heightmap decreased (block changed/removed) - add sky light.
		for i := y; i >= newHeightMap; i-- {
			s.SetAndUpdateLight(x, i, z, 15)
		}
	}
}

// RecalculateChunk is a port of SkyLightUpdate::recalculateChunk: scans for all light sources in
// the target chunk and adds them to the propagation queue, erasing preexisting light in the
// chunk.
func (s *SkyLightUpdate) RecalculateChunk(chunkX, chunkZ int) int {
	if s.SubChunkExplorer.MoveToChunk(chunkX, 0, chunkZ) == utils.StatusInvalid {
		panic(fmt.Sprintf("light: chunk (%d,%d) does not exist", chunkX, chunkZ))
	}
	chunk := s.SubChunkExplorer.CurrentChunk

	recalculateHeightMap(chunk, s.DirectSkyLightBlockers)

	highestHeightMapPlusOne := worldYMin
	hm := chunk.GetHeightMapArray()
	for _, v := range hm {
		highestHeightMapPlusOne = max(highestHeightMapPlusOne, v)
	}
	highestHeightMapPlusOne++

	// setAndUpdateLight won't bother propagating from nodes that are already what we want to
	// change them to, so we have to avoid filling full light for any subchunk that contains a
	// heightmap Y coordinate.
	lowestClearSubChunk := highestHeightMapPlusOne >> format.SubChunkCoordBitSize
	if highestHeightMapPlusOne&format.SubChunkCoordMask != 0 {
		lowestClearSubChunk++
	}

	for y := format.MinSubChunkIndex; y < lowestClearSubChunk && y <= format.MaxSubChunkIndex; y++ {
		chunk.GetSubChunk(y).SetBlockSkyLightArray(format.NewLightArrayFilled(0))
	}
	for y := lowestClearSubChunk; y <= format.MaxSubChunkIndex; y++ {
		chunk.GetSubChunk(y).SetBlockSkyLightArray(format.NewLightArrayFilled(15))
	}

	lightSources := 0
	baseX := chunkX << format.SubChunkCoordBitSize
	baseZ := chunkZ << format.SubChunkCoordBitSize

	for x := 0; x < format.SubChunkEdgeLength; x++ {
		for z := 0; z < format.SubChunkEdgeLength; z++ {
			currentHeight := chunk.GetHeightMap(x, z)

			if currentHeight == worldYMax {
				// This column has a light-filtering block in the top cell - light it from above
				// the world (light from above the world bounds isn't checked during propagation).
				y := currentHeight - 1
				if s.SubChunkExplorer.MoveTo(x+baseX, y, z+baseZ) != utils.StatusInvalid {
					blk := s.SubChunkExplorer.CurrentSubChunk.GetBlockStateID(x, y&format.SubChunkCoordMask, z)
					s.SetAndUpdateLight(x+baseX, y, z+baseZ, max(0, 15-s.lightFilterFor(blk)))
				}
				continue
			}

			maxAdjacentHeight := worldYMin
			if x != 0 {
				maxAdjacentHeight = max(maxAdjacentHeight, chunk.GetHeightMap(x-1, z))
			}
			if x != format.SubChunkEdgeLength-1 {
				maxAdjacentHeight = max(maxAdjacentHeight, chunk.GetHeightMap(x+1, z))
			}
			if z != 0 {
				maxAdjacentHeight = max(maxAdjacentHeight, chunk.GetHeightMap(x, z-1))
			}
			if z != format.SubChunkEdgeLength-1 {
				maxAdjacentHeight = max(maxAdjacentHeight, chunk.GetHeightMap(x, z+1))
			}

			// Skip the top two blocks between current height and max adjacent (if any) - the
			// block next to the highest adjacent does nothing during propagation (surrounded by
			// 15s), and the block below it does the same as the node in the highest adjacent.
			nodeColumnEnd := max(currentHeight, maxAdjacentHeight-2)
			for y := currentHeight; y <= nodeColumnEnd; y++ {
				s.SetAndUpdateLight(x+baseX, y, z+baseZ, 15)
				lightSources++
			}

			yMaxLoop := lowestClearSubChunk * format.SubChunkEdgeLength
			for y := nodeColumnEnd + 1; y < yMaxLoop; y++ {
				if s.SubChunkExplorer.MoveTo(x+baseX, y, z+baseZ) != utils.StatusInvalid {
					s.getCurrentLightArray().Set(x, y&format.SubChunkCoordMask, z, 15)
				}
			}
		}
	}

	return lightSources
}

// recalculateHeightMap is a port of SkyLightUpdate::recalculateHeightMap. Unlike the PHP original
// (which builds a separate HeightArray object and only applies it to the chunk afterwards via
// setHeightMapArray), this mutates chunk's heightmap directly - safe here because nothing in this
// function ever reads the chunk's *old* heightmap while computing the new one, only
// GetSubChunk(...).GetHighestBlockAt/GetBlockStateID, so the two approaches are behaviourally
// identical.
func recalculateHeightMap(chunk *format.Chunk, directSkyLightBlockers map[int32]bool) {
	maxSubChunkY := format.MaxSubChunkIndex
	for ; maxSubChunkY >= format.MinSubChunkIndex; maxSubChunkY-- {
		if !chunk.GetSubChunk(maxSubChunkY).IsEmptyFast() {
			break
		}
	}
	if maxSubChunkY < format.MinSubChunkIndex {
		// The whole column is definitely empty.
		var empty [256]int
		for i := range empty {
			empty[i] = worldYMin
		}
		chunk.SetHeightMapArray(empty)
		return
	}

	for z := 0; z < format.SubChunkEdgeLength; z++ {
		for x := 0; x < format.SubChunkEdgeLength; x++ {
			y, found := worldYMin, false
			for subChunkY := maxSubChunkY; subChunkY >= format.MinSubChunkIndex; subChunkY-- {
				if subHighest, ok := chunk.GetSubChunk(subChunkY).GetHighestBlockAt(x, z); ok {
					y = subChunkY*format.SubChunkEdgeLength + subHighest
					found = true
					break
				}
			}

			if !found {
				chunk.SetHeightMap(x, z, worldYMin)
				continue
			}
			for ; y >= worldYMin; y-- {
				if directSkyLightBlockers[chunk.GetBlockStateID(x, y, z)] {
					chunk.SetHeightMap(x, z, y+1)
					break
				}
			}
		}
	}
}

// recalculateHeightMapColumn is a port of SkyLightUpdate::recalculateHeightMapColumn.
func recalculateHeightMapColumn(chunk *format.Chunk, x, z int, directSkyLightBlockers map[int32]bool) int {
	y, ok := chunk.GetHighestBlockAt(x, z)
	if !ok {
		return worldYMin
	}
	for ; y >= worldYMin; y-- {
		if directSkyLightBlockers[chunk.GetBlockStateID(x, y, z)] {
			break
		}
	}
	return y + 1
}
