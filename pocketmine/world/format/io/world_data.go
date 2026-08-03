// Package io is a port of a slice of pocketmine\world\format\io: level.dat world metadata
// (BedrockWorldData/BaseNbtWorldData) - real little-endian NBT, read and written exactly as a
// vanilla Bedrock client or real PocketMine-MP would, so a world directory this port creates is a
// genuine, portable Bedrock world save, not just something this port itself can read back.
//
// Named "io" (not "worlddata") to mirror pocketmine\world\format\io's own namespace - the LevelDB
// chunk-storage code already ported lives one level down, in the sibling io/leveldb package (chunk
// storage and level.dat are separate concerns in the real class hierarchy too - LevelDBWorldData
// glues them together in a full WorldProvider, which this port doesn't have as a unified type,
// each caller uses this package and io/leveldb directly instead).
package io

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// nbtMaxDepth mirrors io/leveldb's own constant of the same name/value (see its doc comment) -
// duplicated rather than exported cross-package since it's a generic NBT-reader safety limit, not
// something specific to either package.
const nbtMaxDepth = 512

// Generator type constants, a port of BedrockWorldData::GENERATOR_LIMITED/INFINITE/FLAT - the
// legacy vanilla-compatible "Generator" int tag (as opposed to the PocketMine-specific
// generatorName string tag, which is what this port's own code actually reads back).
const (
	GeneratorLimited  = 0
	GeneratorInfinite = 1
	GeneratorFlat     = 2
)

// Difficulty constants, a port of World::DIFFICULTY_PEACEFUL/EASY/NORMAL/HARD - not otherwise
// ported anywhere in this port's own World type yet (no difficulty-gated gameplay exists), needed
// here purely as level.dat's own stored/default value.
const (
	DifficultyPeaceful = 0
	DifficultyEasy     = 1
	DifficultyNormal   = 2
	DifficultyHard     = 3
)

// currentStorageVersion/currentStorageNetworkVersion/currentClientVersionTarget mirror
// WorldDataVersions::STORAGE/NETWORK/LAST_OPENED_IN - copied from the real PHP constants, not
// guessed (see data/bedrock/WorldDataVersions.php).
const (
	currentStorageVersion        = 10
	currentStorageNetworkVersion = 924
)

var currentClientVersionTarget = []int32{1, 26, 0, 2, 0}

// level.dat NBT tag names - a port of BaseNbtWorldData's and BedrockWorldData's own TAG_* private
// constants.
const (
	tagLevelName        = "LevelName"
	tagGeneratorName    = "generatorName"
	tagGeneratorOptions = "generatorOptions"
	tagRandomSeed       = "RandomSeed"
	tagTime             = "Time"
	tagSpawnX           = "SpawnX"
	tagSpawnY           = "SpawnY"
	tagSpawnZ           = "SpawnZ"

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
	tagPVPEnabled            = "pvp"
	tagRainLevel             = "rainLevel"
	tagRainTime              = "rainTime"
	tagSpawnMobs             = "spawnMobs"
	tagTexturePacksRequired  = "texturePacksRequired"
	tagLastOpenedWithVersion = "lastOpenedWithVersion"
	tagCommandsEnabled       = "commandsEnabled"
)

// levelDatFileName is the fixed file name a Bedrock world's metadata always lives at, directly
// inside the world's own directory.
const levelDatFileName = "level.dat"

// WorldData is a port of the relevant slice of BedrockWorldData/BaseNbtWorldData.
type WorldData struct {
	tag *nbt.CompoundTag
}

