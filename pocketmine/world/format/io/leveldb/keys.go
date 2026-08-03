// Package leveldb is a port of a slice of pocketmine\world\format\io\leveldb: reading and writing
// chunks to/from a real Bedrock-compatible LevelDB world database, using
// github.com/syndtr/goleveldb (a well-established pure-Go LevelDB implementation - infrastructure,
// the same category as go-raknet/gophertunnel, not game logic) purely as the underlying key/value
// storage engine. The actual chunk/subchunk binary layout is hand-written here from
// ChunkDataKey.php/SubChunkVersion.php/the relevant LevelDB.php methods, matching this port's
// existing hand-written network chunk serializer.
//
// Scoped to the overworld dimension only (no dimension suffix on keys) and to exactly the tags
// this port's World actually needs to round-trip: SUBCHUNK block data and 3D biomes. Tiles,
// entities, scheduled ticks, legacy conversion data and every other ChunkDataKey tag aren't
// written or read - this port has no Tile-in-World/Entity-in-World system yet for the first two,
// and the rest only exist to support migrating old worlds forward, which doesn't apply to a world
// this port itself created.
package leveldb

import "pocketmine-go/pocketmine/binaryutils"

// Tag bytes are a port of ChunkDataKey's constants (the subset this package uses).
const (
	tagHeightmapAnd3DBiomes byte = 0x2b
	tagNewVersion           byte = 0x2c
	tagSubChunk             byte = 0x2f
	tagFinalization         byte = 0x36
)

// chunkVersion is a port of WorldDataVersions::CHUNK (ChunkVersion::v1_21_120 = 42), written under
// the NEW_VERSION tag.
const chunkVersion byte = 42

// subChunkVersion is a port of WorldDataVersions::SUBCHUNK (SubChunkVersion::PALETTED_MULTI = 8) -
// the same envelope shape (version byte, layer count, per-layer bitsPerBlock/words/palette) this
// port's network ChunkSerializer already writes, differing only in how palette values are encoded
// (see subchunk.go).
const subChunkVersion byte = 8

// finalizationDone is a port of LevelDB::FINALISATION_DONE - every chunk this port saves is always
// fully generated and populated by the time it exists in World.chunks at all (see World.go's
// GetOrLoadChunk/ensurePopulated), so FINALISATION_NEEDS_POPULATION is never written.
const finalizationDone byte = 2

// chunkIndex is a port of LevelDB::chunkIndex - the common key prefix for every tag belonging to
// one chunk (no dimension suffix - see this package's doc comment on why).
func chunkIndex(chunkX, chunkZ int32) []byte {
	buf := make([]byte, 0, 8)
	buf = append(buf, binaryutils.WriteLInt(chunkX)...)
	buf = append(buf, binaryutils.WriteLInt(chunkZ)...)
	return buf
}

func taggedKey(chunkX, chunkZ int32, tag byte) []byte {
	return append(chunkIndex(chunkX, chunkZ), tag)
}

func subChunkKey(chunkX, chunkZ int32, y int) []byte {
	return append(taggedKey(chunkX, chunkZ, tagSubChunk), byte(int8(y)))
}
