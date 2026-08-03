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

// unloadChunk is a port of Player::unloadChunk - minus despawning this chunk's entities to the
// player (no despawnFrom/viewer-tracking mechanism exists on entities yet) and
// NetworkSession::stopUsingChunk (no network session type in this package).
func (p *Player) unloadChunk(chunkX, chunkZ int) {
	key := [2]int{chunkX, chunkZ}
	if _, ok := p.usedChunks[key]; ok {
		delete(p.usedChunks, key)
	}
	p.world.UnregisterChunkLoader(p, chunkX, chunkZ)
	p.world.UnregisterChunkListener(p, chunkX, chunkZ)
	delete(p.loadQueue, key)
	p.world.UnregisterTickingChunk(p, chunkX, chunkZ)
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
	newTickingChunks := map[[2]int]bool{}
	unloadChunks := make(map[[2]int]UsedChunkStatus, len(p.usedChunks))
	for k, v := range p.usedChunks {
		unloadChunks[k] = v
	}

	tickingChunkRadius := p.world.GetChunkTickRadius()

	centerX, centerZ := p.GetPosition().FloorX()>>4, p.GetPosition().FloorZ()>>4
	radius := 0
	for chunk := range SelectChunks(p.viewDistance, centerX, centerZ) {
		if status, ok := p.usedChunks[chunk]; !ok || status == UsedChunkStatusNeeded {
			newLoadQueue[chunk] = true
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
	p.updateTickingChunkRegistrations(p.tickingChunks, newTickingChunks)
	p.tickingChunks = newTickingChunks
}

// updateTickingChunkRegistrations is a port of Player::updateTickingChunkRegistrations.
func (p *Player) updateTickingChunkRegistrations(oldTickingChunks, newTickingChunks map[[2]int]bool) {
	for chunk := range oldTickingChunks {
		if !newTickingChunks[chunk] && !p.loadQueue[chunk] {
			p.world.UnregisterTickingChunk(p, chunk[0], chunk[1])
		}
	}
	for chunk := range newTickingChunks {
		if !oldTickingChunks[chunk] && !p.loadQueue[chunk] {
			p.world.RegisterTickingChunk(p, chunk[0], chunk[1])
		}
	}
}

// RequestChunks is a port of a slice of Player::requestChunks: generates (synchronously - this
// port's whole generation pipeline already runs inline, see World.ensurePopulated's own doc
// comment on why there's no async phase to throttle per tick the way real PHP's own
// $chunksPerTick limit does) every chunk still queued, registers this player as their loader/
// listener(+ticker where applicable), and returns the chunk coordinates that are now ready to
// actually be sent over the network - the caller (cmd/pocketmine-go) is responsible for that part,
// then calling MarkChunkSent once it has.
func (p *Player) RequestChunks() [][2]int {
	var readyToSend [][2]int

	for chunk := range p.loadQueue {
		chunkX, chunkZ := chunk[0], chunk[1]

		p.usedChunks[chunk] = UsedChunkStatusRequestedGeneration
		delete(p.loadQueue, chunk)

		p.world.RegisterChunkLoader(p, chunkX, chunkZ)
		p.world.RegisterChunkListener(p, chunkX, chunkZ)
		if p.tickingChunks[chunk] {
			p.world.RegisterTickingChunk(p, chunkX, chunkZ)
		}

		p.world.GetOrLoadChunk(chunkX, chunkZ)

		p.usedChunks[chunk] = UsedChunkStatusRequestedSending
		readyToSend = append(readyToSend, chunk)
	}

	return readyToSend
}

// MarkChunkSent is a port of the network-completion callback inside Player::requestChunks
// (`$this->usedChunks[$index] = UsedChunkStatus::SENT;`) - called by the caller once it has
// actually written the chunk data to the client.
func (p *Player) MarkChunkSent(chunkX, chunkZ int) {
	key := [2]int{chunkX, chunkZ}
	if p.usedChunks[key] == UsedChunkStatusRequestedSending {
		p.usedChunks[key] = UsedChunkStatusSent
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