// GenerateWorldData is a port of BedrockWorldData::generate: writes a brand new level.dat for a
// freshly created world at path (the world's own directory - level.dat is created directly inside
// it), and returns the WorldData wrapping it (equivalent to immediately loading it back, saving
// callers a redundant read).
func GenerateWorldData(path, name string, seed int64, generatorType int, generatorName, generatorOptions string, spawn math.Vector3) (*WorldData, error) {
	lastOpened := make([]nbt.Tag, len(currentClientVersionTarget))
	for i, v := range currentClientVersionTarget {
		lastOpened[i] = nbt.IntTag(v)
	}
	lastOpenedList, err := nbt.NewListTag(lastOpened, nbt.TagInt)
	if err != nil {
		return nil, err
	}

	tag := nbt.NewCompoundTag().
		SetInt("DayCycleStopTime", -1).
		SetInt(tagDifficulty, DifficultyNormal).
		SetByte(tagForceGameType, 0).
		SetInt(tagGameType, 0).
		SetInt(tagGenerator, nbt.IntTag(generatorType)).
		SetLong(tagLastPlayed, nbt.LongTag(time.Now().Unix())).
		SetString(tagLevelName, nbt.StringTag(name)).
		SetInt(tagNetworkVersion, currentStorageNetworkVersion).
		SetLong(tagRandomSeed, nbt.LongTag(seed)).
		SetInt(tagSpawnX, nbt.IntTag(spawn.FloorX())).
		SetInt(tagSpawnY, nbt.IntTag(spawn.FloorY())).
		SetInt(tagSpawnZ, nbt.IntTag(spawn.FloorZ())).
		SetInt(tagStorageVersion, currentStorageVersion).
		SetLong(tagTime, 0).
		SetByte(tagIsEdu, 0).
		SetByte(tagFallDamageEnabled, 1).
		SetByte(tagFireDamageEnabled, 1).
		SetByte(tagAchievementsDisabled, 1).
		SetByte(tagImmutableWorld, 0).
		SetFloat(tagLightningLevel, 0).
		SetInt(tagLightningTime, 0).
		SetByte(tagPVPEnabled, 1).
		SetFloat(tagRainLevel, 0).
		SetInt(tagRainTime, 0).
		SetByte(tagSpawnMobs, 1).
		SetByte(tagTexturePacksRequired, 0).
		SetByte(tagCommandsEnabled, 1).
		SetTag(tagLastOpenedWithVersion, lastOpenedList).
		SetString(tagGeneratorName, nbt.StringTag(generatorName)).
		SetString(tagGeneratorOptions, nbt.StringTag(generatorOptions))

	wd := &WorldData{tag: tag}
	if err := wd.Save(path); err != nil {
		return nil, err
	}
	return wd, nil
}

// LoadWorldData is a port of BedrockWorldData's constructor (load() + the version checks inlined
// into it + fix()): reads and validates an existing level.dat at path (the world's own directory).
func LoadWorldData(path string) (*WorldData, error) {
	raw, err := os.ReadFile(filepath.Join(path, levelDatFileName))
	if err != nil {
		return nil, fmt.Errorf("world data: reading level.dat: %w", err)
	}
	if len(raw) <= 8 {
		return nil, fmt.Errorf("world data: truncated level.dat")
	}

	deserializer := nbt.NewLittleEndianSerializer()
	root, _, err := deserializer.Read(raw, 8, nbtMaxDepth)
	if err != nil {
		return nil, fmt.Errorf("world data: decoding level.dat: %w", err)
	}
	tag, err := root.MustGetCompoundTag()
	if err != nil {
		return nil, fmt.Errorf("world data: level.dat root is not a compound tag: %w", err)
	}

	version, err := tag.GetInt(tagStorageVersion)
	if err != nil {
		return nil, fmt.Errorf("world data: missing %q tag in level.dat", tagStorageVersion)
	}
	if int(version) > currentStorageVersion {
		return nil, fmt.Errorf("world data: LevelDB world format version %d is currently unsupported", version)
	}

	// StorageVersion is rarely updated - the game instead relies on NetworkVersion, which tracks
	// the network protocol version.
	protocolVersion, err := tag.GetInt(tagNetworkVersion)
	if err != nil {
		return nil, fmt.Errorf("world data: missing %q tag in level.dat", tagNetworkVersion)
	}
	if int(protocolVersion) > currentStorageNetworkVersion {
		return nil, fmt.Errorf("world data: LevelDB world protocol version %d is currently unsupported", protocolVersion)
	}

	wd := &WorldData{tag: tag}
	wd.fix()
	return wd, nil
}

// fix is a port of BedrockWorldData::fix - minus the classpath-hack branch
// (hackyFixForGeneratorClasspathInLevelDat only ever fires for worlds broken by a PocketMine-MP
// version that predates this port's own existence, so replicating that specific historical bug fix
// would be pure dead code here) and minus GENERATOR_LIMITED support (real PHP itself doesn't
// support it either - it throws UnsupportedWorldFormatException, which this port matches by simply
// never producing that case in GenerateWorldData to begin with).
func (d *WorldData) fix() {
	if _, err := d.tag.GetString(tagGeneratorName); err != nil {
		if genType, genErr := d.tag.GetInt(tagGenerator); genErr == nil {
			switch genType {
			case GeneratorFlat:
				d.tag.SetString(tagGeneratorName, "flat")
				d.tag.SetString(tagGeneratorOptions, "2;7,3,3,2;1")
			default:
				d.tag.SetString(tagGeneratorName, "default")
				d.tag.SetString(tagGeneratorOptions, "")
			}
		} else {
			d.tag.SetString(tagGeneratorName, "default")
		}
	}
	if _, err := d.tag.GetString(tagGeneratorOptions); err != nil {
		d.tag.SetString(tagGeneratorOptions, "")
	}
}

