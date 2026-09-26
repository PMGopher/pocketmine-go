// Package worlddata is a port of pocketmine\world\format\io\data: level.dat implementations for
// Bedrock (LevelDB) and Java (region) worlds.
package worlddata

import (
	"os"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format/io/exception"
)

// nbtMaxDepth is the NBT reader's depth limit for level.dat.
const nbtMaxDepth = 512

// BaseNbtWorldData tag names, a port of BaseNbtWorldData::TAG_*.
const (
	tagLevelName        = "LevelName"
	tagGeneratorName    = "generatorName"
	tagGeneratorOptions = "generatorOptions"
	tagRandomSeed       = "RandomSeed"
	tagTime             = "Time"
	tagSpawnX           = "SpawnX"
	tagSpawnY           = "SpawnY"
	tagSpawnZ           = "SpawnZ"
)

// baseNbtWorldData is a port of pocketmine\world\format\io\data\BaseNbtWorldData: the fields
// common to Bedrock and Java level.dat files. Concrete types embed it and set compoundTag through
// their load() and fix().
type baseNbtWorldData struct {
	dataPath    string
	compoundTag *nbt.CompoundTag
}

// newBaseNbtWorldData is BaseNbtWorldData::__construct: load and fix are the concrete type's
// load() and fix().
func newBaseNbtWorldData(dataPath string, load func() (*nbt.CompoundTag, error), fix func(*nbt.CompoundTag) error) (baseNbtWorldData, error) {
	if _, err := os.Stat(dataPath); err != nil {
		return baseNbtWorldData{}, exception.NewCorruptedWorldError("World data not found at %s", dataPath)
	}
	tag, err := load()
	if err != nil {
		if corrupted, ok := err.(*exception.CorruptedWorldError); ok {
			return baseNbtWorldData{}, &exception.CorruptedWorldError{Message: "Corrupted world data: " + corrupted.Message, Cause: corrupted}
		}
		return baseNbtWorldData{}, err
	}
	if err := fix(tag); err != nil {
		return baseNbtWorldData{}, err
	}
	return baseNbtWorldData{dataPath: dataPath, compoundTag: tag}, nil
}

// hackyFixForGeneratorClasspathInLevelDat is a port of
// BaseNbtWorldData::hackyFixForGeneratorClasspathInLevelDat: older PocketMine-MP versions saved
// generator class paths into level.dat of imported worlds. These are deliberately hardcoded.
func hackyFixForGeneratorClasspathInLevelDat(className string) (string, bool) {
	switch className {
	case `pocketmine\level\generator\normal\Normal`:
		return "normal", true
	case `pocketmine\level\generator\Flat`:
		return "flat", true
	}
	return "", false
}

// GetCompoundTag is a port of BaseNbtWorldData::getCompoundTag.
func (d *baseNbtWorldData) GetCompoundTag() *nbt.CompoundTag { return d.compoundTag }

func (d *baseNbtWorldData) GetName() string {
	return string(d.compoundTag.GetStringOr(tagLevelName, ""))
}

func (d *baseNbtWorldData) SetName(value string) {
	d.compoundTag.SetString(tagLevelName, nbt.StringTag(value))
}

func (d *baseNbtWorldData) GetGenerator() string {
	return string(d.compoundTag.GetStringOr(tagGeneratorName, "DEFAULT"))
}

func (d *baseNbtWorldData) GetGeneratorOptions() string {
	return string(d.compoundTag.GetStringOr(tagGeneratorOptions, ""))
}

func (d *baseNbtWorldData) GetSeed() int64 {
	return int64(d.compoundTag.GetLongOr(tagRandomSeed, 0))
}

// GetTime is a port of BaseNbtWorldData::getTime: some older PocketMine-MP worlds saved it as a
// TAG_Int.
func (d *baseNbtWorldData) GetTime() int64 {
	if t, ok := d.compoundTag.GetTag(tagTime); ok {
		if v, isInt := t.(nbt.IntTag); isInt {
			return int64(v)
		}
	}
	return int64(d.compoundTag.GetLongOr(tagTime, 0))
}

func (d *baseNbtWorldData) SetTime(value int64) {
	d.compoundTag.SetLong(tagTime, nbt.LongTag(value))
}

func (d *baseNbtWorldData) GetSpawn() math.Vector3 {
	return math.NewVector3(
		float64(d.compoundTag.GetIntOr(tagSpawnX, 0)),
		float64(d.compoundTag.GetIntOr(tagSpawnY, 0)),
		float64(d.compoundTag.GetIntOr(tagSpawnZ, 0)),
	)
}

func (d *baseNbtWorldData) SetSpawn(pos math.Vector3) {
	d.compoundTag.SetInt(tagSpawnX, nbt.IntTag(pos.FloorX()))
	d.compoundTag.SetInt(tagSpawnY, nbt.IntTag(pos.FloorY()))
	d.compoundTag.SetInt(tagSpawnZ, nbt.IntTag(pos.FloorZ()))
}
