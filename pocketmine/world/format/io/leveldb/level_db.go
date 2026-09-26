package leveldb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/df-mc/goleveldb/leveldb"
	"github.com/df-mc/goleveldb/leveldb/opt"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/binaryutils"
	blockconvert "pocketmine-go/pocketmine/data/bedrock/block/convert"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/format/io"
	worlddata "pocketmine-go/pocketmine/world/format/io/data"
	"pocketmine-go/pocketmine/world/format/io/exception"
)

// LevelDB::FINALISATION_*.
const (
	finalisationNeedsInstaticking = 0
	finalisationNeedsPopulation   = 1
	finalisationDone              = 2
)

// entryFlatWorldLayers is LevelDB::ENTRY_FLAT_WORLD_LAYERS.
const entryFlatWorldLayers = "game_flatworldlayers"

// LevelDB::CURRENT_LEVEL_CHUNK_VERSION / CURRENT_LEVEL_SUBCHUNK_VERSION
// (WorldDataVersions::CHUNK / SUBCHUNK).
const (
	CurrentLevelChunkVersion    = ChunkVersionV1_21_120
	CurrentLevelSubChunkVersion = SubChunkVersionPalettedMulti
)

// cavesCliffsExperimentalSubChunkKeyOffset is LevelDB::CAVES_CLIFFS_EXPERIMENTAL_SUBCHUNK_KEY_OFFSET.
const cavesCliffsExperimentalSubChunkKeyOffset = 4

// biomeOcean is BiomeIds::OCEAN.
const biomeOcean = 0

// nbtMaxDepth is the NBT reader's depth limit for chunk data.
const nbtMaxDepth = 512

// LevelDB is a port of pocketmine\world\format\io\leveldb\LevelDB.
type LevelDB struct {
	io.BaseWorldProvider
	db *leveldb.DB
}

var _ io.WritableWorldProvider = (*LevelDB)(nil)

func init() {
	io.RegisterBuiltinProvider("leveldb", io.NewWritableWorldProviderManagerEntry(
		IsValid,
		func(path string, logger log.Logger) (io.WritableWorldProvider, error) {
			return NewLevelDB(path, logger)
		},
		Generate,
	), true, 0)
}

// createDB is a port of LevelDB::createDB: Bedrock stores its tables zlib (raw deflate)
// compressed, in 64KB blocks (big enough for most chunks).
func createDB(path string) (*leveldb.DB, error) {
	return leveldb.OpenFile(filepath.Join(path, "db"), &opt.Options{
		Compression: opt.FlateCompression,
		BlockSize:   64 * 1024,
	})
}

// NewLevelDB is a port of LevelDB::__construct.
func NewLevelDB(path string, logger log.Logger) (*LevelDB, error) {
	p := &LevelDB{}
	if err := p.InitBase(path, logger, func() (io.WorldData, error) {
		return worlddata.LoadBedrockWorldData(filepath.Join(path, "level.dat"))
	}); err != nil {
		return nil, err
	}
	db, err := createDB(path)
	if err != nil {
		// we can't tell the difference between errors caused by bad permissions and actual corruption :(
		return nil, &exception.CorruptedWorldError{Message: strings.TrimSpace(err.Error()), Cause: err}
	}
	p.db = db
	return p, nil
}

// GetWorldMinY is a port of LevelDB::getWorldMinY.
func (p *LevelDB) GetWorldMinY() int { return -64 }

// GetWorldMaxY is a port of LevelDB::getWorldMaxY.
func (p *LevelDB) GetWorldMaxY() int { return 320 }

// IsValid is a port of LevelDB::isValid: a level.dat and a db directory.
func IsValid(path string) bool {
	if _, err := os.Stat(filepath.Join(path, "level.dat")); err != nil {
		return false
	}
	info, err := os.Stat(filepath.Join(path, "db"))
	return err == nil && info.IsDir()
}

// Generate is a port of LevelDB::generate.
func Generate(path, name string, options io.WorldCreationOptions) error {
	dbPath := filepath.Join(path, "db")
	if _, err := os.Stat(dbPath); err != nil {
		if err := os.MkdirAll(dbPath, 0o777); err != nil {
			return err
		}
	}
	return worlddata.GenerateBedrockWorldData(path, name, options)
}

// corrupted wraps a stream read error as PHP's CorruptedChunkException.
func corrupted(err error) error {
	var c *exception.CorruptedChunkError
	if errors.As(err, &c) {
		return err
	}
	return exception.WrapCorruptedChunk(err)
}