// Save is a port of BedrockWorldData::save: writes this WorldData back to path/level.dat (the
// world's own directory), stamping the current storage/network version and this port's own
// PMMPDataVersion tag, exactly as real PocketMine-MP does on every save.
func (d *WorldData) Save(path string) error {
	d.tag.SetInt(tagNetworkVersion, currentStorageNetworkVersion)
	d.tag.SetInt(tagStorageVersion, currentStorageVersion)

	lastOpened := make([]nbt.Tag, len(currentClientVersionTarget))
	for i, v := range currentClientVersionTarget {
		lastOpened[i] = nbt.IntTag(v)
	}
	lastOpenedList, err := nbt.NewListTag(lastOpened, nbt.TagInt)
	if err != nil {
		return err
	}
	d.tag.SetTag(tagLastOpenedWithVersion, lastOpenedList)
	d.tag.SetLong(pocketmine.TagWorldDataVersion, nbt.LongTag(pocketmine.WorldDataVersion))

	root, err := nbt.NewTreeRoot(d.tag, "")
	if err != nil {
		return err
	}
	buffer, err := nbt.NewLittleEndianSerializer().Write(root)
	if err != nil {
		return fmt.Errorf("world data: encoding level.dat: %w", err)
	}

	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("world data: creating world directory %q: %w", path, err)
	}

	header := append(binaryutils.WriteLInt(int32(currentStorageVersion)), binaryutils.WriteLInt(int32(len(buffer)))...)
	if err := os.WriteFile(filepath.Join(path, levelDatFileName), append(header, buffer...), 0o644); err != nil {
		return fmt.Errorf("world data: writing level.dat: %w", err)
	}
	return nil
}

func (d *WorldData) GetName() string { return string(d.tag.GetStringOr(tagLevelName, "")) }

func (d *WorldData) SetName(name string) { d.tag.SetString(tagLevelName, nbt.StringTag(name)) }

func (d *WorldData) GetGenerator() string {
	return string(d.tag.GetStringOr(tagGeneratorName, "default"))
}

func (d *WorldData) GetGeneratorOptions() string {
	return string(d.tag.GetStringOr(tagGeneratorOptions, ""))
}

func (d *WorldData) GetSeed() int64 { return int64(d.tag.GetLongOr(tagRandomSeed, 0)) }

// GetTime is a port of BaseNbtWorldData::getTime, including its "some older PM worlds had this in
// the wrong (Int, not Long) format" fallback.
func (d *WorldData) GetTime() int64 {
	if v, err := d.tag.GetInt(tagTime); err == nil {
		return int64(v)
	}
	return int64(d.tag.GetLongOr(tagTime, 0))
}

func (d *WorldData) SetTime(t int64) { d.tag.SetLong(tagTime, nbt.LongTag(t)) }

func (d *WorldData) GetSpawn() math.Vector3 {
	return math.NewVector3(
		float64(d.tag.GetIntOr(tagSpawnX, 0)),
		float64(d.tag.GetIntOr(tagSpawnY, 0)),
		float64(d.tag.GetIntOr(tagSpawnZ, 0)),
	)
}

func (d *WorldData) SetSpawn(pos math.Vector3) {
	d.tag.SetInt(tagSpawnX, nbt.IntTag(pos.FloorX()))
	d.tag.SetInt(tagSpawnY, nbt.IntTag(pos.FloorY()))
	d.tag.SetInt(tagSpawnZ, nbt.IntTag(pos.FloorZ()))
}

func (d *WorldData) GetDifficulty() int { return int(d.tag.GetIntOr(tagDifficulty, DifficultyNormal)) }

func (d *WorldData) SetDifficulty(difficulty int) {
	d.tag.SetInt(tagDifficulty, nbt.IntTag(difficulty))
}

func (d *WorldData) GetRainTime() int { return int(d.tag.GetIntOr(tagRainTime, 0)) }

func (d *WorldData) SetRainTime(ticks int) { d.tag.SetInt(tagRainTime, nbt.IntTag(ticks)) }

func (d *WorldData) GetRainLevel() float64 { return float64(d.tag.GetFloatOr(tagRainLevel, 0)) }

func (d *WorldData) SetRainLevel(level float64) { d.tag.SetFloat(tagRainLevel, nbt.FloatTag(level)) }

func (d *WorldData) GetLightningTime() int { return int(d.tag.GetIntOr(tagLightningTime, 0)) }

func (d *WorldData) SetLightningTime(ticks int) { d.tag.SetInt(tagLightningTime, nbt.IntTag(ticks)) }

func (d *WorldData) GetLightningLevel() float64 {
	return float64(d.tag.GetFloatOr(tagLightningLevel, 0))
}

func (d *WorldData) SetLightningLevel(level float64) {
	d.tag.SetFloat(tagLightningLevel, nbt.FloatTag(level))
}
