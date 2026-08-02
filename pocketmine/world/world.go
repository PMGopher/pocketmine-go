// Package world is a minimal, in-memory-only port of a slice of pocketmine\world\World.
package world

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/generator"
	"pocketmine-go/pocketmine/world/sound"
)

var _ block.World = (*World)(nil)

// YMin/YMax mirror pocketmine\world\World::Y_MIN/Y_MAX for the subchunk range format.Chunk
// actually stores (see format.Chunk's own MinSubChunkIndex/MaxSubChunkIndex).
const (
	YMin = format.MinSubChunkIndex * format.SubChunkEdgeLength
	YMax = (format.MaxSubChunkIndex + 1) * format.SubChunkEdgeLength
)

// World is a minimal port of pocketmine\world\World: enough to generate chunks on demand and let
// block.Behavior read/write blocks through the block.World interface it already codes against.
// Persistence (loading/saving to disk - this port has no WorldProvider), real light calculation
// (LightPopulationTask/BlockLightUpdate), entity tracking, and tile support are all documented
// gaps on the methods below, rather than silently wrong behaviour - every one returns an honest,
// clearly-labelled placeholder instead of guessing.
type World struct {
	generator  generator.Generator
	translator *convert.BlockTranslator

	chunks map[[2]int]*format.Chunk

	// stateTemplates lets GetBlockAt reconstruct a block.Behavior from the bare internal state ID
	// a Chunk stores (Chunk deliberately never keeps live Behavior instances around - see
	// format.Chunk's doc comment on why). Every block type the World needs to be able to read back
	// must be registered up front (see New's knownBlocks parameter); SetBlock also self-registers
	// whatever it's given, so anything ever placed through it becomes readable too.
	stateTemplates map[int32]block.Behavior
}

// New constructs a World using gen to generate chunks on demand. knownBlocks must include at
// least one Behavior for every distinct block type gen can place (e.g. for a Flat generator,
// every block.Behavior referenced by its []generator.FlatLayer) - see stateTemplates' doc comment
// for why.
func New(gen generator.Generator, translator *convert.BlockTranslator, knownBlocks []block.Behavior) *World {
	w := &World{
		generator:      gen,
		translator:     translator,
		chunks:         map[[2]int]*format.Chunk{},
		stateTemplates: map[int32]block.Behavior{},
	}
	for _, blk := range knownBlocks {
		w.registerTemplate(blk)
	}
	return w
}

// registerTemplate records blk as the reconstruction template for its state ID and warms the
// BlockTranslator's cache for the same ID (see BlockTranslator.NetworkIDForCachedState's doc
// comment on why chunk serialization needs that cache pre-warmed).
func (w *World) registerTemplate(blk block.Behavior) {
	stateID := int32(blk.GetStateId())
	if _, ok := w.stateTemplates[stateID]; !ok {
		w.stateTemplates[stateID] = blk.Clone()
	}
	w.translator.InternalIDToNetworkID(blk)
}

func chunkKey(chunkX, chunkZ int) [2]int { return [2]int{chunkX, chunkZ} }

// GetOrLoadChunk returns the chunk at the given chunk coordinates, generating it via the
// configured Generator on first access - this port has no on-disk world storage (no WorldProvider
// equivalent), so "load" always means "generate".
func (w *World) GetOrLoadChunk(chunkX, chunkZ int) *format.Chunk {
	key := chunkKey(chunkX, chunkZ)
	if c, ok := w.chunks[key]; ok {
		return c
	}
	c := w.generator.GenerateChunk(chunkX, chunkZ)
	w.chunks[key] = c
	return c
}

// Translator returns the BlockTranslator this World's chunks were populated through - the network
// chunk serializer needs it to translate a chunk's stored internal state IDs to Bedrock runtime
// IDs.
func (w *World) Translator() *convert.BlockTranslator { return w.translator }

// GetBlockAt is a port of World::getBlockAt (simplified: no chunk-load-failure/out-of-bounds
// handling - see IsInWorld). Falls back to air for a state ID with no registered template (see
// stateTemplates' doc comment) rather than panicking; this should never happen for a state this
// World itself ever wrote, only for data corruption.
func (w *World) GetBlockAt(x, y, z int) block.Behavior {
	chunk := w.GetOrLoadChunk(x>>4, z>>4)
	stateID := chunk.GetBlockStateID(x&0xf, y, z&0xf)
	tpl, ok := w.stateTemplates[stateID]
	if !ok {
		tpl = w.stateTemplates[int32(block.VanillaAir().GetStateId())]
	}
	got := tpl.Clone()
	got.(positionable).SetPosition(w, x, y, z)
	return got
}

// positionable is satisfied by every concrete block type via SetPosition's promotion from
// *block.Block - not part of block.Behavior itself (see this port's established convention for
// promoted-but-not-interface methods, e.g. block.asItemOrNil's identical reasoning for AsItem).
type positionable interface {
	SetPosition(world block.World, x, y, z int)
}

