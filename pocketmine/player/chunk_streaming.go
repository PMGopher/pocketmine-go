package player

import (
	"fmt"

	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/timings"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/format"
)

var _ world.ChunkListener = (*Player)(nil)

// noChunkOrderRun is PHP_INT_MAX as nextChunkOrderRun: no reordering scheduled.
const noChunkOrderRun = int(^uint(0) >> 1)

// unloadChunk is a port of Player::unloadChunk: w nil means the player's current world.
func (p *Player) unloadChunk(chunkX, chunkZ int, w *world.World) {
	if w == nil {
		w = p.GetWorld()
	}
	key := [2]int{chunkX, chunkZ}
	if _, ok := p.usedChunks[key]; ok {
		for _, e := range w.GetChunkEntities(chunkX, chunkZ) {
			if e != world.Entity(p) {
				e.DespawnFrom(p, true)
			}
		}
		if p.networkSession != nil {
			p.networkSession.StopUsingChunk(chunkX, chunkZ)
		}
		delete(p.usedChunks, key)
		delete(p.activeChunkGenerationRequests, key)
	}
	w.UnregisterChunkLoader(p, chunkX, chunkZ)
	w.UnregisterChunkListener(p, chunkX, chunkZ)
	delete(p.loadQueue, key)
	w.UnregisterTickingChunk(p, chunkX, chunkZ)
	delete(p.tickingChunks, key)
}

// spawnEntitiesOnAllChunks is a port of Player::spawnEntitiesOnAllChunks.
func (p *Player) spawnEntitiesOnAllChunks() {
	for key, status := range p.usedChunks {
		if status == UsedChunkStatusSent {
			p.spawnEntitiesOnChunk(key[0], key[1])
		}
	}
}

// spawnEntitiesOnChunk is a port of Player::spawnEntitiesOnChunk.
func (p *Player) spawnEntitiesOnChunk(chunkX, chunkZ int) {
	for _, e := range p.GetWorld().GetChunkEntities(chunkX, chunkZ) {
		if e != world.Entity(p) && !e.IsFlaggedForDespawn() {
			e.SpawnTo(p)
		}
	}
}

// requestChunks is a port of Player::requestChunks: requests chunks from the world to be sent, up
// to a set limit every tick. This operates on the results of the most recent chunk order. Chunks
// that aren't populated yet are populated asynchronously (World.RequestChunkPopulation); they're
// sent once that completes.
func (p *Player) requestChunks() {
	if !p.IsConnected() {
		return
	}

	timings.Init()
	timings.PlayerChunkSend.StartTiming()
	defer timings.PlayerChunkSend.StopTiming()

	count := 0
	w := p.GetWorld()

	limit := p.chunksPerTick - len(p.activeChunkGenerationRequests)
	remaining := p.loadQueueOrder[:0:0]
	for _, index := range p.loadQueueOrder {
		if !p.loadQueue[index] {
			continue // unloaded (unloadChunk) since it was queued
		}
		if count >= limit {
			remaining = append(remaining, index)
			continue
		}

		chunkX, chunkZ := index[0], index[1]

		count++

		p.usedChunks[index] = UsedChunkStatusRequestedGeneration
		p.activeChunkGenerationRequests[index] = true
		delete(p.loadQueue, index)
		w.RegisterChunkLoader(p, chunkX, chunkZ)
		w.RegisterChunkListener(p, chunkX, chunkZ)
		if p.tickingChunks[index] {
			w.RegisterTickingChunk(p, chunkX, chunkZ)
		}

		w.RequestChunkPopulation(chunkX, chunkZ, p).OnCompletion(
			func(*format.Chunk) {
				status, ok := p.usedChunks[index]
				if !p.IsConnected() || !ok || w != p.GetWorld() {
					return
				}
				if status != UsedChunkStatusRequestedGeneration {
					//We may have previously requested this, decided we didn't want it, and then decided we did want
					//it again, all before the generation request got executed. In that case, the promise would have
					//multiple callbacks for this player. In that case, only the first one matters.
					return
				}
				delete(p.activeChunkGenerationRequests, index)
				p.usedChunks[index] = UsedChunkStatusRequestedSending

				p.GetNetworkSession().StartUsingChunk(chunkX, chunkZ, func() {
					p.usedChunks[index] = UsedChunkStatusSent
					if p.spawnChunkLoadCount == -1 {
						p.spawnEntitiesOnChunk(chunkX, chunkZ)
					} else {
						loaded := p.spawnChunkLoadCount
						p.spawnChunkLoadCount++
						if loaded == p.spawnThreshold {
							p.spawnChunkLoadCount = -1

							p.spawnEntitiesOnAllChunks()

							p.GetNetworkSession().NotifyTerrainReady()
						}
					}
					event.Call(playerevent.NewPlayerPostChunkSendEvent(p, chunkX, chunkZ))
				})
			},
			func() {
				//NOOP: we'll re-request this if it fails anyway
			},
		)
	}
	p.loadQueueOrder = remaining
}

