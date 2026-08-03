package world

import (
	stdmath "math"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/format"
	worldio "pocketmine-go/pocketmine/world/format/io/leveldb"
)

// Time-of-day constants, a port of World::TIME_DAY/TIME_NOON/TIME_SUNSET/TIME_NIGHT/
// TIME_MIDNIGHT/TIME_SUNRISE/TIME_FULL.
const (
	TimeDay      = 1000
	TimeNoon     = 6000
	TimeSunset   = 12000
	TimeNight    = 13000
	TimeMidnight = 18000
	TimeSunrise  = 23000
	TimeFull     = 24000
)

// computeSunAnglePercentage is a port of World::computeSunAnglePercentage - see World.time's own
// doc comment on why t isn't normalized into [0, TimeFull) first: real PHP doesn't either, and t
// going negative only happens after time wraps at math.MaxInt64, an unreachable-in-practice edge
// case not worth diverging from the original formula's exact behaviour to special-case.
func computeSunAnglePercentage(t int64) float64 {
	timeProgress := float64(t%TimeFull) / TimeFull

	// 0.0 needs to be high noon, not dusk.
	sunProgress := timeProgress
	if timeProgress < 0.25 {
		sunProgress += 0.75
	} else {
		sunProgress -= 0.25
	}

	// Offset the sun progress to be above the horizon longer at dusk and dawn - roughly an
	// inverted sine curve, pushing the sun back at dusk and forwards at dawn.
	diff := ((1 - ((stdmath.Cos(sunProgress*stdmath.Pi) + 1) / 2)) - sunProgress) / 3
	return sunProgress + diff
}

// computeSkyLightReduction is a port of World::computeSkyLightReduction. Matches the PHP
// original's own "TODO: check rain and thunder level" - weather isn't factored in here either.
func computeSkyLightReduction(sunAnglePercentage float64) int {
	sunAngleRadians := sunAnglePercentage * 2 * stdmath.Pi
	percentage := stdmath.Max(0, stdmath.Min(1, -(stdmath.Cos(sunAngleRadians)*2-0.5)))
	return int(percentage * 11)
}

// IsDoingTick is a port of World::isDoingTick - true only for the duration of a DoTick call,
// letting callers (WorldManager.UnloadWorld) refuse to unload a world out from under its own
// currently-running tick.
func (w *World) IsDoingTick() bool { return w.doingTick }

// GetTime is a port of World::getTime.
func (w *World) GetTime() int64 { return w.time }

// GetTimeOfDay is a port of World::getTimeOfDay.
func (w *World) GetTimeOfDay() int64 { return w.time % TimeFull }

// SetTime is a port of World::setTime. Real PHP also calls sendTime() to broadcast the change to
// players - this port has no player-packet-broadcast wiring at the World level yet (matches
// AddSound's own documented gap), so that part is left undone rather than guessed at.
func (w *World) SetTime(t int64) { w.time = t }

// StopTime/StartTime/IsTimeStopped port World::stopTime/startTime/$stopTime - see SetTime's doc
// comment on the same missing sendTime() broadcast.
func (w *World) StopTime()           { w.stopTime = true }
func (w *World) StartTime()          { w.stopTime = false }
func (w *World) IsTimeStopped() bool { return w.stopTime }

// tryAddToNeighbourUpdateQueue is a port of World::tryAddToNeighbourUpdateQueue.
func (w *World) tryAddToNeighbourUpdateQueue(x, y, z int) {
	if !w.IsInWorld(x, y, z) {
		return
	}
	key := [3]int{x, y, z}
	if w.neighbourUpdateQueued[key] {
		return
	}
	w.neighbourUpdateQueued[key] = true
	w.neighbourUpdateQueue = append(w.neighbourUpdateQueue, key)
}

// internalNotifyNeighbourBlockUpdate is a port of World::internalNotifyNeighbourBlockUpdate.
func (w *World) internalNotifyNeighbourBlockUpdate(x, y, z int) {
	w.tryAddToNeighbourUpdateQueue(x, y, z)
	for _, side := range math.AllFacing {
		off := math.FacingOffset[side]
		w.tryAddToNeighbourUpdateQueue(x+off[0], y+off[1], z+off[2])
	}
}

// NotifyNeighbourBlockUpdate is a port of World::notifyNeighbourBlockUpdate.
func (w *World) NotifyNeighbourBlockUpdate(pos math.Vector3) {
	w.internalNotifyNeighbourBlockUpdate(pos.FloorX(), pos.FloorY(), pos.FloorZ())
}

