package worlddata

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sandertv/gophertunnel/minecraft/protocol"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/format/io"
	"pocketmine-go/pocketmine/world/format/io/exception"
)

// BedrockWorldData version constants, a port of WorldDataVersions::STORAGE/NETWORK/LAST_OPENED_IN.
const (
	CurrentStorageVersion = 10
	// CurrentStorageNetworkVersion is the highest NetworkVersion of Bedrock worlds this server
	// loads. PocketMine-MP 5.44.4 caps it at 924 (1.26.0); this port supports the 1.26.50 client
	// (the pinned gophertunnel's protocol), whose worlds are saved with NetworkVersion 2193, and
	// reads their chunk and blockstate format, so the cap is the current protocol.
	CurrentStorageNetworkVersion = protocol.CurrentProtocol
)

// CurrentClientVersionTarget is WorldDataVersions::LAST_OPENED_IN.
var CurrentClientVersionTarget = []int32{1, 26, 0, 2, 0}

// Generator types, a port of BedrockWorldData::GENERATOR_*.
const (
	GeneratorLimited  = 0
	GeneratorInfinite = 1
	GeneratorFlat     = 2
)

// BedrockWorldData tag names, a port of BedrockWorldData::TAG_*.
const (
	tagDayCycleStopTime      = "DayCycleStopTime"
	tagDifficulty            = "Difficulty"
	tagForceGameType         = "ForceGameType"
	tagGameType              = "GameType"
	tagGenerator             = "Generator"
	tagLastPlayed            = "LastPlayed"
	tagNetworkVersion        = "NetworkVersion"
	tagStorageVersion        = "StorageVersion"
	tagIsEdu                 = "eduLevel"
	tagFallDamageEnabled     = "falldamage"
	tagFireDamageEnabled     = "firedamage"
	tagAchievementsDisabled  = "hasBeenLoadedInCreative"
	tagImmutableWorld        = "immutableWorld"
	tagLightningLevel        = "lightningLevel"
	tagLightningTime         = "lightningTime"
	tagPvpEnabled            = "pvp"
	tagRainLevel             = "rainLevel"
	tagRainTime              = "rainTime"
	tagSpawnMobs             = "spawnMobs"
	tagTexturePacksRequired  = "texturePacksRequired"
	tagLastOpenedWithVersion = "lastOpenedWithVersion"
	tagCommandsEnabled       = "commandsEnabled"
)

// BedrockWorldData is a port of pocketmine\world\format\io\data\BedrockWorldData: a LevelDB
// world's level.dat (an 8-byte header followed by little-endian NBT).
type BedrockWorldData struct {
	baseNbtWorldData
}

var _ io.WorldData = (*BedrockWorldData)(nil)

func lastOpenedWithVersionTag() *nbt.ListTag {
	values := make([]nbt.Tag, len(CurrentClientVersionTarget))
	for i, v := range CurrentClientVersionTarget {
		values[i] = nbt.IntTag(v)
	}
	list, err := nbt.NewListTag(values, nbt.TagInt)
	if err != nil {
		panic(err)
	}
	return list
}

func encodeLevelDat(tag *nbt.CompoundTag) ([]byte, error) {
	root, err := nbt.NewTreeRoot(tag, "")
	if err != nil {
		return nil, err
	}
	buffer, err := nbt.NewLittleEndianSerializer().Write(root)
	if err != nil {
		return nil, err
	}
	header := append(binaryutils.WriteLInt(CurrentStorageVersion), binaryutils.WriteLInt(int32(len(buffer)))...)
	return append(header, buffer...), nil
}

