// Package utils is a port of pocketmine\world\utils: SubChunkExplorer, a cached "current
// chunk/subchunk pointer" used by anything (the light engine, in this port) that needs to visit a
// lot of nearby positions without re-looking-up the containing chunk/subchunk every single time.
package utils

import "pocketmine-go/pocketmine/world/format"

// Status is a port of SubChunkExplorerStatus's 3 constants.
type Status int

const (
	// StatusInvalid is returned when the requested position is in terrain not accessible by the
	// current ChunkSource (chunk not loaded, or Y out of the world's subchunk range).
	StatusInvalid Status = 0
	// StatusOK is returned when the explorer remained inside the same (sub)chunk.
	StatusOK Status = 1
	// StatusMoved is returned when the explorer moved to a different (sub)chunk.
	StatusMoved Status = 2
)

// ChunkSource is the minimal, non-generating "get an already-loaded chunk" capability
// SubChunkExplorer needs from a ChunkManager - declared locally (matching this port's established
// forward-compatible-local-interface convention) so *world.World satisfies it structurally without
// world/utils needing to import world (which imports world/utils and world/light - both of which
// use this package - so importing back would cycle).
type ChunkSource interface {
	// GetChunk returns the chunk at the given coordinates only if it's already loaded - never
	// generates one, matching World::getChunk (as opposed to World::getOrLoadChunk/loadChunk).
	GetChunk(chunkX, chunkZ int) (*format.Chunk, bool)
}

// SubChunkExplorer is a port of pocketmine\world\utils\SubChunkExplorer.
type SubChunkExplorer struct {
	world ChunkSource

	CurrentChunk    *format.Chunk
	CurrentSubChunk *format.SubChunk

	currentX, currentY, currentZ int
}

func NewSubChunkExplorer(world ChunkSource) *SubChunkExplorer {
	return &SubChunkExplorer{world: world}
}

// MoveTo is a port of SubChunkExplorer::moveTo.
func (e *SubChunkExplorer) MoveTo(x, y, z int) Status {
	newChunkX := x >> format.SubChunkCoordBitSize
	newChunkZ := z >> format.SubChunkCoordBitSize
	if e.CurrentChunk == nil || e.currentX != newChunkX || e.currentZ != newChunkZ {
		e.currentX = newChunkX
		e.currentZ = newChunkZ
		e.CurrentSubChunk = nil

		chunk, ok := e.world.GetChunk(e.currentX, e.currentZ)
		if !ok {
			e.CurrentChunk = nil
			return StatusInvalid
		}
		e.CurrentChunk = chunk
	}

	newChunkY := y >> format.SubChunkCoordBitSize
	if e.CurrentSubChunk == nil || e.currentY != newChunkY {
		e.currentY = newChunkY

		if e.currentY < format.MinSubChunkIndex || e.currentY > format.MaxSubChunkIndex {
			e.CurrentSubChunk = nil
			return StatusInvalid
		}

		e.CurrentSubChunk = e.CurrentChunk.GetSubChunk(newChunkY)
		return StatusMoved
	}

	return StatusOK
}

// MoveToChunk is a port of SubChunkExplorer::moveToChunk.
func (e *SubChunkExplorer) MoveToChunk(chunkX, chunkY, chunkZ int) Status {
	return e.MoveTo(chunkX<<format.SubChunkCoordBitSize, chunkY<<format.SubChunkCoordBitSize, chunkZ<<format.SubChunkCoordBitSize)
}

// IsValid is a port of SubChunkExplorer::isValid.
func (e *SubChunkExplorer) IsValid() bool { return e.CurrentSubChunk != nil }

// Invalidate is a port of SubChunkExplorer::invalidate.
func (e *SubChunkExplorer) Invalidate() {
	e.CurrentChunk = nil
	e.CurrentSubChunk = nil
}
