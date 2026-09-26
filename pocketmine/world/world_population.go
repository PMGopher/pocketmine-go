package world

import (
	"fmt"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/event"
	worldevent "pocketmine-go/pocketmine/event/world"
	"pocketmine-go/pocketmine/promise"
	"pocketmine-go/pocketmine/scheduler"
	"pocketmine-go/pocketmine/world/format"
)

// This file ports World's chunk loading, locking and asynchronous population: loadChunk,
// setChunk, lockChunk/unlockChunk, requestChunkPopulation/orderChunkPopulation (run by a
// GeneratorExecutor, normally on the AsyncPool), generateChunkCallback and orderLightPopulation.

// ChunkLockID is a port of pocketmine\world\ChunkLockId: identifies who holds a chunk lock.
type ChunkLockID struct{ _ byte }

// temporaryChunkLoader is the anonymous `new class implements ChunkLoader{}` PHP registers while a
// chunk is being loaded or populated (it has a field so every instance has its own address).
type temporaryChunkLoader struct{ _ byte }

// defaultMaxConcurrentChunkPopulationTasks is World::$maxConcurrentChunkPopulationTasks' default
// (pocketmine.yml chunk-generation.population-queue-size overrides it).
const defaultMaxConcurrentChunkPopulationTasks = 2

// SetMaxConcurrentChunkPopulationTasks sets pocketmine.yml's chunk-generation.population-queue-size.
func (w *World) SetMaxConcurrentChunkPopulationTasks(n int) {
	w.maxConcurrentChunkPopulationTasks = max(1, n)
}

// SetGeneratorExecutor sets how population runs (World::__construct picks the AsyncGeneratorExecutor
// unless the generator is fast). workerPool is also used for light population; nil runs light
// population synchronously.
func (w *World) SetGeneratorExecutor(executor GeneratorExecutor, workerPool *scheduler.AsyncPool) {
	w.generatorExecutor = executor
	w.workerPool = workerPool
}

// generatorSetupParameters is the SyncGeneratorExecutor's setup for this world's own generator.
func (w *World) generatorSetupParameters() GeneratorExecutorSetupParameters {
	return GeneratorExecutorSetupParameters{WorldMinY: YMin, WorldMaxY: YMax}
}

// executor returns the world's GeneratorExecutor, a SyncGeneratorExecutor if none was set.
func (w *World) executor() GeneratorExecutor {
	if w.generatorExecutor == nil {
		w.generatorExecutor = NewSyncGeneratorExecutor(w.generatorSetupParameters(), w.generator)
	}
	return w.generatorExecutor
}

// loadChunk is a port of World::loadChunk: the loaded chunk, or the one saved on disk, or nil if
// the chunk has never been generated. It never generates.
func (w *World) loadChunk(chunkX, chunkZ int) *format.Chunk {
	key := chunkKey(chunkX, chunkZ)
	if c, ok := w.chunks[key]; ok {
		return c
	}
	if w.knownUngeneratedChunks[key] {
		return nil
	}
	delete(w.unloadQueue, key) // cancelUnloadChunkRequest

	if w.provider == nil {
		w.knownUngeneratedChunks[key] = true
		return nil
	}
	loadedChunkData, err := w.provider.LoadChunk(chunkX, chunkZ)
	if err != nil && w.logger != nil {
		w.logger.Critical(fmt.Sprintf("Failed to load chunk x=%d z=%d: %v", chunkX, chunkZ, err))
	}
	if err != nil || loadedChunkData == nil {
		w.knownUngeneratedChunks[key] = true
		return nil
	}

	chunkData := loadedChunkData.GetData()
	c := format.NewChunk(chunkData.GetSubChunks(), chunkData.IsPopulated(), int32(block.VanillaAir().GetStateId()), 0)
	if !loadedChunkData.IsUpgraded() {
		c.ClearTerrainDirtyFlags()
	} else if w.logger != nil {
		w.logger.Debug(fmt.Sprintf("Chunk %d %d has been upgraded, will be saved at the next autosave opportunity", chunkX, chunkZ))
	}
	w.chunks[key] = c
	delete(w.changedBlocks, key)

	w.initChunk(chunkX, chunkZ, chunkData, c)
	if event.HasHandlers[worldevent.ChunkLoadEvent]() {
		event.Call(worldevent.NewChunkLoadEvent(w, chunkX, chunkZ, c, false))
	}
	if !w.IsChunkInUse(chunkX, chunkZ) {
		w.UnloadChunkRequest(chunkX, chunkZ, true)
	}
	w.fireOnChunkLoaded(chunkX, chunkZ, c)
	return c
}

