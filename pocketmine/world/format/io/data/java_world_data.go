package worlddata

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	stdio "io"
	"math"
	"os"
	"path/filepath"
	"time"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/format/io"
	"pocketmine-go/pocketmine/world/format/io/exception"
)

// JavaWorldData tag names, a port of JavaWorldData::TAG_*.
const (
	javaTagDayTime          = "DayTime"
	javaTagDifficulty       = "Difficulty"
	javaTagFormatVersion    = "version"
	javaTagGameRules        = "GameRules"
	javaTagGameType         = "GameType"
	javaTagGeneratorVersion = "generatorVersion"
	javaTagHardcore         = "hardcore"
	javaTagInitialized      = "initialized"
	javaTagLastPlayed       = "LastPlayed"
	javaTagRaining          = "raining"
	javaTagRainTime         = "rainTime"
	javaTagRootData         = "Data"
	javaTagSizeOnDisk       = "SizeOnDisk"
	javaTagThundering       = "thundering"
	javaTagThunderTime      = "thunderTime"
)

// JavaWorldData is a port of pocketmine\world\format\io\data\JavaWorldData: a region (Anvil,
// McRegion, PMAnvil) world's level.dat (gzip-compressed big-endian NBT).
type JavaWorldData struct {
	baseNbtWorldData
}

var _ io.WorldData = (*JavaWorldData)(nil)

// ZlibDecode is PHP's zlib_decode: it accepts gzip and zlib data.
func ZlibDecode(data []byte) ([]byte, error) {
	var r stdio.ReadCloser
	var err error
	if len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b {
		r, err = gzip.NewReader(bytes.NewReader(data))
	} else {
		r, err = zlib.NewReader(bytes.NewReader(data))
	}
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return stdio.ReadAll(r)
}

func gzipEncode(data []byte) []byte {
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, _ = w.Write(data)
	_ = w.Close()
	return buf.Bytes()
}

func encodeJavaLevelDat(tag *nbt.CompoundTag) ([]byte, error) {
	root, err := nbt.NewTreeRoot(nbt.NewCompoundTag().SetTag(javaTagRootData, tag), "")
	if err != nil {
		return nil, err
	}
	buffer, err := nbt.NewBigEndianSerializer().Write(root)
	if err != nil {
		return nil, err
	}
	return gzipEncode(buffer), nil
}

// GenerateJavaWorldData is a port of JavaWorldData::generate (version is the PC world format
// version, 19133 for Anvil).
func GenerateJavaWorldData(path, name string, options io.WorldCreationOptions, version int) error {
	//TODO, add extra details
	worldData := nbt.NewCompoundTag().
		SetByte(javaTagHardcore, 0).
		SetByte(javaTagDifficulty, nbt.ByteTag(options.Difficulty)).
		SetByte(javaTagInitialized, 1).
		SetInt(javaTagGameType, 0).
		SetInt(javaTagGeneratorVersion, 1). //2 in MCPE
		SetInt(tagSpawnX, nbt.IntTag(options.SpawnPosition.FloorX())).
		SetInt(tagSpawnY, nbt.IntTag(options.SpawnPosition.FloorY())).
		SetInt(tagSpawnZ, nbt.IntTag(options.SpawnPosition.FloorZ())).
		SetInt(javaTagFormatVersion, nbt.IntTag(version)).
		SetInt(javaTagDayTime, 0).
		SetLong(javaTagLastPlayed, nbt.LongTag(time.Now().UnixMilli())).
		SetLong(tagRandomSeed, nbt.LongTag(options.Seed)).
		SetLong(javaTagSizeOnDisk, 0).
		SetLong(tagTime, 0).
		SetString(tagGeneratorName, nbt.StringTag(options.GeneratorName)).
		SetString(tagGeneratorOptions, nbt.StringTag(options.GeneratorOptions)).
		SetString(tagLevelName, nbt.StringTag(name)).
		SetTag(javaTagGameRules, nbt.NewCompoundTag())

	data, err := encodeJavaLevelDat(worldData)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, "level.dat"), data, 0o644)
}