// deserializeBlockPalette is a port of LevelDB::deserializeBlockPalette.
func (p *LevelDB) deserializeBlockPalette(stream *binaryutils.BinaryStream, logger log.Logger) (*format.PalettedBlockArray, error) {
	b, err := stream.GetByte()
	if err != nil {
		return nil, corrupted(err)
	}
	bitsPerBlock := int(b >> 1)

	wordSize, err := format.GetExpectedWordArraySize(bitsPerBlock)
	if err != nil {
		return nil, exception.NewCorruptedChunkError("Failed to deserialize paletted storage: %v", err)
	}
	words, err := stream.Get(wordSize)
	if err != nil {
		return nil, exception.NewCorruptedChunkError("Failed to deserialize paletted storage: %v", err)
	}
	nbtReader := nbt.NewLittleEndianSerializer()

	var paletteSize int
	if bitsPerBlock == 0 {
		paletteSize = 1
		// Due to code copy-paste in a public plugin, some PM4 worlds have 0 bpb palettes with a
		// length prefix. This is invalid and does not happen in vanilla. These palettes were
		// accepted by PM4 despite being invalid, but PM5 considered them corrupt, causing loss of
		// data. Since many users were affected by this, a workaround is therefore necessary to
		// allow PM5 to read these worlds without data loss.
		// References:
		// - https://github.com/Refaltor77/CustomItemAPI/issues/68
		// - https://github.com/pmmp/PocketMine-MP/issues/5911
		offset := stream.GetOffset()
		byte1, err := stream.GetByte()
		if err != nil {
			return nil, corrupted(err)
		}
		stream.SetOffset(offset) // reset offset

		if byte1 != byte(nbt.TagCompound) { // normally the first byte would be the NBT of the blockstate
			susLength, err := stream.GetLInt()
			if err != nil {
				return nil, corrupted(err)
			}
			if susLength != 1 { // make sure the data isn't complete garbage
				return nil, exception.NewCorruptedChunkError("CustomItemAPI borked 0 bpb palette should always have a length of 1")
			}
			logger.Error("Unexpected palette size for 0 bpb palette")
		}
	} else {
		n, err := stream.GetLInt()
		if err != nil {
			return nil, corrupted(err)
		}
		paletteSize = int(n)
	}

	blockDecodeErrors := map[string][]int{}
	var errorOrder []string
	addError := func(message string, i int) {
		if _, ok := blockDecodeErrors[message]; !ok {
			errorOrder = append(errorOrder, message)
		}
		blockDecodeErrors[message] = append(blockDecodeErrors[message], i)
	}

	palette := make([]int32, 0, max(paletteSize, 0))
	for i := 0; i < paletteSize; i++ {
		root, newOffset, err := nbtReader.Read(stream.GetBuffer(), stream.GetOffset(), nbtMaxDepth)
		if err != nil {
			// NBT borked, unrecoverable
			return nil, exception.NewCorruptedChunkError("Invalid blockstate NBT at offset %d in paletted storage: %v", i, err)
		}
		blockStateNbt, err := root.MustGetCompoundTag()
		if err != nil {
			return nil, exception.NewCorruptedChunkError("Invalid blockstate NBT at offset %d in paletted storage: %v", i, err)
		}
		stream.SetOffset(newOffset)

		//TODO: remember data for unknown states so we can implement them later
		blockStateData, err := p.BlockDataUpgrader.UpgradeBlockStateNbt(blockStateNbt)
		if err != nil {
			// while not ideal, this is not a fatal error
			addError(fmt.Sprintf("Upgrade error: %v, NBT: %s", err, blockStateNbt), i)
			palette = append(palette, p.unknownState())
			continue
		}
		stateID, err := p.BlockStateDeserializer.Deserialize(blockStateData)
		if err != nil {
			var unsupported *blockconvert.UnsupportedBlockStateError
			if errors.As(err, &unsupported) {
				addError(err.Error(), i)
			} else {
				addError(fmt.Sprintf("Deserialize error: %v, NBT: %s", err, blockStateNbt), i)
			}
			palette = append(palette, p.unknownState())
			continue
		}
		palette = append(palette, int32(stateID))
	}

	if len(errorOrder) > 0 {
		finalErrors := make([]string, 0, len(errorOrder))
		for _, message := range errorOrder {
			offsets := make([]string, len(blockDecodeErrors[message]))
			for i, o := range blockDecodeErrors[message] {
				offsets[i] = fmt.Sprint(o)
			}
			finalErrors = append(finalErrors, fmt.Sprintf("%s (palette offsets: %s)", message, strings.Join(offsets, ", ")))
		}
		logger.Error("Errors decoding blocks:\n - " + strings.Join(finalErrors, "\n - "))
	}

	//TODO: exceptions
	result, err := format.NewPalettedBlockArrayFromRaw(bitsPerBlock, words, palette)
	if err != nil {
		return nil, exception.NewCorruptedChunkError("Failed to deserialize paletted storage: %v", err)
	}
	return result, nil
}

// unknownState is the "update!" block every unreadable state becomes.
func (p *LevelDB) unknownState() int32 {
	id, err := p.BlockStateDeserializer.Deserialize(io.GetUnknownBlockStateData())
	if err != nil {
		panic("the unknown block state must always deserialize: " + err.Error())
	}
	return int32(id)
}