// LoadChunk is World::loadChunk for other packages.
func (w *World) LoadChunk(chunkX, chunkZ int) (*format.Chunk, bool) {
	c := w.loadChunk(chunkX, chunkZ)
	return c, c != nil
}

// SetChunk is a port of World::setChunk: replaces (or adds) the chunk, keeping tiles the new chunk
// still has a block for.
func (w *World) SetChunk(chunkX, chunkZ int, chunk *format.Chunk) {
	key := chunkKey(chunkX, chunkZ)
	oldChunk := w.loadChunk(chunkX, chunkZ)
	if oldChunk != nil && oldChunk != chunk {
		for _, oldTile := range oldChunk.GetTiles() {
			tilePos := oldTile.GetPosition()
			localX, localY, localZ := tilePos.FloorX()&0xf, tilePos.FloorY(), tilePos.FloorZ()&0xf
			newTile, hasNewTile := chunk.GetTile(localX, localY, localZ)
			// PHP compares the new block's expected tile class; this port's tiles don't record it,
			// so a tile survives when the block at its position is unchanged.
			sameBlock := chunk.GetBlockStateID(localX, localY, localZ) == oldChunk.GetBlockStateID(localX, localY, localZ)
			if !sameBlock || (hasNewTile && newTile != oldTile) {
				oldTile.Close()
			} else {
				chunk.AddTile(oldTile)
				oldChunk.RemoveTile(oldTile)
			}
		}
	}

	w.chunks[key] = chunk
	delete(w.knownUngeneratedChunks, key)
	delete(w.changedBlocks, key)
	chunk.SetTerrainDirty()

	if !w.IsChunkInUse(chunkX, chunkZ) {
		w.UnloadChunkRequest(chunkX, chunkZ, true)
	}

	if oldChunk == nil {
		if event.HasHandlers[worldevent.ChunkLoadEvent]() {
			event.Call(worldevent.NewChunkLoadEvent(w, chunkX, chunkZ, chunk, true))
		}
		w.fireOnChunkLoaded(chunkX, chunkZ, chunk)
	} else {
		for _, listener := range w.GetChunkListeners(chunkX, chunkZ) {
			listener.OnChunkChanged(chunkX, chunkZ, chunk)
		}
	}

	for cx := -1; cx <= 1; cx++ {
		for cz := -1; cz <= 1; cz++ {
			for _, e := range w.GetChunkEntities(chunkX+cx, chunkZ+cz) {
				e.OnNearbyBlockChange()
			}
		}
	}
}

// LockChunk is a port of World::lockChunk: flags a chunk as locked, usually for async
// modification. This is an advisory lock: modifying the chunk on the main thread breaks it, which
// UnlockChunk with the same lock ID then reports.
func (w *World) LockChunk(chunkX, chunkZ int, lockID *ChunkLockID) {
	key := chunkKey(chunkX, chunkZ)
	if _, locked := w.chunkLock[key]; locked {
		panic(fmt.Sprintf("Chunk %d %d is already locked", chunkX, chunkZ))
	}
	w.chunkLock[key] = lockID
}

// UnlockChunk is a port of World::unlockChunk: a nil lockID removes any lock. It reports whether
// unlocking was successful.
func (w *World) UnlockChunk(chunkX, chunkZ int, lockID *ChunkLockID) bool {
	key := chunkKey(chunkX, chunkZ)
	if current, locked := w.chunkLock[key]; locked && (lockID == nil || current == lockID) {
		delete(w.chunkLock, key)
		return true
	}
	return false
}

// IsChunkLocked is a port of World::isChunkLocked.
func (w *World) IsChunkLocked(chunkX, chunkZ int) bool {
	_, locked := w.chunkLock[chunkKey(chunkX, chunkZ)]
	return locked
}

// getAdjacentChunks is a port of World::getAdjacentChunks: relative [dx, dz] => chunk (nil if not
// generated), without the centre.
func (w *World) getAdjacentChunks(chunkX, chunkZ int) map[[2]int]*format.Chunk {
	result := map[[2]int]*format.Chunk{}
	for xx := -1; xx <= 1; xx++ {
		for zz := -1; zz <= 1; zz++ {
			if xx == 0 && zz == 0 {
				continue //center chunk
			}
			result[[2]int{xx, zz}] = w.loadChunk(chunkX+xx, chunkZ+zz)
		}
	}
	return result
}