// nearbyBlockChangeNotifiable is the optional surface an entity can implement to receive
// onNearbyBlockChange notifications (see updateNeighbourBlockUpdates) - declared locally since
// registeredEntity/block.Entity don't require it: pocketmine/entity has no concrete spawnable
// type yet to actually implement it (see registeredEntity's own doc comment on the same gap), so
// this is a forward-compatible optional interface, checked via type assertion, rather than a
// required method every future entity is forced to carry even if it never cares.
type nearbyBlockChangeNotifiable interface{ OnNearbyBlockChange() }

// updateNeighbourBlockUpdates is a port of the "Normal updates" loop in World::actuallyDoTick.
// Real PHP also fires a cancellable BlockUpdateEvent here - this port has no plugin/event system
// wired into World yet (matches every other "no event bus yet" gap elsewhere in this port), so
// every notification always proceeds, as if no plugin ever cancelled it.
func (w *World) updateNeighbourBlockUpdates() {
	for len(w.neighbourUpdateQueue) > 0 {
		pos := w.neighbourUpdateQueue[0]
		w.neighbourUpdateQueue = w.neighbourUpdateQueue[1:]
		delete(w.neighbourUpdateQueued, pos)

		x, y, z := pos[0], pos[1], pos[2]
		if _, ok := w.GetChunk(x>>4, z>>4); !ok {
			continue
		}

		bb := math.AxisAlignedBB{
			MinX: float64(x), MinY: float64(y), MinZ: float64(z),
			MaxX: float64(x + 1), MaxY: float64(y + 1), MaxZ: float64(z + 1),
		}
		for _, e := range w.GetNearbyEntities(bb) {
			if n, ok := e.(nearbyBlockChangeNotifiable); ok {
				n.OnNearbyBlockChange()
			}
		}

		w.GetBlockAt(x, y, z).OnNearbyBlockChange()
	}
}

// updateScheduledBlocks is a port of the "Delayed updates" loop in World::actuallyDoTick.
func (w *World) updateScheduledBlocks(currentTick int64) {
	var due, remaining []scheduledBlockUpdate
	for _, u := range w.scheduledUpdates {
		if u.dueTick <= currentTick {
			due = append(due, u)
		} else {
			remaining = append(remaining, u)
		}
	}
	w.scheduledUpdates = remaining

	for _, u := range due {
		delete(w.scheduledUpdateDelay, u.pos)
		x, y, z := u.pos[0], u.pos[1], u.pos[2]
		if _, ok := w.GetChunk(x>>4, z>>4); !ok {
			continue
		}
		w.GetBlockAt(x, y, z).OnScheduledUpdate()
	}
}

// RegisterChunkLoader is a port of World::registerChunkLoader. Real PocketMine-MP's ChunkLoader is
// a completely empty marker interface (see ChunkLoader.php - zero methods), so loader is typed
// `any` here and tracked purely by identity, exactly like PHP's own spl_object_id-keyed inner
// array - no method set is ever needed from it.
func (w *World) RegisterChunkLoader(loader any, chunkX, chunkZ int) {
	key := chunkKey(chunkX, chunkZ)
	delete(w.unloadQueue, key)
	if w.chunkLoaders[key] == nil {
		w.chunkLoaders[key] = map[any]bool{}
	}
	w.chunkLoaders[key][loader] = true
}

// UnregisterChunkLoader is a port of World::unregisterChunkLoader - queues the chunk for unloading
// (see unloadChunks) once its last loader is gone, exactly like real PHP's own $unloadQueue.
func (w *World) UnregisterChunkLoader(loader any, chunkX, chunkZ int) {
	key := chunkKey(chunkX, chunkZ)
	loaders, ok := w.chunkLoaders[key]
	if !ok {
		return
	}
	delete(loaders, loader)
	if len(loaders) == 0 {
		delete(w.chunkLoaders, key)
		w.unloadQueue[key] = w.currentTick
	}
}

// IsChunkInUse is a port of World::isChunkInUse.
func (w *World) IsChunkInUse(chunkX, chunkZ int) bool {
	loaders, ok := w.chunkLoaders[chunkKey(chunkX, chunkZ)]
	return ok && len(loaders) > 0
}

