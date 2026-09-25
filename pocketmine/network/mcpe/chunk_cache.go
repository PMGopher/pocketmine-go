package mcpe

import (
	"sync"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/serializer"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/format"
)

// ChunkCache is a port of pocketmine\network\mcpe\cache\ChunkCache: caches the serialized network
// data of chunks for a world, so that sending a chunk to many players only serializes it once.
// It listens to the chunks it caches and drops them when they change.
//
// In the sub-chunk request mode this port sends chunks in (see sub_chunk_request.go), LevelChunk
// only carries the chunk's biomes, so those are what's cached; the packet itself is assembled per
// session, since clients using the blob cache get the biomes as a hash instead.
type ChunkCache struct {
	world  *world.World
	mu     sync.Mutex
	caches map[[2]int]*cachedChunk
	hits   int
	misses int
}

type cachedChunk struct {
	subChunkCount int
	biomes        []byte // SerializeBiomes: the biome blob
	payload       []byte // SerializeBiomesPayload: biomes + border block count
}

var (
	chunkCachesMu sync.Mutex
	chunkCaches   = map[*world.World]*ChunkCache{}
)

// GetChunkCache is ChunkCache::getInstance.
func GetChunkCache(w *world.World) *ChunkCache {
	chunkCachesMu.Lock()
	defer chunkCachesMu.Unlock()
	c, ok := chunkCaches[w]
	if !ok {
		c = &ChunkCache{world: w, caches: map[[2]int]*cachedChunk{}}
		chunkCaches[w] = c
	}
	return c
}

// PruneChunkCaches is ChunkCache::pruneCaches: drops every cached chunk (low memory).
func PruneChunkCaches() {
	chunkCachesMu.Lock()
	defer chunkCachesMu.Unlock()
	for w, c := range chunkCaches {
		if w.IsClosed() {
			delete(chunkCaches, w)
			continue
		}
		c.mu.Lock()
		c.caches = map[[2]int]*cachedChunk{}
		c.mu.Unlock()
	}
}

// request is ChunkCache::request: the cached data for the chunk, serializing it on a miss.
func (c *ChunkCache) request(chunkX, chunkZ int, chunk *format.Chunk) *cachedChunk {
	key := [2]int{chunkX, chunkZ}
	c.mu.Lock()
	if cached, ok := c.caches[key]; ok {
		c.hits++
		c.mu.Unlock()
		return cached
	}
	c.misses++
	c.mu.Unlock()

	c.world.RegisterChunkListener(c, chunkX, chunkZ)
	cached := &cachedChunk{
		subChunkCount: serializer.GetSubChunkCount(chunk),
		biomes:        serializer.SerializeBiomes(chunk),
		payload:       serializer.SerializeBiomesPayload(chunk),
	}
	c.mu.Lock()
	c.caches[key] = cached
	c.mu.Unlock()
	return cached
}

func (c *ChunkCache) destroy(chunkX, chunkZ int) {
	c.mu.Lock()
	delete(c.caches, [2]int{chunkX, chunkZ})
	c.mu.Unlock()
}

// OnChunkChanged is ChunkCache::onChunkChanged.
func (c *ChunkCache) OnChunkChanged(chunkX, chunkZ int, chunk *format.Chunk) {
	c.destroy(chunkX, chunkZ)
}

// OnBlockChanged is ChunkCache::onBlockChanged.
func (c *ChunkCache) OnBlockChanged(pos math.Vector3) { c.destroy(pos.FloorX()>>4, pos.FloorZ()>>4) }

// OnChunkUnloaded is ChunkCache::onChunkUnloaded.
func (c *ChunkCache) OnChunkUnloaded(chunkX, chunkZ int, chunk *format.Chunk) {
	c.destroy(chunkX, chunkZ)
	c.world.UnregisterChunkListener(c, chunkX, chunkZ)
}

func (c *ChunkCache) OnChunkLoaded(chunkX, chunkZ int, chunk *format.Chunk)    {}
func (c *ChunkCache) OnChunkPopulated(chunkX, chunkZ int, chunk *format.Chunk) {}

// CalculateCacheSize is ChunkCache::calculateCacheSize.
func (c *ChunkCache) CalculateCacheSize() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	size := 0
	for _, cached := range c.caches {
		size += len(cached.biomes) + len(cached.payload)
	}
	return size
}

// GetHitPercentage is ChunkCache::getHitPercentage.
func (c *ChunkCache) GetHitPercentage() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	total := c.hits + c.misses
	if total == 0 {
		return 0
	}
	return float64(c.hits) / float64(total)
}

var _ world.ChunkListener = (*ChunkCache)(nil)