// recheckBroadcastPermissions is a port of Player::recheckBroadcastPermissions.
func (p *Player) recheckBroadcastPermissions() {
	for _, entry := range []struct{ permission, channel string }{
		{"pocketmine.broadcast.admin", BroadcastChannelAdministrative},
		{"pocketmine.broadcast.user", BroadcastChannelUsers},
	} {
		if p.HasPermission(entry.permission) {
			p.server.SubscribeToBroadcastChannel(entry.channel, p)
		} else {
			p.server.UnsubscribeFromBroadcastChannel(entry.channel, p)
		}
	}
}

// Broadcast channels, Server::BROADCAST_CHANNEL_*.
const (
	BroadcastChannelAdministrative = "pocketmine.broadcast.admin"
	BroadcastChannelUsers          = "pocketmine.broadcast.user"
)

// updateTickingChunkRegistrations is a port of Player::updateTickingChunkRegistrations.
func (p *Player) updateTickingChunkRegistrations(oldTickingChunks, newTickingChunks map[[2]int]bool) {
	w := p.GetWorld()
	for chunk := range oldTickingChunks {
		if !newTickingChunks[chunk] && !p.loadQueue[chunk] {
			//we are (probably) still using this chunk, but it's no longer within ticking range
			w.UnregisterTickingChunk(p, chunk[0], chunk[1])
		}
	}
	for chunk := range newTickingChunks {
		if !oldTickingChunks[chunk] && !p.loadQueue[chunk] {
			//we were already using this chunk, but it is now within ticking range
			w.RegisterTickingChunk(p, chunk[0], chunk[1])
		}
	}
}

// orderChunks is a port of Player::orderChunks: calculates which new chunks this player needs to
// use, and which currently-used chunks it needs to stop using.
func (p *Player) orderChunks() {
	if !p.IsConnected() || p.viewDistance == -1 {
		return
	}

	timings.Init()
	timings.PlayerChunkOrder.StartTiming()
	defer timings.PlayerChunkOrder.StopTiming()

	newOrder := map[[2]int]bool{}
	var newOrderList [][2]int
	tickingChunks := map[[2]int]bool{}
	unloadChunks := make(map[[2]int]UsedChunkStatus, len(p.usedChunks))
	for k, v := range p.usedChunks {
		unloadChunks[k] = v
	}

	w := p.GetWorld()
	tickingChunkRadius := w.GetChunkTickRadius()

	loc := p.GetLocation()
	radius := 0
	for hash := range SelectChunks(p.server.GetAllowedViewDistance(p.viewDistance), loc.FloorX()>>4, loc.FloorZ()>>4) {
		if status, ok := p.usedChunks[hash]; !ok || status == UsedChunkStatusNeeded {
			newOrder[hash] = true
			newOrderList = append(newOrderList, hash)
		}
		if radius < tickingChunkRadius {
			tickingChunks[hash] = true
		}
		delete(unloadChunks, hash)
		radius++
	}

	for index := range unloadChunks {
		p.unloadChunk(index[0], index[1], nil)
	}

	p.loadQueue = newOrder
	p.loadQueueOrder = newOrderList

	p.updateTickingChunkRegistrations(p.tickingChunks, tickingChunks)
	p.tickingChunks = tickingChunks

	if len(p.loadQueue) > 0 || len(unloadChunks) > 0 {
		p.GetNetworkSession().SyncViewAreaCenterPoint(loc.Vector3, p.viewDistance)
	}
}