func (w *World) addChunkHashToPopulationRequestQueue(key [2]int) {
	if !w.chunkPopulationRequestQueueIndex[key] {
		w.chunkPopulationRequestQueue = append(w.chunkPopulationRequestQueue, key)
		w.chunkPopulationRequestQueueIndex[key] = true
	}
}

// enqueuePopulationRequest is a port of World::enqueuePopulationRequest.
func (w *World) enqueuePopulationRequest(chunkX, chunkZ int, associatedChunkLoader any) *promise.Promise[*format.Chunk] {
	key := chunkKey(chunkX, chunkZ)
	w.addChunkHashToPopulationRequestQueue(key)
	resolver := promise.NewResolver[*format.Chunk]()
	w.chunkPopulationRequestMap[key] = resolver
	if associatedChunkLoader == nil {
		temporaryLoader := &temporaryChunkLoader{}
		w.RegisterChunkLoader(temporaryLoader, chunkX, chunkZ)
		resolver.GetPromise().OnCompletion(
			func(*format.Chunk) { w.UnregisterChunkLoader(temporaryLoader, chunkX, chunkZ) },
			func() {},
		)
	}
	return resolver.GetPromise()
}

// drainPopulationRequestQueue is a port of World::drainPopulationRequestQueue.
func (w *World) drainPopulationRequestQueue() {
	var failed [][2]int
	for len(w.activeChunkPopulationTasks) < w.maxConcurrentChunkPopulationTasks && len(w.chunkPopulationRequestQueue) > 0 {
		next := w.chunkPopulationRequestQueue[0]
		w.chunkPopulationRequestQueue = w.chunkPopulationRequestQueue[1:]
		delete(w.chunkPopulationRequestQueueIndex, next)
		if _, requested := w.chunkPopulationRequestMap[next]; requested {
			if _, active := w.activeChunkPopulationTasks[next]; active {
				panic(fmt.Sprintf("Population for chunk %d %d already running", next[0], next[1]))
			}
			if !w.OrderChunkPopulation(next[0], next[1], nil).IsResolved() {
				if _, active := w.activeChunkPopulationTasks[next]; !active {
					failed = append(failed, next)
				}
			}
		}
	}

	//these requests failed even though they weren't rate limited; we can't directly re-add them to the back of the
	//queue because it would result in an infinite loop
	for _, key := range failed {
		w.addChunkHashToPopulationRequestQueue(key)
	}
}

// checkChunkPopulationPreconditions is a port of World::checkChunkPopulationPreconditions: the
// existing request's resolver (if any) and whether population should go ahead.
func (w *World) checkChunkPopulationPreconditions(chunkX, chunkZ int) (*promise.Resolver[*format.Chunk], bool) {
	key := chunkKey(chunkX, chunkZ)
	resolver := w.chunkPopulationRequestMap[key]
	if _, active := w.activeChunkPopulationTasks[key]; resolver != nil && active {
		//generation is already running
		return resolver, false
	}

	temporaryLoader := &temporaryChunkLoader{}
	w.RegisterChunkLoader(temporaryLoader, chunkX, chunkZ)
	chunk := w.loadChunk(chunkX, chunkZ)
	w.UnregisterChunkLoader(temporaryLoader, chunkX, chunkZ)
	if chunk != nil && chunk.IsPopulated() {
		//chunk is already populated; return a pre-resolved promise that will directly fire callbacks assigned
		if resolver == nil {
			resolver = promise.NewResolver[*format.Chunk]()
		}
		delete(w.chunkPopulationRequestMap, key)
		resolver.Resolve(chunk)
		return resolver, false
	}
	return resolver, true
}

// RequestChunkPopulation is a port of World::requestChunkPopulation: starts asynchronous
// generation/population of the chunk if it's reasonable to do so now (and it isn't already
// generated/populated); if the generator is busy, the request is queued. associatedChunkLoader
// (nil for none) cancels the request once no loader is left on the chunk.
func (w *World) RequestChunkPopulation(chunkX, chunkZ int, associatedChunkLoader any) *promise.Promise[*format.Chunk] {
	resolver, proceed := w.checkChunkPopulationPreconditions(chunkX, chunkZ)
	if !proceed {
		if resolver != nil {
			return resolver.GetPromise()
		}
		return w.enqueuePopulationRequest(chunkX, chunkZ, associatedChunkLoader)
	}

	if len(w.activeChunkPopulationTasks) >= w.maxConcurrentChunkPopulationTasks {
		//too many chunks are already generating; delay resolution of the request until later
		if resolver != nil {
			return resolver.GetPromise()
		}
		return w.enqueuePopulationRequest(chunkX, chunkZ, associatedChunkLoader)
	}
	return w.internalOrderChunkPopulation(chunkX, chunkZ, associatedChunkLoader, resolver)
}