// SetBlock is a port of World::setBlock. Also registers blk as a state template (see
// registerTemplate) so it can be read back later even if it wasn't in New's knownBlocks list.
func (w *World) SetBlock(pos block.Position, blk block.Behavior) error {
	x, y, z := pos.FloorX(), pos.FloorY(), pos.FloorZ()
	w.registerTemplate(blk)
	chunk := w.GetOrLoadChunk(x>>4, z>>4)
	chunk.SetBlockStateID(x&0xf, y, z&0xf, int32(blk.GetStateId()))
	return nil
}

// GetTile/AddTile are ports of World::getTile/addTile. This port has no Tile-in-World system yet
// (see format.Chunk's doc comment on the same gap) - GetTile always reports "no tile", AddTile is
// a no-op.
func (w *World) GetTile(pos block.Position) (block.Tile, bool) { return nil, false }
func (w *World) AddTile(tile block.Tile)                       {}

// chunkAdapter satisfies block.Chunk (just SetBlockStateID(x,y,z,stateID int)) over a
// *format.Chunk (whose own SetBlockStateID takes int32 state IDs, matching PalettedBlockArray's
// storage type) - a thin type adapter, not a second implementation of anything.
type chunkAdapter struct{ chunk *format.Chunk }

func (a chunkAdapter) SetBlockStateID(x, y, z int, stateID int) {
	a.chunk.SetBlockStateID(x, y, z, int32(stateID))
}

// GetOrLoadChunkAtPosition is a port of World::getOrLoadChunkAtPosition.
func (w *World) GetOrLoadChunkAtPosition(pos block.Position) (block.Chunk, bool) {
	return chunkAdapter{w.GetOrLoadChunk(pos.FloorX()>>4, pos.FloorZ()>>4)}, true
}

// AddSound is a port of World::addSound. This port has no player-session/packet-broadcast system
// yet to actually deliver a sound packet to anyone, so this is a documented no-op rather than a
// guess at how broadcasting should work.
func (w *World) AddSound(pos math.Vector3, s sound.Sound) {}

// ScheduleDelayedBlockUpdate is a port of World::scheduleDelayedBlockUpdate. This port has no
// world tick scheduler yet (see pocketmine/scheduler, which exists but isn't wired to a running
// World), so this is a documented no-op.
func (w *World) ScheduleDelayedBlockUpdate(pos math.Vector3, delay int) {}

// GetFullLightAt/GetBlockLightAt/GetRealBlockSkyLightAt/GetHighestAdjacentFullLightAt/
// GetHighestAdjacentBlockLightAt/GetPotentialLightAt are ports of World's respective light-query
// methods. This port has no light engine yet (no LightPopulationTask/HeightMap-driven skylight,
// no BlockLightUpdate) - every position simply reports full light (15) rather than a guessed or
// zeroed value, so light-gated behaviour (crop growth, mob spawning rules, ...) defaults to
// "always allowed" until a real light engine exists.
func (w *World) GetFullLightAt(x, y, z int) int                 { return 15 }
func (w *World) GetBlockLightAt(x, y, z int) int                { return 15 }
func (w *World) GetRealBlockSkyLightAt(x, y, z int) int         { return 15 }
func (w *World) GetHighestAdjacentFullLightAt(x, y, z int) int  { return 15 }
func (w *World) GetHighestAdjacentBlockLightAt(x, y, z int) int { return 15 }
func (w *World) GetPotentialLightAt(x, y, z int) int            { return 15 }

// GetSunAnglePercentage is a port of World::getSunAnglePercentage. This port has no day/night
// cycle yet, so this reports a fixed midday value (0.5) rather than a guess.
func (w *World) GetSunAnglePercentage() float64 { return 0.5 }

// GetNearbyEntities is a port of World::getNearbyEntities. This port has no entity-in-world
// tracking yet, so this always reports no entities nearby.
func (w *World) GetNearbyEntities(bb math.AxisAlignedBB) []block.Entity { return nil }

// IsInWorld is a port of World::isInWorld, minus the PHP original's +/-30000000 horizontal bound
// (Bedrock's real world border, which nothing in this port enforces or needs yet - only the
// vertical bound is actually load-bearing for any current caller, e.g. Sugarcane.grow's upward
// climb).
func (w *World) IsInWorld(x, y, z int) bool { return y >= YMin && y < YMax }

// UseBreakOn is the simplified (no item/player/particles/drops) form of World::useBreakOn that
// block.World documents as the only form the block package itself needs - replaces the block with
// air and reports success unconditionally, matching the existing test doubles' behaviour
// throughout the block package's own test suite.
func (w *World) UseBreakOn(pos math.Vector3) bool {
	_ = w.SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, w), block.VanillaAir())
	return true
}