// RegisterTickingChunk/UnregisterTickingChunk port World::registerTickingChunk/
// unregisterTickingChunk. Real PocketMine-MP's ChunkTicker is likewise a completely empty marker
// interface (see ChunkTicker.php) - see RegisterChunkLoader's own doc comment on why `any` and
// identity-only tracking is the exact right shape here too.
//
// This port skips real PHP's validTickingChunks/recheckTickingChunks caching layer entirely - see
// isChunkTickable's own doc comment on why that's a safe simplification here, not a behavioural
// shortcut.
func (w *World) RegisterTickingChunk(loader any, chunkX, chunkZ int) {
	key := chunkKey(chunkX, chunkZ)
	if w.tickingChunks[key] == nil {
		w.tickingChunks[key] = map[any]bool{}
	}
	w.tickingChunks[key][loader] = true
}

func (w *World) UnregisterTickingChunk(loader any, chunkX, chunkZ int) {
	key := chunkKey(chunkX, chunkZ)
	loaders, ok := w.tickingChunks[key]
	if !ok {
		return
	}
	delete(loaders, loader)
	if len(loaders) == 0 {
		delete(w.tickingChunks, key)
	}
}

// isChunkTickable is a port of World::isChunkTickable's actual eligibility check (chunk + its 8
// neighbours all loaded, populated and light-populated), minus the validTickingChunks/
// recheckTickingChunks caching machinery real PHP wraps around it.
//
// That caching exists in PHP purely to avoid re-running this check every tick while a chunk's
// neighbourhood is mid-async-light-population (LightPopulationTask runs on a worker thread there,
// so a chunk can sit in a "generated, not yet light-populated" state for an unpredictable stretch
// of real time). This port's whole generate -> populate -> light-populate pipeline runs
// synchronously, inline, in ensurePopulated (see its own doc comment) - by the time any chunk is
// reachable via GetChunk at all, it is already fully generated, populated, and light-populated.
// There is no async gap for a recheck-queue to paper over, so the check can just run directly
// every tick with identical results and no correctness cost - a legitimate architectural
// simplification of PHP's own internal bookkeeping, not a difference in observable behaviour.
func (w *World) isChunkTickable(chunkX, chunkZ int) bool {
	for cx := -1; cx <= 1; cx++ {
		for cz := -1; cz <= 1; cz++ {
			chunk, ok := w.GetChunk(chunkX+cx, chunkZ+cz)
			if !ok || !chunk.IsPopulated() {
				return false
			}
			if lit, known := chunk.IsLightPopulated(); !known || !lit {
				return false
			}
		}
	}
	return true
}

// tickedBlocksPerSubchunkPerTick mirrors World::DEFAULT_TICKED_BLOCKS_PER_SUBCHUNK_PER_TICK.
const tickedBlocksPerSubchunkPerTick = 3

// tickChunk is a port of World::tickChunk's random-tick block sampling. Real PHP also ticks
// per-chunk entities here (foreach($this->getChunkEntities(...) as $entity) $entity->
// onRandomUpdate()) - this port's flat entity registry (see registeredEntity's doc comment) has no
// per-chunk index to iterate the same way, and no concrete Entity type implements onRandomUpdate
// yet regardless, so that part is a documented gap rather than a guess.
//
// The 60-bit-random/12-bits-per-axis decoding is ported exactly (x = k&0xf, y = (k>>4)&0xf,
// z = (k>>8)&0xf, refilled every 5th iteration) - only the random source itself differs (this
// port's own math/rand generator, not PHP's mt_rand stream), matching this port's established
// precedent of not replicating PHP's specific PRNG byte-for-byte where nothing depends on doing so
// (see e.g. biome selection's own noise-mixing doc comment).
func (w *World) tickChunk(chunkX, chunkZ int) {
	chunk, ok := w.GetChunk(chunkX, chunkZ)
	if !ok {
		// The chunk may have been unloaded during a previous chunk's update in this same tick.
		return
	}

	for subY, subChunk := range chunk.GetSubChunks() {
		if subChunk.IsEmptyFast() {
			continue
		}

		var k int64
		for i := 0; i < tickedBlocksPerSubchunkPerTick; i++ {
			if i%5 == 0 {
				// 60 bits will be used by 5 blocks (12 bits each).
				k = w.rng.Int63n(1 << 60)
			}
			x := int(k & 0xf)
			y := int((k >> 4) & 0xf)
			z := int((k >> 8) & 0xf)
			k >>= 12

			state := subChunk.GetBlockStateID(x, y, z)
			if !w.randomTickBlocks[state] {
				continue
			}
			tpl, ok := w.stateTemplates[state]
			if !ok {
				continue
			}

			blk := tpl.Clone()
			worldX := chunkX*format.SubChunkEdgeLength + x
			worldY := subY*format.SubChunkEdgeLength + y
			worldZ := chunkZ*format.SubChunkEdgeLength + z
			blk.(positionable).SetPosition(w, worldX, worldY, worldZ)
			blk.OnRandomTick()
		}
	}
}