// OrderChunkPopulation is a port of World::orderChunkPopulation: like RequestChunkPopulation, but
// ignoring the concurrency limit.
func (w *World) OrderChunkPopulation(chunkX, chunkZ int, associatedChunkLoader any) *promise.Promise[*format.Chunk] {
	resolver, proceed := w.checkChunkPopulationPreconditions(chunkX, chunkZ)
	if !proceed {
		if resolver != nil {
			return resolver.GetPromise()
		}
		return w.enqueuePopulationRequest(chunkX, chunkZ, associatedChunkLoader)
	}
	return w.internalOrderChunkPopulation(chunkX, chunkZ, associatedChunkLoader, resolver)
}

// internalOrderChunkPopulation is a port of World::internalOrderChunkPopulation.
func (w *World) internalOrderChunkPopulation(chunkX, chunkZ int, associatedChunkLoader any, resolver *promise.Resolver[*format.Chunk]) *promise.Promise[*format.Chunk] {
	key := chunkKey(chunkX, chunkZ)
	for xx := -1; xx <= 1; xx++ {
		for zz := -1; zz <= 1; zz++ {
			if w.IsChunkLocked(chunkX+xx, chunkZ+zz) {
				//chunk is already in use by another generation request; queue the request for later
				if resolver != nil {
					return resolver.GetPromise()
				}
				return w.enqueuePopulationRequest(chunkX, chunkZ, associatedChunkLoader)
			}
		}
	}

	w.activeChunkPopulationTasks[key] = true
	if resolver == nil {
		resolver = promise.NewResolver[*format.Chunk]()
		w.chunkPopulationRequestMap[key] = resolver
	}

	lockID := &ChunkLockID{}
	temporaryLoader := &temporaryChunkLoader{}
	for xx := -1; xx <= 1; xx++ {
		for zz := -1; zz <= 1; zz++ {
			w.LockChunk(chunkX+xx, chunkZ+zz, lockID)
			w.RegisterChunkLoader(temporaryLoader, chunkX+xx, chunkZ+zz)
		}
	}

	// The worker gets copies (PHP serializes them into the task).
	var center *format.Chunk
	if c := w.loadChunk(chunkX, chunkZ); c != nil {
		center = c.Clone()
	}
	adjacent := w.getAdjacentChunks(chunkX, chunkZ)
	for relative, c := range adjacent {
		if c != nil {
			adjacent[relative] = c.Clone()
		}
	}

	w.executor().Populate(chunkX, chunkZ, center, adjacent, w.registrySnapshot(), func(result PopulationResult) {
		if w.closed {
			return
		}
		w.generateChunkCallback(lockID, chunkX, chunkZ, result, temporaryLoader)
	})

	return resolver.GetPromise()
}