// IsUsingChunk returns whether the player is using the chunk with the given coordinates,
// irrespective of whether the chunk has been sent yet.
func (p *Player) IsUsingChunk(chunkX, chunkZ int) bool {
	_, ok := p.usedChunks[[2]int{chunkX, chunkZ}]
	return ok
}

// GetUsedChunks is a port of Player::getUsedChunks.
func (p *Player) GetUsedChunks() map[[2]int]UsedChunkStatus { return p.usedChunks }

// GetUsedChunkStatus returns a usage status of the given chunk, and false if the player is not
// using the given chunk.
func (p *Player) GetUsedChunkStatus(chunkX, chunkZ int) (UsedChunkStatus, bool) {
	s, ok := p.usedChunks[[2]int{chunkX, chunkZ}]
	return s, ok
}

// HasReceivedChunk returns whether the target chunk has been sent to this player.
func (p *Player) HasReceivedChunk(chunkX, chunkZ int) bool {
	s, ok := p.GetUsedChunkStatus(chunkX, chunkZ)
	return ok && s == UsedChunkStatusSent
}

// DoChunkRequests is a port of Player::doChunkRequests: ticks the chunk-requesting mechanism.
func (p *Player) DoChunkRequests() {
	if p.nextChunkOrderRun != noChunkOrderRun {
		run := p.nextChunkOrderRun
		p.nextChunkOrderRun--
		if run <= 0 {
			p.nextChunkOrderRun = noChunkOrderRun
			p.orderChunks()
		}
	}

	if len(p.loadQueue) > 0 {
		p.requestChunks()
	}
}

// OnChunkChanged is a port of Player::onChunkChanged.
func (p *Player) OnChunkChanged(chunkX, chunkZ int, chunk *format.Chunk) {
	key := [2]int{chunkX, chunkZ}
	if status, ok := p.usedChunks[key]; ok && status == UsedChunkStatusSent {
		p.usedChunks[key] = UsedChunkStatusNeeded
		p.nextChunkOrderRun = 0
	}
}

// OnChunkLoaded is ChunkListenerNoOpTrait's no-op.
func (p *Player) OnChunkLoaded(chunkX, chunkZ int, chunk *format.Chunk) {}

// OnChunkUnloaded is a port of Player::onChunkUnloaded.
func (p *Player) OnChunkUnloaded(chunkX, chunkZ int, chunk *format.Chunk) {
	if p.IsUsingChunk(chunkX, chunkZ) {
		p.logger.Debug(fmt.Sprintf("Detected forced unload of chunk %d %d", chunkX, chunkZ))
		p.unloadChunk(chunkX, chunkZ, nil)
	}
}

// OnChunkPopulated is ChunkListenerNoOpTrait's no-op.
func (p *Player) OnChunkPopulated(chunkX, chunkZ int, chunk *format.Chunk) {}

// OnBlockChanged is a port of Player::onBlockChanged: sleeping stops if the bed is gone.
func (p *Player) OnBlockChanged(pos math.Vector3) {
	if p.sleeping != nil && pos.Floor() == *p.sleeping {
		if _, isBed := p.GetWorld().GetBlock(pos).(bedBlock); !isBed {
			p.logger.Debug("Bed was changed or deleted, aborting sleep")
			p.StopSleep()
		}
	}
}
