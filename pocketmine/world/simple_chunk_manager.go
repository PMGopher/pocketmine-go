package world

import (
	"fmt"
	stdmath "math"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/sound"
)

// blockStateRegistry is the read-only view of a World's per-state tables (PHP's
// RuntimeBlockStateRegistry singleton) handed to worker goroutines. The World keeps adding states
// to its own tables on the main thread (registerTemplate), so workers get a frozen copy, rebuilt
// only when something was added since the last one.
type blockStateRegistry struct {
	templates              map[int32]block.Behavior
	lightFilters           map[int32]int
	lightEmitters          map[int32]int
	directSkyLightBlockers map[int32]bool
	airStateID             int32
}

// registrySnapshot returns a frozen copy of the World's per-state tables for worker use.
func (w *World) registrySnapshot() *blockStateRegistry {
	if w.registrySnap != nil && w.registrySnapVersion == w.registryVersion {
		return w.registrySnap
	}
	r := &blockStateRegistry{
		templates:              make(map[int32]block.Behavior, len(w.stateTemplates)),
		lightFilters:           make(map[int32]int, len(w.lightFilters)),
		lightEmitters:          make(map[int32]int, len(w.lightEmitters)),
		directSkyLightBlockers: make(map[int32]bool, len(w.directSkyLightBlockers)),
		airStateID:             int32(block.VanillaAir().GetStateId()),
	}
	for k, v := range w.stateTemplates {
		r.templates[k] = v
	}
	for k, v := range w.lightFilters {
		r.lightFilters[k] = v
	}
	for k, v := range w.lightEmitters {
		r.lightEmitters[k] = v
	}
	for k, v := range w.directSkyLightBlockers {
		r.directSkyLightBlockers[k] = v
	}
	w.registrySnap, w.registrySnapVersion = r, w.registryVersion
	return r
}

// SimpleChunkManager is a port of pocketmine\world\SimpleChunkManager: a detached set of chunks
// that generators and populators read and write on a worker goroutine, with no events,
// neighbour updates, light updates or changed-block tracking.
//
// The populators of this port are written against block.World rather than PHP's narrower
// ChunkManager, so SimpleChunkManager implements block.World; the methods ChunkManager doesn't
// have (tiles, sounds, light, entities, scheduled updates) do nothing, as nothing a populator does
// reaches them.
type SimpleChunkManager struct {
	chunks   map[[2]int]*format.Chunk
	minY     int
	maxY     int
	registry *blockStateRegistry
	// newTemplates are block states written here that the registry snapshot didn't know yet; the
	// World registers them when it takes the chunks back (registerTemplate).
	newTemplates map[int32]block.Behavior
}

// NewSimpleChunkManager is a port of SimpleChunkManager::__construct.
func NewSimpleChunkManager(minY, maxY int, registry *blockStateRegistry) *SimpleChunkManager {
	return &SimpleChunkManager{chunks: map[[2]int]*format.Chunk{}, minY: minY, maxY: maxY, registry: registry, newTemplates: map[int32]block.Behavior{}}
}

func (m *SimpleChunkManager) template(stateID int32) block.Behavior {
	if tpl, ok := m.registry.templates[stateID]; ok {
		return tpl
	}
	if tpl, ok := m.newTemplates[stateID]; ok {
		return tpl
	}
	return m.registry.templates[m.registry.airStateID]
}

// GetBlockAt is a port of SimpleChunkManager::getBlockAt: air outside loaded terrain.
func (m *SimpleChunkManager) GetBlockAt(x, y, z int) block.Behavior {
	var got block.Behavior
	if chunk, ok := m.GetChunk(x>>4, z>>4); ok && m.IsInWorld(x, y, z) {
		got = m.template(chunk.GetBlockStateID(x&0xf, y, z&0xf)).Clone()
	} else {
		got = block.VanillaAir()
	}
	got.(positionable).SetPosition(m, x, y, z)
	return got
}

