package player

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/format"
)

// var _ world.ChunkListener = (*Player)(nil) confirms *Player structurally satisfies
// world.ChunkListener's real interface (OnChunkChanged/OnChunkLoaded/OnChunkUnloaded/
// OnChunkPopulated/OnBlockChanged) - the same compile-time check block.Player already gets in
// player.go.
var _ world.ChunkListener = (*Player)(nil)

// GetViewDistance is a port of Player::getViewDistance.
func (p *Player) GetViewDistance() int { return p.viewDistance }

// SetViewDistance is a port of a slice of Player::setViewDistance - minus the server-wide
// allowed-view-distance clamp (no Server type exists in this package) and the cancellable
// PlayerViewDistanceChangeEvent (no event bus wired up here yet).
func (p *Player) SetViewDistance(distance int) {
	if distance == p.viewDistance {
		return
	}
	p.viewDistance = distance
}

// IsUsingChunk is a port of Player::isUsingChunk.
func (p *Player) IsUsingChunk(chunkX, chunkZ int) bool {
	_, ok := p.usedChunks[[2]int{chunkX, chunkZ}]
	return ok
}

// GetUsedChunks is a port of Player::getUsedChunks.
func (p *Player) GetUsedChunks() map[[2]int]UsedChunkStatus { return p.usedChunks }

// GetUsedChunkStatus is a port of Player::getUsedChunkStatus.
func (p *Player) GetUsedChunkStatus(chunkX, chunkZ int) (UsedChunkStatus, bool) {
	s, ok := p.usedChunks[[2]int{chunkX, chunkZ}]
	return s, ok
}

// HasReceivedChunk is a port of Player::hasReceivedChunk.
func (p *Player) HasReceivedChunk(chunkX, chunkZ int) bool {
	s, ok := p.GetUsedChunkStatus(chunkX, chunkZ)
	return ok && s == UsedChunkStatusSent
}

// unloadChunk is a port of Player::unloadChunk - minus NetworkSession::stopUsingChunk (no network
// session type in this package).
func (p *Player) unloadChunk(chunkX, chunkZ int) {
	key := [2]int{chunkX, chunkZ}
	if _, ok := p.usedChunks[key]; ok {
		for _, e := range p.GetWorld().GetChunkEntities(chunkX, chunkZ) {
			if e != world.Entity(p) {
				e.DespawnFrom(p, true)
			}
		}
		delete(p.usedChunks, key)
	}
	p.GetWorld().UnregisterChunkLoader(p, chunkX, chunkZ)
	p.GetWorld().UnregisterChunkListener(p, chunkX, chunkZ)
	delete(p.loadQueue, key)
	p.GetWorld().UnregisterTickingChunk(p, chunkX, chunkZ)
	delete(p.tickingChunks, key)
}

// OrderChunks is a port of a slice of Player::orderChunks - minus the network-session view-area
// sync (no network session type here) and Timings instrumentation. Returns early (a no-op) while
// viewDistance is still -1 (its zero-value default, matching real PHP's own uninitialized state
// before the first SetViewDistance call).
func (p *Player) OrderChunks() {
	if p.viewDistance == -1 {
		return
	}

	newLoadQueue := map[[2]int]bool{}
	var newLoadQueueOrder [][2]int
	newTickingChunks := map[[2]int]bool{}
	unloadChunks := make(map[[2]int]UsedChunkStatus, len(p.usedChunks))
	for k, v := range p.usedChunks {
		unloadChunks[k] = v
	}

	tickingChunkRadius := p.GetWorld().GetChunkTickRadius()

	centerX, centerZ := p.GetPosition().FloorX()>>4, p.GetPosition().FloorZ()>>4
	radius := 0
	for chunk := range SelectChunks(p.viewDistance, centerX, centerZ) {
		if status, ok := p.usedChunks[chunk]; !ok || status == UsedChunkStatusNeeded {
			newLoadQueue[chunk] = true
			newLoadQueueOrder = append(newLoadQueueOrder, chunk)
		}
		if radius < tickingChunkRadius {
			newTickingChunks[chunk] = true
		}
		delete(unloadChunks, chunk)
		radius++
	}

	for chunk := range unloadChunks {
		p.unloadChunk(chunk[0], chunk[1])
	}

	p.loadQueue = newLoadQueue
	p.loadQueueOrder = newLoadQueueOrder
	p.updateTickingChunkRegistrations(p.tickingChunks, newTickingChunks)
	p.tickingChunks = newTickingChunks
}

// updateTickingChunkRegistrations is a port of Player::updateTickingChunkRegistrations.
func (p *Player) updateTickingChunkRegistrations(oldTickingChunks, newTickingChunks map[[2]int]bool) {
	for chunk := range oldTickingChunks {
		if !newTickingChunks[chunk] && !p.loadQueue[chunk] {
			p.GetWorld().UnregisterTickingChunk(p, chunk[0], chunk[1])
		}
	}
	for chunk := range newTickingChunks {
		if !oldTickingChunks[chunk] && !p.loadQueue[chunk] {
			p.GetWorld().RegisterTickingChunk(p, chunk[0], chunk[1])
		}
	}
}