// tickChunks is a port of World::tickChunks (minus the recheck-cache layer - see isChunkTickable's
// doc comment).
func (w *World) tickChunks() {
	if w.chunkTickRadius <= 0 || len(w.tickingChunks) == 0 {
		return
	}
	for key := range w.tickingChunks {
		if w.isChunkTickable(key[0], key[1]) {
			w.tickChunk(key[0], key[1])
		}
	}
}

// unloadChunk is a port of World::unloadChunk (the safe=true branch only - this port never forces
// an in-use chunk to unload, matching the only way real PHP's own unloadChunks ever calls it).
func (w *World) unloadChunk(chunkX, chunkZ int, safe bool) bool {
	key := chunkKey(chunkX, chunkZ)
	chunk, ok := w.chunks[key]
	if !ok {
		return true
	}
	if safe && w.IsChunkInUse(chunkX, chunkZ) {
		return false
	}

	if w.provider != nil {
		if err := worldio.SaveChunk(w.provider, int32(chunkX), int32(chunkZ), chunk, w.lookupBlockState); err != nil {
			return false
		}
	}

	for _, listener := range w.GetChunkListeners(chunkX, chunkZ) {
		listener.OnChunkUnloaded(chunkX, chunkZ, chunk)
	}

	chunk.OnUnload()
	delete(w.chunks, key)
	delete(w.populated, key)
	return true
}

// maxChunkUnloadsPerTick mirrors World::unloadChunks' own $maxUnload = 96 rate limit.
const maxChunkUnloadsPerTick = 96

// unloadChunks is a port of World::unloadChunks: a chunk sits in the unload queue for
// unloadGraceTicks (see its own doc comment) after losing its last loader before actually being
// unloaded, giving a player a grace window to walk back into a chunk they just left the edge of
// without it being torn down and regenerated on the spot.
func (w *World) unloadChunks() {
	if len(w.unloadQueue) == 0 {
		return
	}
	unloaded := 0
	for key, queuedAtTick := range w.unloadQueue {
		if unloaded >= maxChunkUnloadsPerTick {
			break
		}
		if w.currentTick-queuedAtTick < unloadGraceTicks {
			continue
		}
		if w.unloadChunk(key[0], key[1], true) {
			delete(w.unloadQueue, key)
			unloaded++
		}
	}
}

// DoTick is a port of World::doTick/actuallyDoTick: the per-tick update real PocketMine-MP's
// Server calls on every loaded World once per game tick. currentTick is the server's own
// monotonically increasing tick counter (matching the same parameter in the PHP original), used
// both for scheduled-update due-time comparisons and as the "when did this chunk lose its last
// loader" timestamp unloadChunks measures its grace window against.
//
// Not ported: sendTime()/provider garbage collection (both are pure network/disk-housekeeping
// concerns with nothing behavioural to get wrong by omitting), and the entity-tick pass (see
// tickChunk's own doc comment on why - no concrete Entity type exists yet to tick).
func (w *World) DoTick(currentTick int64) {
	w.doingTick = true
	defer func() { w.doingTick = false }()

	w.currentTick = currentTick

	if !w.stopTime {
		w.time++
	}
	w.sunAnglePercentage = computeSunAnglePercentage(w.time)
	w.skyLightReduction = computeSkyLightReduction(w.sunAnglePercentage)

	w.unloadChunks()

	w.updateScheduledBlocks(currentTick)
	w.updateNeighbourBlockUpdates()

	w.tickChunks()

	// Matches actuallyDoTick's own ordering: queued light recalculation (from SetBlock's
	// RecalculateNode calls made since the last tick, plus anything tickChunk/updateScheduledBlocks/
	// updateNeighbourBlockUpdates triggered just now) is only actually executed once per tick, here
	// at the end - not immediately on every individual block change.
	w.blockLightUpdate.Execute()
	w.skyLightUpdate.Execute()
}
