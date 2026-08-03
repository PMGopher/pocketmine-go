package leveldb

import (
	"errors"
	"fmt"

	"github.com/syndtr/goleveldb/leveldb"

	"pocketmine-go/pocketmine/world/format"
)

// biomeHeightmapPlaceholderSize is a port of the 512 zero bytes LevelDB::writeChunk itself writes
// ahead of the 3D biome palettes ("fake heightmap" - PHP's own inline comment). This port has no
// heightmap/light engine yet (see World's light-query methods' doc comments), so - like real
// PocketMine-MP's own save path - it never writes a real one either; a real light engine
// recomputes it from scratch on load in any implementation, this port's included once one exists.
const biomeHeightmapPlaceholderSize = 512

// SaveChunk is a port of the relevant slice of LevelDB::writeChunk: NEW_VERSION, one SUBCHUNK
// entry per non-empty subchunk (deleting the key for empty ones, matching real behaviour), the 3D
// biome palettes, and FINALIZATION. Tiles/entities/scheduled ticks aren't written - see this
// package's doc comment on why.
func SaveChunk(db *leveldb.DB, chunkX, chunkZ int32, chunk *format.Chunk, lookup StateLookup) error {
	batch := new(leveldb.Batch)

	batch.Put(taggedKey(chunkX, chunkZ, tagNewVersion), []byte{chunkVersion})

	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		key := subChunkKey(chunkX, chunkZ, y)
		sc := chunk.GetSubChunk(y)
		if sc.IsEmptyFast() {
			batch.Delete(key)
			continue
		}
		data, err := SerializeSubChunk(sc, lookup)
		if err != nil {
			return fmt.Errorf("leveldb: saving chunk (%d,%d) subchunk %d: %w", chunkX, chunkZ, y, err)
		}
		batch.Put(key, data)
	}

	biomeBuf := make([]byte, biomeHeightmapPlaceholderSize)
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		biomeBuf = append(biomeBuf, serializeBiomePalette(chunk.GetSubChunk(y).GetBiomeArray())...)
	}
	batch.Put(taggedKey(chunkX, chunkZ, tagHeightmapAnd3DBiomes), biomeBuf)

	batch.Put(taggedKey(chunkX, chunkZ, tagFinalization), []byte{finalizationDone})

	if err := db.Write(batch, nil); err != nil {
		return fmt.Errorf("leveldb: saving chunk (%d,%d): %w", chunkX, chunkZ, err)
	}
	return nil
}

// LoadChunk is the reverse of SaveChunk. Returns ok=false (no error) if the chunk simply isn't
// present in the database yet - matching World.GetOrLoadChunk's "load, or else generate" contract.
func LoadChunk(db *leveldb.DB, chunkX, chunkZ int32, emptyBlockID, defaultBiomeID int32, resolve StateResolver) (*format.Chunk, bool, error) {
	versionKey := taggedKey(chunkX, chunkZ, tagNewVersion)
	if _, err := db.Get(versionKey, nil); err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("leveldb: loading chunk (%d,%d): %w", chunkX, chunkZ, err)
	}

	biomeData, err := db.Get(taggedKey(chunkX, chunkZ, tagHeightmapAnd3DBiomes), nil)
	if err != nil {
		return nil, false, fmt.Errorf("leveldb: loading chunk (%d,%d) biomes: %w", chunkX, chunkZ, err)
	}
	if len(biomeData) < biomeHeightmapPlaceholderSize {
		return nil, false, fmt.Errorf("leveldb: chunk (%d,%d) biome data too short (%d bytes)", chunkX, chunkZ, len(biomeData))
	}

	offset := biomeHeightmapPlaceholderSize
	biomeArrays := make(map[int]*format.PalettedBlockArray, format.MaxSubChunks)
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		arr, newOffset, err := deserializeBiomePalette(biomeData, offset)
		if err != nil {
			return nil, false, fmt.Errorf("leveldb: loading chunk (%d,%d) biome subchunk %d: %w", chunkX, chunkZ, y, err)
		}
		biomeArrays[y] = arr
		offset = newOffset
	}

	subChunks := map[int]*format.SubChunk{}
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		data, err := db.Get(subChunkKey(chunkX, chunkZ, y), nil)
		if err != nil {
			if errors.Is(err, leveldb.ErrNotFound) {
				subChunks[y] = format.NewSubChunk(emptyBlockID, nil, biomeArrays[y])
				continue
			}
			return nil, false, fmt.Errorf("leveldb: loading chunk (%d,%d) subchunk %d: %w", chunkX, chunkZ, y, err)
		}
		layers, err := DeserializeSubChunkLayers(data, emptyBlockID, resolve)
		if err != nil {
			return nil, false, fmt.Errorf("leveldb: loading chunk (%d,%d) subchunk %d: %w", chunkX, chunkZ, y, err)
		}
		subChunks[y] = format.NewSubChunk(emptyBlockID, layers, biomeArrays[y])
	}

	chunk := format.NewChunk(subChunks, true, emptyBlockID, defaultBiomeID)
	return chunk, true, nil
}