// SetBlockAt is a port of SimpleChunkManager::setBlockAt.
func (m *SimpleChunkManager) SetBlockAt(x, y, z int, blk block.Behavior) error {
	chunk, ok := m.GetChunk(x>>4, z>>4)
	if !ok {
		return fmt.Errorf("Cannot set block at coordinates x=%d,y=%d,z=%d, terrain is not loaded or out of bounds", x, y, z)
	}
	stateID := int32(blk.GetStateId())
	if _, known := m.registry.templates[stateID]; !known {
		if _, recorded := m.newTemplates[stateID]; !recorded {
			m.newTemplates[stateID] = blk.Clone()
		}
	}
	chunk.SetBlockStateID(x&0xf, y, z&0xf, stateID)
	return nil
}

// SetBlock is block.World's form of SetBlockAt.
func (m *SimpleChunkManager) SetBlock(pos block.Position, blk block.Behavior) error {
	return m.SetBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ(), blk)
}

// GetChunk is a port of SimpleChunkManager::getChunk.
func (m *SimpleChunkManager) GetChunk(chunkX, chunkZ int) (*format.Chunk, bool) {
	c, ok := m.chunks[[2]int{chunkX, chunkZ}]
	return c, ok
}

// SetChunk is a port of SimpleChunkManager::setChunk.
func (m *SimpleChunkManager) SetChunk(chunkX, chunkZ int, chunk *format.Chunk) {
	m.chunks[[2]int{chunkX, chunkZ}] = chunk
}

// CleanChunks is a port of SimpleChunkManager::cleanChunks.
func (m *SimpleChunkManager) CleanChunks() { m.chunks = map[[2]int]*format.Chunk{} }

func (m *SimpleChunkManager) GetMinY() int { return m.minY }
func (m *SimpleChunkManager) GetMaxY() int { return m.maxY }

// IsInWorld is a port of SimpleChunkManager::isInWorld.
func (m *SimpleChunkManager) IsInWorld(x, y, z int) bool {
	return x <= stdmath.MaxInt32 && x >= stdmath.MinInt32 &&
		y < m.maxY && y >= m.minY &&
		z <= stdmath.MaxInt32 && z >= stdmath.MinInt32
}

// GetOrLoadChunkAtPosition is ChunkManager::getChunk for a block position (block.World's form).
func (m *SimpleChunkManager) GetOrLoadChunkAtPosition(pos block.Position) (block.Chunk, bool) {
	chunk, ok := m.GetChunk(pos.FloorX()>>4, pos.FloorZ()>>4)
	if !ok {
		return nil, false
	}
	return chunkAdapter{chunk}, true
}

// The rest of block.World isn't part of ChunkManager: no-ops (see the type doc comment).

func (m *SimpleChunkManager) GetTile(pos block.Position) (block.Tile, bool)          { return nil, false }
func (m *SimpleChunkManager) AddTile(t block.Tile)                                   {}
func (m *SimpleChunkManager) AddSound(pos math.Vector3, s sound.Sound)               {}
func (m *SimpleChunkManager) ScheduleDelayedBlockUpdate(pos math.Vector3, delay int) {}
func (m *SimpleChunkManager) GetFullLightAt(x, y, z int) int                         { return 0 }
func (m *SimpleChunkManager) GetBlockLightAt(x, y, z int) int                        { return 0 }
func (m *SimpleChunkManager) GetRealBlockSkyLightAt(x, y, z int) int                 { return 0 }
func (m *SimpleChunkManager) GetSunAnglePercentage() float64                         { return 0 }
func (m *SimpleChunkManager) GetNearbyEntities(bb math.AxisAlignedBB) []block.Entity { return nil }
func (m *SimpleChunkManager) GetHighestAdjacentFullLightAt(x, y, z int) int          { return 0 }
func (m *SimpleChunkManager) GetHighestAdjacentBlockLightAt(x, y, z int) int         { return 0 }
func (m *SimpleChunkManager) GetPotentialLightAt(x, y, z int) int                    { return 0 }
func (m *SimpleChunkManager) UseBreakOn(pos math.Vector3) bool                       { return false }

var _ block.World = (*SimpleChunkManager)(nil)