// GenerateBedrockWorldData is a port of BedrockWorldData::generate: writes a new level.dat into
// path (the world's directory).
func GenerateBedrockWorldData(path, name string, options io.WorldCreationOptions) error {
	generatorType := GeneratorInfinite
	if options.GeneratorName == "flat" {
		generatorType = GeneratorFlat
	}
	//TODO: add support for limited worlds

	worldData := nbt.NewCompoundTag().
		// Vanilla fields
		SetInt(tagDayCycleStopTime, -1).
		SetInt(tagDifficulty, nbt.IntTag(options.Difficulty)).
		SetByte(tagForceGameType, 0).
		SetInt(tagGameType, 0).
		SetInt(tagGenerator, nbt.IntTag(generatorType)).
		SetLong(tagLastPlayed, nbt.LongTag(time.Now().Unix())).
		SetString(tagLevelName, nbt.StringTag(name)).
		SetInt(tagNetworkVersion, CurrentStorageNetworkVersion).
		//->setInt("Platform", 2) //TODO: find out what the possible values are for
		SetLong(tagRandomSeed, nbt.LongTag(options.Seed)).
		SetInt(tagSpawnX, nbt.IntTag(options.SpawnPosition.FloorX())).
		SetInt(tagSpawnY, nbt.IntTag(options.SpawnPosition.FloorY())).
		SetInt(tagSpawnZ, nbt.IntTag(options.SpawnPosition.FloorZ())).
		SetInt(tagStorageVersion, CurrentStorageVersion).
		SetLong(tagTime, 0).
		SetByte(tagIsEdu, 0).
		SetByte(tagFallDamageEnabled, 1).
		SetByte(tagFireDamageEnabled, 1).
		SetByte(tagAchievementsDisabled, 1). // badly named, this actually determines whether achievements can be earned in this world...
		SetByte(tagImmutableWorld, 0).
		SetFloat(tagLightningLevel, 0).
		SetInt(tagLightningTime, 0).
		SetByte(tagPvpEnabled, 1).
		SetFloat(tagRainLevel, 0).
		SetInt(tagRainTime, 0).
		SetByte(tagSpawnMobs, 1).
		SetByte(tagTexturePacksRequired, 0). //TODO
		SetByte(tagCommandsEnabled, 1).
		SetTag(tagLastOpenedWithVersion, lastOpenedWithVersionTag()).
		// Additional PocketMine-MP fields
		SetString(tagGeneratorName, nbt.StringTag(options.GeneratorName)).
		SetString(tagGeneratorOptions, nbt.StringTag(options.GeneratorOptions))

	data, err := encodeLevelDat(worldData)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, "level.dat"), data, 0o644)
}

// LoadBedrockWorldData is a port of `new BedrockWorldData($dataPath)`: reads, validates and fixes
// the level.dat at dataPath.
func LoadBedrockWorldData(dataPath string) (*BedrockWorldData, error) {
	d := &BedrockWorldData{}
	base, err := newBaseNbtWorldData(dataPath, func() (*nbt.CompoundTag, error) { return loadBedrockLevelDat(dataPath) }, fixBedrockLevelDat)
	if err != nil {
		return nil, err
	}
	d.baseNbtWorldData = base
	return d, nil
}

// loadBedrockLevelDat is a port of BedrockWorldData::load.
func loadBedrockLevelDat(dataPath string) (*nbt.CompoundTag, error) {
	raw, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, &exception.CorruptedWorldError{Message: err.Error(), Cause: err}
	}
	if len(raw) <= 8 {
		return nil, exception.NewCorruptedWorldError("Truncated level.dat")
	}
	root, _, err := nbt.NewLittleEndianSerializer().Read(raw, 8, nbtMaxDepth)
	if err != nil {
		return nil, &exception.CorruptedWorldError{Message: err.Error(), Cause: err}
	}
	worldData, err := root.MustGetCompoundTag()
	if err != nil {
		return nil, &exception.CorruptedWorldError{Message: err.Error(), Cause: err}
	}

	version, err := worldData.GetInt(tagStorageVersion)
	if err != nil {
		return nil, exception.NewCorruptedWorldError("Missing '%s' tag in level.dat", tagStorageVersion)
	}
	if version > CurrentStorageVersion {
		return nil, exception.NewUnsupportedWorldFormatError("LevelDB world format version %d is currently unsupported", version)
	}
	// StorageVersion is rarely updated - instead, the game relies on the NetworkVersion tag, which
	// is synced with the network protocol version for that version.
	protocolVersion, err := worldData.GetInt(tagNetworkVersion)
	if err != nil {
		return nil, exception.NewCorruptedWorldError("Missing '%s' tag in level.dat", tagNetworkVersion)
	}
	if protocolVersion > CurrentStorageNetworkVersion {
		return nil, exception.NewUnsupportedWorldFormatError("LevelDB world protocol version %d is currently unsupported", protocolVersion)
	}
	return worldData, nil
}