// serializeBlockPalette is a port of LevelDB::serializeBlockPalette.
func (p *LevelDB) serializeBlockPalette(stream *binaryutils.BinaryStream, blocks *format.PalettedBlockArray) error {
	stream.PutByte(byte(blocks.GetBitsPerBlock() << 1))
	stream.Put(blocks.GetWordArray())

	palette := blocks.GetPalette()
	if blocks.GetBitsPerBlock() != 0 {
		stream.PutLInt(int32(len(palette)))
	}
	tags := make([]*nbt.TreeRoot, 0, len(palette))
	for _, stateID := range palette {
		data, err := p.BlockStateSerializer.Serialize(int(stateID))
		if err != nil {
			return err
		}
		root, err := nbt.NewTreeRoot(data.ToNbt(), "")
		if err != nil {
			return err
		}
		tags = append(tags, root)
	}
	encoded, err := nbt.NewLittleEndianSerializer().WriteMultiple(tags)
	if err != nil {
		return err
	}
	stream.Put(encoded)
	return nil
}

// getExpected3dBiomesCount is a port of LevelDB::getExpected3dBiomesCount.
func getExpected3dBiomesCount(chunkVersion int) (int, error) {
	switch {
	case chunkVersion >= ChunkVersionV1_18_30:
		return 24, nil
	case chunkVersion >= ChunkVersionV1_18_0_25Beta:
		return 25, nil
	case chunkVersion >= ChunkVersionV1_18_0_24Beta:
		return 32, nil
	case chunkVersion >= ChunkVersionV1_18_0_22Beta:
		return 65, nil
	case chunkVersion >= ChunkVersionV1_17_40_20BetaExperimentalCaves:
		return 32, nil
	}
	return 0, exception.NewCorruptedChunkError("Chunk version %d should not have 3D biomes", chunkVersion)
}

// deserializeBiomePalette is a port of LevelDB::deserializeBiomePalette.
func deserializeBiomePalette(stream *binaryutils.BinaryStream, bitsPerBlock int) (*format.PalettedBlockArray, error) {
	wordSize, err := format.GetExpectedWordArraySize(bitsPerBlock)
	if err != nil {
		return nil, exception.NewCorruptedChunkError("Failed to deserialize paletted biomes: %v", err)
	}
	words, err := stream.Get(wordSize)
	if err != nil {
		return nil, exception.NewCorruptedChunkError("Failed to deserialize paletted biomes: %v", err)
	}
	paletteSize := 1
	if bitsPerBlock != 0 {
		n, err := stream.GetLInt()
		if err != nil {
			return nil, corrupted(err)
		}
		paletteSize = int(n)
	}
	palette := make([]int32, 0, max(paletteSize, 0))
	for i := 0; i < paletteSize; i++ {
		v, err := stream.GetLInt()
		if err != nil {
			return nil, corrupted(err)
		}
		palette = append(palette, v)
	}
	//TODO: exceptions
	result, err := format.NewPalettedBlockArrayFromRaw(bitsPerBlock, words, palette)
	if err != nil {
		return nil, exception.NewCorruptedChunkError("Failed to deserialize paletted biomes: %v", err)
	}
	return result, nil
}

// serializeBiomePalette is a port of LevelDB::serializeBiomePalette.
func serializeBiomePalette(stream *binaryutils.BinaryStream, biomes *format.PalettedBlockArray) {
	stream.PutByte(byte(biomes.GetBitsPerBlock() << 1))
	stream.Put(biomes.GetWordArray())
	palette := biomes.GetPalette()
	if biomes.GetBitsPerBlock() != 0 {
		stream.PutLInt(int32(len(palette)))
	}
	for _, v := range palette {
		stream.PutLInt(v)
	}
}

// deserialize3dBiomes is a port of LevelDB::deserialize3dBiomes.
func deserialize3dBiomes(stream *binaryutils.BinaryStream, chunkVersion int, logger log.Logger) (map[int]*format.PalettedBlockArray, error) {
	var previous *format.PalettedBlockArray
	result := map[int]*format.PalettedBlockArray{}
	nextIndex := format.MinSubChunkIndex

	expectedCount, err := getExpected3dBiomesCount(chunkVersion)
	if err != nil {
		return nil, err
	}
	for i := 0; i < expectedCount; i++ {
		b, err := stream.GetByte()
		if err != nil {
			return nil, exception.NewCorruptedChunkError("Failed to deserialize biome palette %d: %v", i, err)
		}
		bitsPerBlock := int(b >> 1)
		var decoded *format.PalettedBlockArray
		if bitsPerBlock == 127 {
			if previous == nil {
				return nil, exception.NewCorruptedChunkError("Serialized biome palette %d has no previous palette to copy from", i)
			}
			decoded = previous.Clone()
		} else {
			if decoded, err = deserializeBiomePalette(stream, bitsPerBlock); err != nil {
				return nil, exception.NewCorruptedChunkError("Failed to deserialize biome palette %d: %v", i, err)
			}
		}
		previous = decoded
		if nextIndex <= format.MaxSubChunkIndex { // older versions wrote additional superfluous biome palettes
			result[nextIndex] = decoded
			nextIndex++
		} else if stream.Feof() {
			// not enough padding biome arrays for the given version - this is non-critical since we
			// discard the excess anyway, but this should be logged
			logger.Error(fmt.Sprintf("Wrong number of 3D biome palettes for this chunk version: expected %d, but got %d - this is not a problem, but may indicate a corrupted chunk", expectedCount, i+1))
			break
		}
	}
	if !stream.Feof() {
		// maybe bad output produced by a third-party conversion tool like Chunker
		logger.Error("Unexpected trailing data after 3D biomes data")
	}
	return result, nil
}