// RequestChunks is a port of Player::requestChunks: at most chunksPerTick chunks from the load
// queue (nearest first) are generated, registered for this player and returned as ready to send;
// the caller (NetworkSession, PHP's startUsingChunk) sends them and calls MarkChunkSent.
// Generation is synchronous in this port (see World.ensurePopulated), so there are never active
// generation requests counting against the limit.
func (p *Player) RequestChunks() [][2]int {
	var readyToSend [][2]int

	count := 0
	limit := p.chunksPerTick
	remaining := p.loadQueueOrder[:0]
	for _, chunk := range p.loadQueueOrder {
		if !p.loadQueue[chunk] {
			continue // unloaded (unloadChunk) since it was queued
		}
		if count >= limit {
			remaining = append(remaining, chunk)
			continue
		}
		count++

		chunkX, chunkZ := chunk[0], chunk[1]

		p.usedChunks[chunk] = UsedChunkStatusRequestedGeneration
		delete(p.loadQueue, chunk)

		p.GetWorld().RegisterChunkLoader(p, chunkX, chunkZ)
		p.GetWorld().RegisterChunkListener(p, chunkX, chunkZ)
		if p.tickingChunks[chunk] {
			p.GetWorld().RegisterTickingChunk(p, chunkX, chunkZ)
		}

		p.GetWorld().GetOrLoadChunk(chunkX, chunkZ)

		p.usedChunks[chunk] = UsedChunkStatusRequestedSending
		readyToSend = append(readyToSend, chunk)
	}
	p.loadQueueOrder = remaining

	return readyToSend
}

// MarkChunkSent is a port of the network-completion callback inside Player::requestChunks
// (`$this->usedChunks[$index] = UsedChunkStatus::SENT;`) - called by the caller once it has
// actually written the chunk data to the client.
func (p *Player) MarkChunkSent(chunkX, chunkZ int) {
	key := [2]int{chunkX, chunkZ}
	if p.usedChunks[key] == UsedChunkStatusRequestedSending {
		p.usedChunks[key] = UsedChunkStatusSent
		if p.spawned {
			p.spawnEntitiesOnChunk(chunkX, chunkZ)
		}
	}
}

// spawnEntitiesOnAllChunks is a port of Player::spawnEntitiesOnAllChunks.
func (p *Player) spawnEntitiesOnAllChunks() {
	for key, status := range p.usedChunks {
		if status == UsedChunkStatusSent {
			p.spawnEntitiesOnChunk(key[0], key[1])
		}
	}
}

// spawnEntitiesOnChunk is a port of Player::spawnEntitiesOnChunk: every entity in a chunk the
// player has just received is spawned to it.
func (p *Player) spawnEntitiesOnChunk(chunkX, chunkZ int) {
	for _, e := range p.GetWorld().GetChunkEntities(chunkX, chunkZ) {
		if e != world.Entity(p) && !e.IsFlaggedForDespawn() {
			e.SpawnTo(p)
		}
	}
}

// OnChunkChanged is a port of Player::onChunkChanged: if a chunk this player has already fully
// sent gets replaced outright, it needs to be sent again.
func (p *Player) OnChunkChanged(chunkX, chunkZ int, chunk *format.Chunk) {
	key := [2]int{chunkX, chunkZ}
	if p.usedChunks[key] == UsedChunkStatusSent {
		p.usedChunks[key] = UsedChunkStatusNeeded
	}
}

// OnChunkLoaded is a port of Player::onChunkLoaded - real PHP never overrides
// ChunkListenerNoOpTrait's default (a no-op) for this one.
func (p *Player) OnChunkLoaded(chunkX, chunkZ int, chunk *format.Chunk) {}

// OnChunkUnloaded is a port of Player::onChunkUnloaded: a chunk being forcibly unloaded out from
// under this player needs the same teardown as if the player had stopped using it themselves.
func (p *Player) OnChunkUnloaded(chunkX, chunkZ int, chunk *format.Chunk) {
	if p.IsUsingChunk(chunkX, chunkZ) {
		p.unloadChunk(chunkX, chunkZ)
	}
}

// OnChunkPopulated is a port of Player::onChunkPopulated - real PHP never overrides
// ChunkListenerNoOpTrait's default (a no-op) for this one either.
func (p *Player) OnChunkPopulated(chunkX, chunkZ int, chunk *format.Chunk) {}

// OnBlockChanged is a port of Player::onBlockChanged - minus the sleep-interruption check (no
// sleep system exists in this port yet, matching Player's own doc comment on what's left out).
func (p *Player) OnBlockChanged(pos math.Vector3) {}
