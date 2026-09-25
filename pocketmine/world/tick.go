package world

import (
	stdmath "math"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

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

// SetTime is a port of World::setTime.
func (w *World) SetTime(t int64) {
	w.time = t
	w.SendTime()
}

// StopTime/StartTime/IsTimeStopped port World::stopTime/startTime/$stopTime.
func (w *World) StopTime() {
	w.stopTime = true
	w.SendTime()
}

func (w *World) StartTime() {
	w.stopTime = false
	w.SendTime()
}

func (w *World) IsTimeStopped() bool { return w.stopTime }

// SendTime is a port of World::sendTime: every player in the world (NetworkSession::syncWorldTime)
// gets the time. Players are the entities that are packet viewers (see viewer).
func (w *World) SendTime() {
	pk := &packet.SetTime{Time: int32(w.time)}
	for _, e := range w.entities {
		if v, ok := e.(viewer); ok {
			v.SendPacket(pk)
		}
	}
}

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

// updateNeighbourBlockUpdates is a port of the "Normal updates" loop in World::actuallyDoTick.
func (w *World) updateNeighbourBlockUpdates() {
	for len(w.neighbourUpdateQueue) > 0 {
		pos := w.neighbourUpdateQueue[0]
		w.neighbourUpdateQueue = w.neighbourUpdateQueue[1:]
		delete(w.neighbourUpdateQueued, pos)

		x, y, z := pos[0], pos[1], pos[2]
		if _, ok := w.GetChunk(x>>4, z>>4); !ok {
			continue
		}

		blk := w.GetBlockAt(x, y, z)
		if event.HasHandlers[blockevent.BlockUpdateEvent]() {
			ev := blockevent.NewBlockUpdateEvent(blk)
			event.Call(ev)
			if ev.IsCancelled() {
				continue
			}
		}

		bb := math.AxisAlignedBB{
			MinX: float64(x), MinY: float64(y), MinZ: float64(z),
			MaxX: float64(x + 1), MaxY: float64(y + 1), MaxZ: float64(z + 1),
		}
		for _, e := range w.GetNearbyEntitiesExcept(bb, nil) {
			e.OnNearbyBlockChange()
		}

		blk.OnNearbyBlockChange()
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
		w.UnloadChunkRequest(chunkX, chunkZ, true)
		if resolver, ok := w.chunkPopulationRequestMap[key]; ok {
			if _, active := w.activeChunkPopulationTasks[key]; !active {
				resolver.Reject()
				delete(w.chunkPopulationRequestMap, key)
			}
		}
	}
}

// IsChunkInUse is a port of World::isChunkInUse.
func (w *World) IsChunkInUse(chunkX, chunkZ int) bool {
	loaders, ok := w.chunkLoaders[chunkKey(chunkX, chunkZ)]
	return ok && len(loaders) > 0
}

// GetChunkTickRadius is a port of World::getChunkTickRadius.
func (w *World) GetChunkTickRadius() int { return w.chunkTickRadius }

// SetChunkTickRadius is a port of World::setChunkTickRadius.
func (w *World) SetChunkTickRadius(radius int) { w.chunkTickRadius = radius }

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

// isChunkTickable is a port of World::isChunkTickable: the chunk and its 8 neighbours must be
// loaded, populated, unlocked and lit. A neighbour whose light hasn't been calculated yet gets it
// ordered (orderLightPopulation). PHP caches the result per tick and rechecks only chunks marked
// for it; this port checks the registered ticking chunks every tick instead.
func (w *World) isChunkTickable(chunkX, chunkZ int) bool {
	for cx := -1; cx <= 1; cx++ {
		for cz := -1; cz <= 1; cz++ {
			if w.IsChunkLocked(chunkX+cx, chunkZ+cz) {
				return false
			}
			chunk, ok := w.GetChunk(chunkX+cx, chunkZ+cz)
			if !ok || !chunk.IsPopulated() {
				return false
			}
			lit, known := chunk.IsLightPopulated()
			if !known || !lit {
				if known && !lit {
					w.orderLightPopulation(chunkX+cx, chunkZ+cz)
				}
				return false
			}
		}
	}
	return true
}

// tickedBlocksPerSubchunkPerTick mirrors World::DEFAULT_TICKED_BLOCKS_PER_SUBCHUNK_PER_TICK.
const tickedBlocksPerSubchunkPerTick = 3

// tickChunk is a port of World::tickChunk: random updates for the chunk's entities, then
// random-tick block sampling.
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
	for _, entity := range w.GetChunkEntities(chunkX, chunkZ) {
		entity.OnRandomUpdate()
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
		if err := worldio.SaveEntities(w.provider, int32(chunkX), int32(chunkZ), w.saveChunkEntities(chunkX, chunkZ)); err != nil {
			return false
		}
	}

	for _, listener := range w.GetChunkListeners(chunkX, chunkZ) {
		listener.OnChunkUnloaded(chunkX, chunkZ, chunk)
	}

	w.closeChunkEntities(chunkX, chunkZ)

	chunk.OnUnload()
	delete(w.chunks, key)
	delete(w.changedBlocks, key)
	w.rejectPopulationRequest(chunkX, chunkZ)
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
// concerns with nothing behavioural to get wrong by omitting).
func (w *World) DoTick(currentTick int64) {
	w.doingTick = true
	defer func() { w.doingTick = false }()

	w.currentTick = currentTick

	if !w.stopTime {
		w.time++
	}
	w.sendTimeTicker++
	if w.sendTimeTicker == 200 {
		w.SendTime()
		w.sendTimeTicker = 0
	}
	w.sunAnglePercentage = computeSunAnglePercentage(w.time)
	w.skyLightReduction = computeSkyLightReduction(w.sunAnglePercentage)

	w.unloadChunks()

	w.updateScheduledBlocks(currentTick)
	w.updateNeighbourBlockUpdates()

	w.tickEntities(currentTick)

	w.tickChunks()

	// Matches actuallyDoTick's own ordering: queued light recalculation (from SetBlock's
	// RecalculateNode calls made since the last tick, plus anything tickChunk/updateScheduledBlocks/
	// updateNeighbourBlockUpdates triggered just now) is only actually executed once per tick, here
	// at the end - not immediately on every individual block change.
	w.blockLightUpdate.Execute()
	w.skyLightUpdate.Execute()

	w.sendChangedBlocks()

	if w.sleepTicks > 0 {
		w.sleepTicks--
		if w.sleepTicks <= 0 {
			w.CheckSleep()
		}
	}
}

// sleeper is the part of Player the sleep check needs.
type sleeper interface {
	IsSleeping() bool
	StopSleep()
}

// SetSleepTicks is a port of World::setSleepTicks: the sleep check runs after this many ticks.
func (w *World) SetSleepTicks(ticks int) { w.sleepTicks = ticks }

// CheckSleep is a port of World::checkSleep: when every player is asleep at night, the time skips
// to the next morning and everyone wakes up.
func (w *World) CheckSleep() {
	players := w.GetPlayers()
	if len(players) == 0 {
		return
	}

	for _, p := range players {
		if s, ok := p.(sleeper); !ok || !s.IsSleeping() {
			return
		}
	}

	time := w.GetTimeOfDay()
	if time >= TimeNight && time < TimeSunrise {
		w.SetTime(w.GetTime() + TimeFull - time)
		for _, p := range players {
			p.(sleeper).StopSleep()
		}
	}
}

// sendChangedBlocks is the changedBlocks part of World::actuallyDoTick: every block set this tick
// is sent to the players using its chunk, or the whole chunk is resent (Player::onChunkChanged)
// when more than 512 blocks changed in it.
func (w *World) sendChangedBlocks() {
	if len(w.changedBlocks) == 0 {
		return
	}
	for chunkPos, blocks := range w.changedBlocks {
		if len(blocks) == 0 { //blocks can be set normally and then later re-set with direct send
			continue
		}
		chunk, ok := w.GetChunk(chunkPos[0], chunkPos[1])
		if !ok {
			//a previous chunk may have caused this one to be unloaded by a ChunkListener
			continue
		}
		if len(blocks) > 512 {
			for _, l := range w.GetChunkListeners(chunkPos[0], chunkPos[1]) {
				if _, isPlayer := l.(viewer); isPlayer {
					l.OnChunkChanged(chunkPos[0], chunkPos[1], chunk)
				}
			}
			continue
		}
		positions := make([]math.Vector3, 0, len(blocks))
		for _, pos := range blocks {
			positions = append(positions, pos)
		}
		for _, pk := range w.CreateBlockUpdatePackets(positions) {
			w.broadcastPacketToPlayersUsingChunk(chunkPos[0], chunkPos[1], pk)
		}
	}
	w.changedBlocks = nil
}

// CreateBlockUpdatePackets is a port of World::createBlockUpdatePackets. Not ported: the tile
// parts (the render-update workaround state and BlockActorDataPacket), since tiles aren't sent to
// clients yet.
func (w *World) CreateBlockUpdatePackets(blocks []math.Vector3) []packet.Packet {
	packets := make([]packet.Packet, 0, len(blocks))
	for _, b := range blocks {
		x, y, z := b.FloorX(), b.FloorY(), b.FloorZ()
		stateID := int32(block.VanillaAir().GetStateId())
		if chunk := w.generateChunkOnly(x>>4, z>>4); chunk != nil {
			stateID = chunk.GetBlockStateID(x&0xf, y, z&0xf)
		}
		packets = append(packets, &packet.UpdateBlock{
			Position:          protocol.BlockPos{int32(x), int32(y), int32(z)},
			NewBlockRuntimeID: uint32(w.translator.NetworkIDForCachedState(stateID)),
			Flags:             packet.BlockUpdateNetwork,
			Layer:             0, // UpdateBlockPacket::DATA_LAYER_NORMAL
		})
	}
	return packets
}

// broadcastPacketToPlayersUsingChunk is a port of World::broadcastPacketToPlayersUsingChunk.
func (w *World) broadcastPacketToPlayersUsingChunk(chunkX, chunkZ int, pk packet.Packet) {
	for _, l := range w.GetChunkListeners(chunkX, chunkZ) {
		if v, ok := l.(viewer); ok {
			v.SendPacket(pk)
		}
	}
}

// UnloadChunkRequest is a port of World::unloadChunkRequest: queues the chunk for unloading unless
// it's in use (when safe) or a spawn chunk.
func (w *World) UnloadChunkRequest(chunkX, chunkZ int, safe bool) bool {
	if (safe && w.IsChunkInUse(chunkX, chunkZ)) || w.IsSpawnChunk(chunkX, chunkZ) {
		return false
	}
	w.unloadQueue[chunkKey(chunkX, chunkZ)] = w.currentTick
	return true
}

// DoChunkGarbageCollection is a port of World::doChunkGarbageCollection. The provider's own
// garbage collection has nothing to do for LevelDB here.
func (w *World) DoChunkGarbageCollection() {
	for key, chunk := range w.chunks {
		if _, queued := w.unloadQueue[key]; !queued {
			if !w.IsSpawnChunk(key[0], key[1]) {
				w.UnloadChunkRequest(key[0], key[1], true)
			}
		}
		chunk.CollectGarbage()
	}
}

// UnloadChunks is World::unloadChunks($force): with force, every queued chunk that can be unloaded
// is, without waiting for its grace period.
func (w *World) UnloadChunks(force bool) {
	if !force {
		w.unloadChunks()
		return
	}
	for key := range w.unloadQueue {
		//If the chunk can't be unloaded, it stays on the queue
		if w.unloadChunk(key[0], key[1], true) {
			delete(w.unloadQueue, key)
		}
	}
}