// serialize3dBiomes is a port of LevelDB::serialize3dBiomes.
func serialize3dBiomes(stream *binaryutils.BinaryStream, subChunks map[int]*format.SubChunk) {
	//TODO: the server-side min/max may not coincide with the world storage min/max - we may need additional logic to handle this
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		//TODO: is it worth trying to use the previous palette if it's the same as the current one? vanilla supports
		//this, but it's not clear if it's worth the effort to implement.
		serializeBiomePalette(stream, subChunks[y].GetBiomeArray())
	}
}

// deserializeExtraDataKey is a port of LevelDB::deserializeExtraDataKey.
func deserializeExtraDataKey(chunkVersion int, key int32) (x, y, z int) {
	if chunkVersion >= ChunkVersionV1_0_0 {
		return int(key>>12) & 0xf, int(key) & 0xff, int(key>>8) & 0xf
	}
	// pre-1.0, 7 bits were used because the build height limit was lower
	return int(key>>11) & 0xf, int(key) & 0x7f, int(key>>7) & 0xf
}

// get is $this->db->get(): the value, or false (ok=false) if the key doesn't exist.
func (p *LevelDB) get(key string) ([]byte, bool, error) {
	v, err := p.db.Get([]byte(key), nil)
	if err != nil {
		if errors.Is(err, leveldb.ErrNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return v, true, nil
}

// deserializeLegacyExtraData is a port of LevelDB::deserializeLegacyExtraData.
func (p *LevelDB) deserializeLegacyExtraData(index string, chunkVersion int, logger log.Logger) (map[int]*format.PalettedBlockArray, error) {
	extraRawData, ok, err := p.get(index + ChunkDataKeyLegacyBlockExtraData)
	if err != nil {
		return nil, err
	}
	if !ok || len(extraRawData) == 0 {
		return map[int]*format.PalettedBlockArray{}, nil
	}

	extraDataLayers := map[int]*format.PalettedBlockArray{}
	stream := binaryutils.NewBinaryStream(extraRawData, 0)
	count, err := stream.GetLInt()
	if err != nil {
		return nil, corrupted(err)
	}
	for i := int32(0); i < count; i++ {
		key, err := stream.GetLInt()
		if err != nil {
			return nil, corrupted(err)
		}
		value, err := stream.GetLShort()
		if err != nil {
			return nil, corrupted(err)
		}
		x, fullY, z := deserializeExtraDataKey(chunkVersion, key)

		ySub := fullY >> 4
		y := int(key) & 0xf

		blockID := int(value) & 0xff
		blockData := int(value>>8) & 0xf
		blockStateData, err := p.BlockDataUpgrader.UpgradeIntIdMeta(blockID, blockData)
		if err != nil {
			//TODO: we could preserve this in case it's supported in the future, but this was historically only
			//used for grass anyway, so we probably don't need to care
			logger.Error(fmt.Sprintf("Failed to upgrade legacy extra block: %v (%d:%d)", err, blockID, blockData))
			continue
		}
		// assume this won't throw
		blockStateID, err := p.BlockStateDeserializer.Deserialize(blockStateData)
		if err != nil {
			blockStateID = int(p.unknownState())
		}
		if extraDataLayers[ySub] == nil {
			extraDataLayers[ySub] = format.NewPalettedBlockArray(io.EmptyStateID())
		}
		extraDataLayers[ySub].Set(x, y, z, int32(blockStateID))
	}
	return extraDataLayers, nil
}

// readVersion is a port of LevelDB::readVersion.
func (p *LevelDB) readVersion(chunkX, chunkZ int) (int, bool, error) {
	index := ChunkIndex(chunkX, chunkZ)
	raw, ok, err := p.get(index + ChunkDataKeyNewVersion)
	if err != nil {
		return 0, false, err
	}
	if !ok {
		if raw, ok, err = p.get(index + ChunkDataKeyOldVersion); err != nil || !ok {
			return 0, false, err
		}
	}
	if len(raw) == 0 {
		return 0, false, exception.NewCorruptedChunkError("Empty chunk version")
	}
	return int(raw[0]), true, nil
}

// deserializeLegacyTerrainData is a port of LevelDB::deserializeLegacyTerrainData: terrain stored
// in the 0.9 full-chunk format.
func (p *LevelDB) deserializeLegacyTerrainData(index string, chunkVersion int, logger log.Logger) (map[int]*format.SubChunk, error) {
	convertedLegacyExtraData, err := p.deserializeLegacyExtraData(index, chunkVersion, logger)
	if err != nil {
		return nil, err
	}

	legacyTerrain, ok, err := p.get(index + ChunkDataKeyLegacyTerrain)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, exception.NewCorruptedChunkError("Missing expected LEGACY_TERRAIN tag for format version %d", chunkVersion)
	}
	stream := binaryutils.NewBinaryStream(legacyTerrain, 0)
	fullIds, err := stream.Get(32768)
	if err != nil {
		return nil, corrupted(err)
	}
	fullData, err := stream.Get(16384)
	if err != nil {
		return nil, corrupted(err)
	}
	if _, err := stream.Get(32768); err != nil { // legacy light info, discard it
		return nil, corrupted(err)
	}
	if _, err := stream.Get(256); err != nil { // heightmap, discard it
		return nil, corrupted(err)
	}
	biomeColorsRaw, err := stream.Get(1024)
	if err != nil {
		return nil, corrupted(err)
	}
	biomeColors := make([]int32, 256)
	for i := range biomeColors {
		biomeColors[i] = int32(uint32(biomeColorsRaw[i*4])<<24 | uint32(biomeColorsRaw[i*4+1])<<16 | uint32(biomeColorsRaw[i*4+2])<<8 | uint32(biomeColorsRaw[i*4+3])) // unpack("N*")
	}
	biomes3d := io.Extrapolate3DBiomes(io.ConvertBiomeColors(biomeColors))
	if !stream.Feof() {
		logger.Error("Unexpected trailing data in legacy terrain data")
	}

	subChunks := map[int]*format.SubChunk{}
	for yy := 0; yy < 8; yy++ {
		storages := []*format.PalettedBlockArray{p.PalettizeLegacySubChunkFromColumn(fullIds, fullData, yy, log.NewPrefixedLogger(logger, fmt.Sprintf("Subchunk y=%d", yy)))}
		if extra, ok := convertedLegacyExtraData[yy]; ok {
			storages = append(storages, extra)
		}
		subChunks[yy] = format.NewSubChunk(io.EmptyStateID(), storages, biomes3d.Clone())
	}

	// make sure extrapolated biomes get filled in correctly
	for yy := format.MinSubChunkIndex; yy <= format.MaxSubChunkIndex; yy++ {
		if _, ok := subChunks[yy]; !ok {
			subChunks[yy] = format.NewSubChunk(io.EmptyStateID(), nil, biomes3d.Clone())
		}
	}
	return subChunks, nil
}

// deserializeNonPalettedSubChunkData is a port of LevelDB::deserializeNonPalettedSubChunkData: a
// subchunk in the legacy non-paletted format used from 1.0 until 1.2.13.
func (p *LevelDB) deserializeNonPalettedSubChunkData(stream *binaryutils.BinaryStream, chunkVersion int, convertedLegacyExtraData *format.PalettedBlockArray, biomePalette *format.PalettedBlockArray, logger log.Logger) (*format.SubChunk, error) {
	blocks, err := stream.Get(4096)
	if err != nil {
		return nil, corrupted(err)
	}
	blockData, err := stream.Get(2048)
	if err != nil {
		return nil, corrupted(err)
	}

	if chunkVersion < ChunkVersionV1_1_0 {
		if _, err := stream.Get(4096); err != nil { // legacy light info, discard it
			logger.Error(fmt.Sprintf("Failed to read legacy subchunk light info: %v", err))
		} else if !stream.Feof() {
			logger.Error("Unexpected trailing data in legacy subchunk data")
		}
	}

	storages := []*format.PalettedBlockArray{p.PalettizeLegacySubChunkXZY(blocks, blockData, logger)}
	if convertedLegacyExtraData != nil {
		storages = append(storages, convertedLegacyExtraData)
	}
	return format.NewSubChunk(io.EmptyStateID(), storages, biomePalette), nil
}

// deserializeSubChunkData is a port of LevelDB::deserializeSubChunkData: a subchunk stored under
// a SUBCHUNK key.
func (p *LevelDB) deserializeSubChunkData(stream *binaryutils.BinaryStream, chunkVersion, subChunkVersion int, convertedLegacyExtraData *format.PalettedBlockArray, biomePalette *format.PalettedBlockArray, logger log.Logger) (*format.SubChunk, error) {
	switch subChunkVersion {
	case SubChunkVersionClassic, SubChunkVersionClassicBug2, SubChunkVersionClassicBug3, SubChunkVersionClassicBug4,
		SubChunkVersionClassicBug5, SubChunkVersionClassicBug6, SubChunkVersionClassicBug7:
		// these are all identical to version 0, but vanilla respects these so we should also
		return p.deserializeNonPalettedSubChunkData(stream, chunkVersion, convertedLegacyExtraData, biomePalette, logger)
	case SubChunkVersionPalettedSingle:
		layer, err := p.deserializeBlockPalette(stream, logger)
		if err != nil {
			return nil, err
		}
		storages := []*format.PalettedBlockArray{layer}
		if convertedLegacyExtraData != nil {
			storages = append(storages, convertedLegacyExtraData)
		}
		return format.NewSubChunk(io.EmptyStateID(), storages, biomePalette), nil
	case SubChunkVersionPalettedMulti, SubChunkVersionPalettedMultiWithOffset:
		// legacy extradata layers intentionally ignored because they aren't supposed to exist in v8
		storageCount, err := stream.GetByte()
		if err != nil {
			return nil, corrupted(err)
		}
		if subChunkVersion >= SubChunkVersionPalettedMultiWithOffset {
			// height ignored; this seems pointless since this is already in the key anyway
			if _, err := stream.GetByte(); err != nil {
				return nil, corrupted(err)
			}
		}
		storages := make([]*format.PalettedBlockArray, 0, storageCount)
		for k := 0; k < int(storageCount); k++ {
			layer, err := p.deserializeBlockPalette(stream, logger)
			if err != nil {
				return nil, err
			}
			storages = append(storages, layer)
		}
		return format.NewSubChunk(io.EmptyStateID(), storages, biomePalette), nil
	}
	// this should never happen - an unsupported chunk appearing in a supported world is a sign of corruption
	return nil, exception.NewCorruptedChunkError("don't know how to decode LevelDB subchunk format version %d", subChunkVersion)
}

// hasOffsetCavesAndCliffsSubChunks is a port of LevelDB::hasOffsetCavesAndCliffsSubChunks.
func hasOffsetCavesAndCliffsSubChunks(chunkVersion int) bool {
	return chunkVersion >= ChunkVersionV1_16_220_50Unused && chunkVersion <= ChunkVersionV1_16_230_50Unused
}

// deserializeAllSubChunkData is a port of LevelDB::deserializeAllSubChunkData.
func (p *LevelDB) deserializeAllSubChunkData(index string, chunkVersion int, hasBeenUpgraded *bool, convertedLegacyExtraData map[int]*format.PalettedBlockArray, biomeArrays map[int]*format.PalettedBlockArray, logger log.Logger) (map[int]*format.SubChunk, error) {
	subChunks := map[int]*format.SubChunk{}

	subChunkKeyOffset := 0
	if hasOffsetCavesAndCliffsSubChunks(chunkVersion) {
		subChunkKeyOffset = cavesCliffsExperimentalSubChunkKeyOffset
	}
	for y := format.MinSubChunkIndex; y <= format.MaxSubChunkIndex; y++ {
		data, ok, err := p.get(index + ChunkDataKeySubChunk + string([]byte{byte((y + subChunkKeyOffset) & 0xff)}))
		if err != nil {
			return nil, err
		}
		if !ok {
			subChunks[y] = format.NewSubChunk(io.EmptyStateID(), nil, biomeArrays[y])
			continue
		}

		stream := binaryutils.NewBinaryStream(data, 0)
		if stream.Feof() {
			return nil, exception.NewCorruptedChunkError("Unexpected empty data for subchunk %d", y)
		}
		v, _ := stream.GetByte()
		subChunkVersion := int(v)
		if subChunkVersion < CurrentLevelSubChunkVersion {
			*hasBeenUpgraded = true
		}

		subChunk, err := p.deserializeSubChunkData(stream, chunkVersion, subChunkVersion, convertedLegacyExtraData[y], biomeArrays[y],
			log.NewPrefixedLogger(logger, fmt.Sprintf("Subchunk y=%d v%d", y, subChunkVersion)))
		if err != nil {
			return nil, err
		}
		subChunks[y] = subChunk
	}
	return subChunks, nil
}

// deserializeBiomeData is a port of LevelDB::deserializeBiomeData: old 2D biomes are extrapolated
// to 3D.
func (p *LevelDB) deserializeBiomeData(index string, chunkVersion int, logger log.Logger) (map[int]*format.PalettedBlockArray, error) {
	biomeArrays := map[int]*format.PalettedBlockArray{}
	maps2d, ok2d, err := p.get(index + ChunkDataKeyHeightmapAnd2DBiomes)
	if err != nil {
		return nil, err
	}
	if ok2d {
		stream := binaryutils.NewBinaryStream(maps2d, 0)
		if _, err := stream.Get(512); err != nil { // heightmap, discard it
			return nil, corrupted(err)
		}
		biomes2d, err := stream.Get(256)
		if err != nil {
			return nil, corrupted(err)
		}
		biomes3d := io.Extrapolate3DBiomes(biomes2d) // never throws
		if !stream.Feof() {
			logger.Error("Unexpected trailing data after 2D biome data")
		}
		for i := format.MinSubChunkIndex; i <= format.MaxSubChunkIndex; i++ {
			biomeArrays[i] = biomes3d.Clone()
		}
		return biomeArrays, nil
	}

	maps3d, ok3d, err := p.get(index + ChunkDataKeyHeightmapAnd3DBiomes)
	if err != nil {
		return nil, err
	}
	if ok3d {
		stream := binaryutils.NewBinaryStream(maps3d, 0)
		if _, err := stream.Get(512); err != nil {
			return nil, corrupted(err)
		}
		return deserialize3dBiomes(stream, chunkVersion, logger)
	}

	logger.Error("Missing biome data, using default ocean biome")
	for i := format.MinSubChunkIndex; i <= format.MaxSubChunkIndex; i++ {
		biomeArrays[i] = format.NewPalettedBlockArray(biomeOcean) // polyfill
	}
	return biomeArrays, nil
}

// LoadChunk is a port of LevelDB::loadChunk.
func (p *LevelDB) LoadChunk(chunkX, chunkZ int) (*io.LoadedChunkData, error) {
	index := ChunkIndex(chunkX, chunkZ)

	chunkVersion, ok, err := p.readVersion(chunkX, chunkZ)
	if err != nil {
		return nil, err
	}
	if !ok {
		//TODO: this might be a slightly-corrupted chunk with a missing version field
		return nil, nil
	}

	//TODO: read PM_DATA_VERSION - we'll need it to fix up old chunks

	logger := log.NewPrefixedLogger(p.Logger, fmt.Sprintf("Loading chunk x=%d z=%d v%d", chunkX, chunkZ, chunkVersion))

	hasBeenUpgraded := chunkVersion < CurrentLevelChunkVersion

	var subChunks map[int]*format.SubChunk
	switch {
	case chunkVersion >= ChunkVersionV1_0_0 && chunkVersion <= ChunkVersionV1_21_120:
		// v1_21_120: TODO ???; v1_21_40: TODO BiomeStates became shorts instead of bytes;
		// v1_16_0_51_beta..v1_16_0: TODO check walls; v1_1_0..: TODO check beds
		convertedLegacyExtraData, err := p.deserializeLegacyExtraData(index, chunkVersion, logger)
		if err != nil {
			return nil, err
		}
		biomeArrays, err := p.deserializeBiomeData(index, chunkVersion, logger)
		if err != nil {
			return nil, err
		}
		if subChunks, err = p.deserializeAllSubChunkData(index, chunkVersion, &hasBeenUpgraded, convertedLegacyExtraData, biomeArrays, logger); err != nil {
			return nil, err
		}
	case chunkVersion == ChunkVersionV0_9_5 || chunkVersion == ChunkVersionV0_9_2 || chunkVersion == ChunkVersionV0_9_0:
		if subChunks, err = p.deserializeLegacyTerrainData(index, chunkVersion, logger); err != nil {
			return nil, err
		}
	default:
		return nil, exception.NewCorruptedChunkError("don't know how to decode chunk format version %d", chunkVersion)
	}

	entities, err := p.readTags(index + ChunkDataKeyEntities)
	if err != nil {
		return nil, err
	}
	tiles, err := p.readTags(index + ChunkDataKeyBlockEntities)
	if err != nil {
		return nil, err
	}

	terrainPopulated := true // older versions didn't have this tag
	finalisation, ok, err := p.get(index + ChunkDataKeyFinalization)
	if err != nil {
		return nil, err
	}
	if ok && len(finalisation) > 0 {
		terrainPopulated = finalisation[0] == finalisationDone
	}

	//TODO: tile ticks, biome states (?)

	return io.NewLoadedChunkData(
		io.NewChunkData(subChunks, terrainPopulated, entities, tiles),
		hasBeenUpgraded,
		io.FixerFlagAll, //TODO: fill this by version rather than just setting all flags
	), nil
}

// readTags reads a list of little-endian NBT compounds (entities, tiles).
func (p *LevelDB) readTags(key string) ([]*nbt.CompoundTag, error) {
	data, ok, err := p.get(key)
	if err != nil {
		return nil, err
	}
	if !ok || len(data) == 0 {
		return nil, nil
	}
	roots, err := nbt.NewLittleEndianSerializer().ReadMultiple(data, nbtMaxDepth)
	if err != nil {
		return nil, exception.WrapCorruptedChunk(err)
	}
	tags := make([]*nbt.CompoundTag, 0, len(roots))
	for _, root := range roots {
		tag, err := root.MustGetCompoundTag()
		if err != nil {
			return nil, exception.WrapCorruptedChunk(err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// SaveChunk is a port of LevelDB::saveChunk.
func (p *LevelDB) SaveChunk(chunkX, chunkZ int, chunkData *io.ChunkData, dirtyFlags int) error {
	index := ChunkIndex(chunkX, chunkZ)

	write := new(leveldb.Batch)
	write.Put([]byte(index+ChunkDataKeyNewVersion), []byte{CurrentLevelChunkVersion})
	write.Put([]byte(index+ChunkDataKeyPmDataVersion), binaryutils.WriteLLong(pocketmine.WorldDataVersion))

	subChunks := chunkData.GetSubChunks()

	if dirtyFlags&format.DirtyFlagBlocks != 0 {
		ys := make([]int, 0, len(subChunks))
		for y := range subChunks {
			ys = append(ys, y)
		}
		sort.Ints(ys)
		for _, y := range ys {
			subChunk := subChunks[y]
			key := []byte(index + ChunkDataKeySubChunk + string([]byte{byte(y & 0xff)}))
			if subChunk.IsEmptyAuthoritative() {
				write.Delete(key)
				continue
			}
			subStream := binaryutils.NewBinaryStream(nil, 0)
			subStream.PutByte(CurrentLevelSubChunkVersion)

			layers := subChunk.GetBlockLayers()
			subStream.PutByte(byte(len(layers)))
			for _, blocks := range layers {
				if err := p.serializeBlockPalette(subStream, blocks); err != nil {
					return fmt.Errorf("saving chunk x=%d z=%d subchunk %d: %w", chunkX, chunkZ, y, err)
				}
			}
			write.Put(key, subStream.GetBuffer())
		}
	}

	if dirtyFlags&format.DirtyFlagBiomes != 0 {
		write.Delete([]byte(index + ChunkDataKeyHeightmapAnd2DBiomes))
		stream := binaryutils.NewBinaryStream(nil, 0)
		stream.Put(make([]byte, 512)) // fake heightmap
		serialize3dBiomes(stream, subChunks)
		write.Put([]byte(index+ChunkDataKeyHeightmapAnd3DBiomes), stream.GetBuffer())
	}

	//TODO: use this properly
	finalisation := byte(finalisationNeedsPopulation)
	if chunkData.IsPopulated() {
		finalisation = finalisationDone
	}
	write.Put([]byte(index+ChunkDataKeyFinalization), []byte{finalisation})

	if err := writeTags(chunkData.GetTileNBT(), index+ChunkDataKeyBlockEntities, write); err != nil {
		return err
	}
	if err := writeTags(chunkData.GetEntityNBT(), index+ChunkDataKeyEntities, write); err != nil {
		return err
	}

	write.Delete([]byte(index + ChunkDataKeyHeightmapAnd2DBiomeColors))
	write.Delete([]byte(index + ChunkDataKeyLegacyTerrain))

	return p.db.Write(write, nil)
}

// writeTags is a port of LevelDB::writeTags.
func writeTags(targets []*nbt.CompoundTag, index string, write *leveldb.Batch) error {
	if len(targets) == 0 {
		write.Delete([]byte(index))
		return nil
	}
	roots := make([]*nbt.TreeRoot, 0, len(targets))
	for _, tag := range targets {
		root, err := nbt.NewTreeRoot(tag, "")
		if err != nil {
			return err
		}
		roots = append(roots, root)
	}
	data, err := nbt.NewLittleEndianSerializer().WriteMultiple(roots)
	if err != nil {
		return err
	}
	write.Put([]byte(index), data)
	return nil
}

// GetDatabase is a port of LevelDB::getDatabase.
func (p *LevelDB) GetDatabase() *leveldb.DB { return p.db }

// ChunkIndex is a port of LevelDB::chunkIndex: the key prefix of a chunk's entries.
func ChunkIndex(chunkX, chunkZ int) string {
	return string(binaryutils.WriteLInt(int32(chunkX))) + string(binaryutils.WriteLInt(int32(chunkZ)))
}

// DoGarbageCollection is a port of LevelDB::doGarbageCollection (nothing to do).
func (p *LevelDB) DoGarbageCollection() {}

// Close is a port of LevelDB::close.
func (p *LevelDB) Close() error {
	if p.db == nil {
		return nil
	}
	err := p.db.Close()
	p.db = nil
	return err
}

// chunkKeys calls fn with the coordinates of every chunk version key (NEW_VERSION or
// OLD_VERSION).
func (p *LevelDB) chunkKeys(fn func(chunkX, chunkZ int) bool) error {
	iter := p.db.NewIterator(nil, nil)
	defer iter.Release()
	for iter.Next() {
		key := iter.Key()
		if len(key) == 9 && (key[8] == ChunkDataKeyNewVersion[0] || key[8] == ChunkDataKeyOldVersion[0]) {
			chunkX, _ := binaryutils.ReadLInt(key[0:4])
			chunkZ, _ := binaryutils.ReadLInt(key[4:8])
			if !fn(int(chunkX), int(chunkZ)) {
				break
			}
		}
	}
	return iter.Error()
}

// GetAllChunks is a port of LevelDB::getAllChunks.
func (p *LevelDB) GetAllChunks(skipCorrupted bool, logger log.Logger, yield func(coords io.ChunkCoords, chunk *io.LoadedChunkData) bool) error {
	// Collect the coordinates first: loading chunks while iterating would keep the iterator's
	// snapshot open for the whole conversion.
	var coords []io.ChunkCoords
	if err := p.chunkKeys(func(x, z int) bool {
		coords = append(coords, io.ChunkCoords{X: x, Z: z})
		return true
	}); err != nil {
		return err
	}
	for _, c := range coords {
		chunk, err := p.LoadChunk(c.X, c.Z)
		if err != nil {
			var corruptedErr *exception.CorruptedChunkError
			if !errors.As(err, &corruptedErr) || !skipCorrupted {
				return err
			}
			if logger != nil {
				logger.Error(fmt.Sprintf("Skipped corrupted chunk %d %d (%v)", c.X, c.Z, err))
			}
			continue
		}
		if chunk != nil && !yield(c, chunk) {
			return nil
		}
	}
	return nil
}

// CalculateChunkCount is a port of LevelDB::calculateChunkCount.
func (p *LevelDB) CalculateChunkCount() (int, error) {
	count := 0
	err := p.chunkKeys(func(int, int) bool {
		count++
		return true
	})
	return count, err
}