// generateChunkCallback is a port of World::generateChunkCallback.
func (w *World) generateChunkCallback(lockID *ChunkLockID, x, z int, result PopulationResult, temporaryLoader *temporaryChunkLoader) {
	dirtyChunks := 0
	for xx := -1; xx <= 1; xx++ {
		for zz := -1; zz <= 1; zz++ {
			w.UnregisterChunkLoader(temporaryLoader, x+xx, z+zz)
			if !w.UnlockChunk(x+xx, z+zz, lockID) {
				dirtyChunks++
			}
		}
	}

	// The worker may have written block states the World didn't know yet.
	for _, blk := range result.NewTemplates {
		w.registerTemplate(blk)
	}

	key := chunkKey(x, z)
	active, ok := w.activeChunkPopulationTasks[key]
	if !ok {
		panic("This should always be set, regardless of whether the task was orphaned or not")
	}
	if !active {
		w.debug(fmt.Sprintf("Discarding orphaned population result for chunk x=%d,z=%d", x, z))
		delete(w.activeChunkPopulationTasks, key)
		return
	}

	chunk := result.Center
	if dirtyChunks == 0 {
		oldChunk := w.loadChunk(x, z)
		w.SetChunk(x, z, chunk)

		for relative, adjacentChunk := range result.Adjacent {
			if relative[0] < -1 || relative[0] > 1 || relative[1] < -1 || relative[1] > 1 {
				panic("Adjacent chunks should be in range -1 ... +1 coordinates")
			}
			w.SetChunk(x+relative[0], z+relative[1], adjacentChunk)
		}

		if (oldChunk == nil || !oldChunk.IsPopulated()) && chunk.IsPopulated() {
			if event.HasHandlers[worldevent.ChunkPopulateEvent]() {
				event.Call(worldevent.NewChunkPopulateEvent(w, x, z, chunk))
			}
			for _, listener := range w.GetChunkListeners(x, z) {
				listener.OnChunkPopulated(x, z, chunk)
			}
		}
	} else {
		w.debug(fmt.Sprintf("Discarding population result for chunk x=%d,z=%d - terrain was modified on the main thread before async population completed", x, z))
	}

	//This needs to be in this specific spot because user code might call back to orderChunkPopulation().
	//If it does, and finds the promise, and doesn't find an active task associated with it, it will schedule
	//another PopulationTask. We don't want that because we're here processing the results.
	delete(w.activeChunkPopulationTasks, key)

	if dirtyChunks == 0 {
		if resolver, ok := w.chunkPopulationRequestMap[key]; ok {
			delete(w.chunkPopulationRequestMap, key)
			resolver.Resolve(chunk)
		} else {
			//Handlers of ChunkPopulateEvent, ChunkLoadEvent, or just ChunkListeners can cause this
			w.debug(fmt.Sprintf("Unable to resolve population promise for chunk x=%d,z=%d - populated chunk was forcibly unloaded while setting modified chunks", x, z))
		}
	} else {
		//request failed, stick it back on the queue
		//we didn't resolve the promise or touch it in any way, so any fake chunk loaders are still valid and
		//don't need to be added a second time.
		w.addChunkHashToPopulationRequestQueue(key)
	}

	w.drainPopulationRequestQueue()
}

// rejectPopulationRequest is the part of World::unloadChunk that fails a pending population
// request for the unloaded chunk.
func (w *World) rejectPopulationRequest(chunkX, chunkZ int) {
	key := chunkKey(chunkX, chunkZ)
	if resolver, ok := w.chunkPopulationRequestMap[key]; ok {
		w.debug(fmt.Sprintf("Rejecting population promise for chunk %d %d", chunkX, chunkZ))
		delete(w.chunkPopulationRequestMap, key)
		resolver.Reject()
		if _, active := w.activeChunkPopulationTasks[key]; active {
			w.debug(fmt.Sprintf("Marking population task for chunk %d %d as orphaned", chunkX, chunkZ))
			w.activeChunkPopulationTasks[key] = false
		}
	}
}

// orderLightPopulation is a port of World::orderLightPopulation: the chunk's light is calculated
// on a worker (synchronously without an AsyncPool).
func (w *World) orderLightPopulation(chunkX, chunkZ int) {
	chunk, ok := w.chunks[chunkKey(chunkX, chunkZ)]
	if !ok {
		return
	}
	if lit, known := chunk.IsLightPopulated(); !known || lit {
		return
	}
	chunk.SetLightPopulated(false, false) // null: in progress

	apply := func(result LightPopulationResult) {
		c, ok := w.GetChunk(chunkX, chunkZ)
		if w.closed || !ok {
			return
		}
		if lit, known := c.IsLightPopulated(); known && lit {
			return
		}
		//TODO: calculated light information might not be valid if the terrain changed during light calculation
		c.SetHeightMapArray(result.HeightMap)
		for y, lightArray := range result.BlockLight {
			c.GetSubChunk(y).SetBlockLightArray(lightArray)
		}
		for y, lightArray := range result.SkyLight {
			c.GetSubChunk(y).SetBlockSkyLightArray(lightArray)
		}
		c.SetLightPopulated(true, true)
	}

	if w.workerPool == nil {
		apply(computeChunkLight(chunk.Clone(), w.registrySnapshot()))
		return
	}
	w.workerPool.SubmitTask(NewLightPopulationTask(chunk.Clone(), w.registrySnapshot(), apply))
}

// debug logs a debug message on the world's logger, if it has one.
func (w *World) debug(message string) {
	if w.logger != nil {
		w.logger.Debug(message)
	}
}