// LoadJavaWorldData is a port of `new JavaWorldData($dataPath)`.
func LoadJavaWorldData(dataPath string) (*JavaWorldData, error) {
	d := &JavaWorldData{}
	base, err := newBaseNbtWorldData(dataPath, func() (*nbt.CompoundTag, error) { return loadJavaLevelDat(dataPath) }, fixJavaLevelDat)
	if err != nil {
		return nil, err
	}
	d.baseNbtWorldData = base
	return d, nil
}

// loadJavaLevelDat is a port of JavaWorldData::load.
func loadJavaLevelDat(dataPath string) (*nbt.CompoundTag, error) {
	raw, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, &exception.CorruptedWorldError{Message: err.Error(), Cause: err}
	}
	decompressed, err := ZlibDecode(raw)
	if err != nil {
		return nil, exception.NewCorruptedWorldError("Failed to decompress level.dat contents")
	}
	root, _, err := nbt.NewBigEndianSerializer().Read(decompressed, 0, nbtMaxDepth)
	if err != nil {
		return nil, &exception.CorruptedWorldError{Message: err.Error(), Cause: err}
	}
	worldData, err := root.MustGetCompoundTag()
	if err != nil {
		return nil, &exception.CorruptedWorldError{Message: err.Error(), Cause: err}
	}
	dataTag, _ := worldData.GetTag(javaTagRootData)
	compound, ok := dataTag.(*nbt.CompoundTag)
	if !ok {
		return nil, exception.NewCorruptedWorldError("Missing '%s' key or wrong type", javaTagRootData)
	}
	return compound, nil
}

// fixJavaLevelDat is a port of JavaWorldData::fix.
func fixJavaLevelDat(tag *nbt.CompoundTag) error {
	generatorNameTag, _ := tag.GetTag(tagGeneratorName)
	if generatorName, ok := generatorNameTag.(nbt.StringTag); !ok {
		tag.SetString(tagGeneratorName, "default")
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

// Save is a port of JavaWorldData::save.
func (d *JavaWorldData) Save() error {
	d.compoundTag.SetLong(pocketmine.TagWorldDataVersion, nbt.LongTag(pocketmine.WorldDataVersion))
	data, err := encodeJavaLevelDat(d.compoundTag)
	if err != nil {
		return err
	}
	return utils.SafeFilePutContents(d.dataPath, data)
}

func (d *JavaWorldData) GetDifficulty() int {
	return int(d.compoundTag.GetByteOr(javaTagDifficulty, io.DifficultyNormal))
}

func (d *JavaWorldData) SetDifficulty(difficulty int) {
	d.compoundTag.SetByte(javaTagDifficulty, nbt.ByteTag(difficulty))
}

func (d *JavaWorldData) GetRainTime() int { return int(d.compoundTag.GetIntOr(javaTagRainTime, 0)) }

func (d *JavaWorldData) SetRainTime(ticks int) {
	d.compoundTag.SetInt(javaTagRainTime, nbt.IntTag(ticks))
}

func (d *JavaWorldData) GetRainLevel() float64 {
	return float64(d.compoundTag.GetByteOr(javaTagRaining, 0))
}

func (d *JavaWorldData) SetRainLevel(level float64) {
	d.compoundTag.SetByte(javaTagRaining, nbt.ByteTag(int(math.Ceil(level))))
}

func (d *JavaWorldData) GetLightningTime() int {
	return int(d.compoundTag.GetIntOr(javaTagThunderTime, 0))
}

func (d *JavaWorldData) SetLightningTime(ticks int) {
	d.compoundTag.SetInt(javaTagThunderTime, nbt.IntTag(ticks))
}

func (d *JavaWorldData) GetLightningLevel() float64 {
	return float64(d.compoundTag.GetByteOr(javaTagThundering, 0))
}

func (d *JavaWorldData) SetLightningLevel(level float64) {
	d.compoundTag.SetByte(javaTagThundering, nbt.ByteTag(int(math.Ceil(level))))
}
