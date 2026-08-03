package world

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/format"
)

// ChunkListener is a port of pocketmine\world\ChunkListener: register one (see
// RegisterChunkListener) to receive callbacks for specific chunks - block changes, the chunk being
// loaded/unloaded/populated, or replaced outright.
//
// Real PHP's own doc comment warning applies equally here: a listener is never automatically
// unregistered when a chunk unloads - callers must UnregisterChunkListener (or
// UnregisterChunkListenerFromAll) themselves once they're done with it, or it leaks.
//
// Concrete implementations should be pointer types - ChunkListener values are used as map keys
// (see World's own chunkListeners field), which needs identity comparison (matching PHP's
// spl_object_id-keyed inner arrays) rather than by-value equality.
type ChunkListener interface {
	// OnChunkChanged is a port of ChunkListener::onChunkChanged - called when a chunk is replaced
	// by a new one (see World's own doc comment on why this port's single-provider, no-async
	// pipeline never actually replaces an already-loaded chunk wholesale - this fires in principle,
	// just never in practice yet).
	OnChunkChanged(chunkX, chunkZ int, chunk *format.Chunk)
	// OnChunkLoaded is a port of ChunkListener::onChunkLoaded.
	OnChunkLoaded(chunkX, chunkZ int, chunk *format.Chunk)
	// OnChunkUnloaded is a port of ChunkListener::onChunkUnloaded.
	OnChunkUnloaded(chunkX, chunkZ int, chunk *format.Chunk)
	// OnChunkPopulated is a port of ChunkListener::onChunkPopulated.
	OnChunkPopulated(chunkX, chunkZ int, chunk *format.Chunk)
	// OnBlockChanged is a port of ChunkListener::onBlockChanged.
	OnBlockChanged(pos math.Vector3)
}

// RegisterChunkListener is a port of World::registerChunkListener. Not ported: the separate
// playerChunkListeners side-table real PHP also maintains here (`if($listener instanceof Player)`)
// - a pure lookup-performance optimisation for player-only iteration elsewhere, and this port has
// no Player type in the world package to distinguish in the first place.
func (w *World) RegisterChunkListener(listener ChunkListener, chunkX, chunkZ int) {
	key := chunkKey(chunkX, chunkZ)
	if w.chunkListeners[key] == nil {
		w.chunkListeners[key] = map[ChunkListener]bool{}
	}
	w.chunkListeners[key][listener] = true
}

// UnregisterChunkListener is a port of World::unregisterChunkListener.
func (w *World) UnregisterChunkListener(listener ChunkListener, chunkX, chunkZ int) {
	key := chunkKey(chunkX, chunkZ)
	listeners, ok := w.chunkListeners[key]
	if !ok {
		return
	}
	delete(listeners, listener)
	if len(listeners) == 0 {
		delete(w.chunkListeners, key)
	}
}

// UnregisterChunkListenerFromAll is a port of World::unregisterChunkListenerFromAll.
func (w *World) UnregisterChunkListenerFromAll(listener ChunkListener) {
	for key, listeners := range w.chunkListeners {
		if listeners[listener] {
			delete(listeners, listener)
			if len(listeners) == 0 {
				delete(w.chunkListeners, key)
			}
		}
	}
}

// GetChunkListeners is a port of World::getChunkListeners.
func (w *World) GetChunkListeners(chunkX, chunkZ int) []ChunkListener {
	listeners := w.chunkListeners[chunkKey(chunkX, chunkZ)]
	if len(listeners) == 0 {
		return nil
	}
	result := make([]ChunkListener, 0, len(listeners))
	for l := range listeners {
		result = append(result, l)
	}
	return result
}
