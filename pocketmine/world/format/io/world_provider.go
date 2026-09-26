package io

import "pocketmine-go/pocketmine/log"

// ChunkCoords are a chunk's X and Z coordinates, as yielded by WorldProvider.GetAllChunks.
type ChunkCoords struct{ X, Z int }

// WorldProvider is a port of pocketmine\world\format\io\WorldProvider.
type WorldProvider interface {
	GetWorldMinY() int
	GetWorldMaxY() int
	GetPath() string
	// LoadChunk returns nil (and no error) if the chunk doesn't exist. A corrupted chunk is a
	// *exception.CorruptedChunkError.
	LoadChunk(chunkX, chunkZ int) (*LoadedChunkData, error)
	// DoGarbageCollection performs garbage collection in the world provider, such as cleaning up
	// regions in Region-based worlds.
	DoGarbageCollection()
	// GetWorldData returns information about the world.
	GetWorldData() WorldData
	Close() error
	// GetAllChunks calls yield for every chunk in the world (PHP returns a generator); returning
	// false from yield stops. With skipCorrupted, corrupted chunks are logged to logger (if not
	// nil) and skipped; otherwise the first one is returned as the error.
	GetAllChunks(skipCorrupted bool, logger log.Logger, yield func(coords ChunkCoords, chunk *LoadedChunkData) bool) error
	// CalculateChunkCount returns the number of chunks in the provider. Used for world conversion
	// time estimations.
	CalculateChunkCount() (int, error)
}

// WritableWorldProvider is a port of pocketmine\world\format\io\WritableWorldProvider.
type WritableWorldProvider interface {
	WorldProvider
	// SaveChunk saves a chunk (the terrain parts selected by dirtyFlags, a combination of
	// format.DirtyFlag*) to disk.
	SaveChunk(chunkX, chunkZ int, chunkData *ChunkData, dirtyFlags int) error
}