// fixBedrockLevelDat is a port of BedrockWorldData::fix.
func fixBedrockLevelDat(tag *nbt.CompoundTag) error {
	generatorNameTag, _ := tag.GetTag(tagGeneratorName)
	if generatorName, ok := generatorNameTag.(nbt.StringTag); !ok {
		mcpeGeneratorTypeTag, _ := tag.GetTag(tagGenerator)
		if generatorType, ok := mcpeGeneratorTypeTag.(nbt.IntTag); ok {
			switch generatorType { // Detect correct generator from MCPE data
			case GeneratorFlat:
				tag.SetString(tagGeneratorName, "flat")
				tag.SetString(tagGeneratorOptions, "2;7,3,3,2;1")
			case GeneratorInfinite:
				//TODO: add a null generator which does not generate missing chunks (to allow importing back to MCPE and generating more normal terrain without PocketMine messing things up)
				tag.SetString(tagGeneratorName, "default")
				tag.SetString(tagGeneratorOptions, "")
			case GeneratorLimited:
				return exception.NewUnsupportedWorldFormatError("Limited worlds are not currently supported")
			default:
				return exception.NewUnsupportedWorldFormatError("Unknown LevelDB generator type")
			}
		} else {
			tag.SetString(tagGeneratorName, "default")
		}
	} else if fixed, ok := hackyFixForGeneratorClasspathInLevelDat(string(generatorName)); ok {
		tag.SetString(tagGeneratorName, nbt.StringTag(fixed))
	}

	if t, _ := tag.GetTag(tagGeneratorOptions); t == nil {
		tag.SetString(tagGeneratorOptions, "")
	} else if _, ok := t.(nbt.StringTag); !ok {
		tag.SetString(tagGeneratorOptions, "")
	}
	return nil
}

// Save is a port of BedrockWorldData::save.
func (d *BedrockWorldData) Save() error {
	d.compoundTag.SetInt(tagNetworkVersion, CurrentStorageNetworkVersion)
	d.compoundTag.SetInt(tagStorageVersion, CurrentStorageVersion)
	d.compoundTag.SetTag(tagLastOpenedWithVersion, lastOpenedWithVersionTag())
	d.compoundTag.SetLong(pocketmine.TagWorldDataVersion, nbt.LongTag(pocketmine.WorldDataVersion))

	data, err := encodeLevelDat(d.compoundTag)
	if err != nil {
		return err
	}
	return utils.SafeFilePutContents(d.dataPath, data)
}

func (d *BedrockWorldData) GetDifficulty() int {
	return int(d.compoundTag.GetIntOr(tagDifficulty, io.DifficultyNormal))
}

// SetDifficulty is a port of BedrockWorldData::setDifficulty (yes, a TAG_Int: in PE it's an int,
// in PC a byte).
func (d *BedrockWorldData) SetDifficulty(difficulty int) {
	d.compoundTag.SetInt(tagDifficulty, nbt.IntTag(difficulty))
}

func (d *BedrockWorldData) GetRainTime() int { return int(d.compoundTag.GetIntOr(tagRainTime, 0)) }

func (d *BedrockWorldData) SetRainTime(ticks int) {
	d.compoundTag.SetInt(tagRainTime, nbt.IntTag(ticks))
}

func (d *BedrockWorldData) GetRainLevel() float64 {
	return float64(d.compoundTag.GetFloatOr(tagRainLevel, 0))
}

func (d *BedrockWorldData) SetRainLevel(level float64) {
	d.compoundTag.SetFloat(tagRainLevel, nbt.FloatTag(level))
}

func (d *BedrockWorldData) GetLightningTime() int {
	return int(d.compoundTag.GetIntOr(tagLightningTime, 0))
}

func (d *BedrockWorldData) SetLightningTime(ticks int) {
	d.compoundTag.SetInt(tagLightningTime, nbt.IntTag(ticks))
}

func (d *BedrockWorldData) GetLightningLevel() float64 {
	return float64(d.compoundTag.GetFloatOr(tagLightningLevel, 0))
}

func (d *BedrockWorldData) SetLightningLevel(level float64) {
	d.compoundTag.SetFloat(tagLightningLevel, nbt.FloatTag(level))
}
