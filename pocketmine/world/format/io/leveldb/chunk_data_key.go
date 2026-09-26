// Package leveldb is a port of pocketmine\world\format\io\leveldb: the LevelDB world provider,
// which reads and writes Minecraft: Bedrock Edition worlds (level.dat + db/). The key/value store
// is github.com/df-mc/goleveldb, a goleveldb fork that reads and writes the zlib (raw deflate)
// compressed tables Bedrock uses (PHP uses Mojang's leveldb fork through ext-leveldb).
package leveldb

import "pocketmine-go/pocketmine"

// ChunkDataKey is a port of pocketmine\world\format\io\leveldb\ChunkDataKey: the tag byte after a
// chunk's index in each of its keys.
const (
	ChunkDataKeyHeightmapAnd3DBiomes       = "\x2b"
	ChunkDataKeyNewVersion                 = "\x2c" // since 1.16.100?
	ChunkDataKeyHeightmapAnd2DBiomes       = "\x2d" // obsolete since 1.18
	ChunkDataKeyHeightmapAnd2DBiomeColors  = "\x2e" // obsolete since 1.0
	ChunkDataKeySubChunk                   = "\x2f"
	ChunkDataKeyLegacyTerrain              = "\x30" // obsolete since 1.0
	ChunkDataKeyBlockEntities              = "\x31"
	ChunkDataKeyEntities                   = "\x32"
	ChunkDataKeyPendingScheduledTicks      = "\x33"
	ChunkDataKeyLegacyBlockExtraData       = "\x34" // obsolete since 1.2.13
	ChunkDataKeyBiomeStates                = "\x35" //TODO: is this still applicable to 1.18.0?
	ChunkDataKeyFinalization               = "\x36"
	ChunkDataKeyConverterTag               = "\x37" // ???
	ChunkDataKeyBorderBlocks               = "\x38"
	ChunkDataKeyHardcodedSpawners          = "\x39"
	ChunkDataKeyPendingRandomTicks         = "\x3a"
	ChunkDataKeyXxhashChecksums            = "\x3b" // obsolete since 1.18
	ChunkDataKeyGenerationSeed             = "\x3c"
	ChunkDataKeyGeneratedBeforeCncBlending = "\x3d"

	ChunkDataKeyOldVersion = "\x76"

	ChunkDataKeyPmDataVersion = pocketmine.TagWorldDataVersion
)
