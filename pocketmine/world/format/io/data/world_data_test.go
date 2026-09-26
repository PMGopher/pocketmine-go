package worlddata

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/format/io"
	"pocketmine-go/pocketmine/world/format/io/exception"
)

func TestBedrockWorldDataRoundTrip(t *testing.T) {
	dir := t.TempDir()
	spawn := math.NewVector3(12, 70, -34)
	if err := GenerateBedrockWorldData(dir, "My World", io.WorldCreationOptions{
		GeneratorName: "flat", GeneratorOptions: "2;7,3,3,2;1", Seed: 123456789, Difficulty: io.DifficultyEasy, SpawnPosition: spawn,
	}); err != nil {
		t.Fatal(err)
	}
	wd, err := LoadBedrockWorldData(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	if wd.GetName() != "My World" || wd.GetSeed() != 123456789 || wd.GetGenerator() != "flat" ||
		wd.GetGeneratorOptions() != "2;7,3,3,2;1" || wd.GetSpawn() != spawn || wd.GetDifficulty() != io.DifficultyEasy || wd.GetTime() != 0 {
		t.Fatalf("unexpected world data: %v", wd.GetCompoundTag())
	}
	if gen, _ := wd.GetCompoundTag().GetInt(tagGenerator); gen != GeneratorFlat {
		t.Errorf("Generator = %d, want flat", gen)
	}

	wd.SetTime(123456)
	wd.SetSpawn(math.NewVector3(5, 80, -5))
	wd.SetDifficulty(io.DifficultyHard)
	wd.SetRainLevel(0.75)
	wd.SetRainTime(500)
	wd.SetLightningLevel(0.5)
	wd.SetLightningTime(200)
	if err := wd.Save(); err != nil {
		t.Fatal(err)
	}
	reloaded, err := LoadBedrockWorldData(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.GetTime() != 123456 || reloaded.GetSpawn() != math.NewVector3(5, 80, -5) || reloaded.GetDifficulty() != io.DifficultyHard ||
		reloaded.GetRainLevel() != 0.75 || reloaded.GetRainTime() != 500 || reloaded.GetLightningLevel() != 0.5 || reloaded.GetLightningTime() != 200 {
		t.Fatalf("values didn't persist: %v", reloaded.GetCompoundTag())
	}
}

// writeLevelDat writes a level.dat with the given root compound, like vanilla does.
func writeLevelDat(t *testing.T, dir string, tag *nbt.CompoundTag) string {
	t.Helper()
	root, _ := nbt.NewTreeRoot(tag, "")
	buffer, err := nbt.NewLittleEndianSerializer().Write(root)
	if err != nil {
		t.Fatal(err)
	}
	data := append(binaryutils.WriteLInt(10), binaryutils.WriteLInt(int32(len(buffer)))...)
	path := filepath.Join(dir, "level.dat")
	if err := os.WriteFile(path, append(data, buffer...), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// A world saved by a 1.26.5x client (NetworkVersion 2193, no PocketMine-MP generatorName) loads
// as a "default" world; newer versions are rejected as unsupported.
func TestBedrockWorldDataVersions(t *testing.T) {
	vanilla := func(networkVersion int32) *nbt.CompoundTag {
		return nbt.NewCompoundTag().
			SetInt(tagStorageVersion, 10).
			SetInt(tagNetworkVersion, nbt.IntTag(networkVersion)).
			SetInt(tagGenerator, GeneratorInfinite).
			SetString(tagLevelName, "vanilla").
			SetLong(tagRandomSeed, -5857023233620755701).
			SetInt(tagSpawnX, 89).SetInt(tagSpawnY, 32767).SetInt(tagSpawnZ, -36)
	}

	wd, err := LoadBedrockWorldData(writeLevelDat(t, t.TempDir(), vanilla(2193)))
	if err != nil {
		t.Fatalf("1.26.50 world rejected: %v", err)
	}
	if wd.GetGenerator() != "default" || wd.GetGeneratorOptions() != "" || wd.GetName() != "vanilla" {
		t.Errorf("fix() gave generator %q options %q name %q", wd.GetGenerator(), wd.GetGeneratorOptions(), wd.GetName())
	}

	_, err = LoadBedrockWorldData(writeLevelDat(t, t.TempDir(), vanilla(CurrentStorageNetworkVersion+1)))
	var unsupported *exception.UnsupportedWorldFormatError
	if !errors.As(err, &unsupported) {
		t.Errorf("newer network version: got %v, want UnsupportedWorldFormatError", err)
	}

	limited := vanilla(2193).SetInt(tagGenerator, GeneratorLimited)
	if _, err := LoadBedrockWorldData(writeLevelDat(t, t.TempDir(), limited)); !errors.As(err, &unsupported) {
		t.Errorf("limited world: got %v, want UnsupportedWorldFormatError", err)
	}

	missing := vanilla(2193)
	missing.RemoveTag(tagStorageVersion)
	var corrupted *exception.CorruptedWorldError
	if _, err := LoadBedrockWorldData(writeLevelDat(t, t.TempDir(), missing)); !errors.As(err, &corrupted) {
		t.Errorf("missing StorageVersion: got %v, want CorruptedWorldError", err)
	}

	legacyClass := vanilla(2193).SetString(tagGeneratorName, `pocketmine\level\generator\Flat`)
	if wd, err := LoadBedrockWorldData(writeLevelDat(t, t.TempDir(), legacyClass)); err != nil || wd.GetGenerator() != "flat" {
		t.Errorf("generator class path fix: %v, %v", wd, err)
	}
}

func TestJavaWorldDataRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := GenerateJavaWorldData(dir, "Java", io.WorldCreationOptions{GeneratorName: "normal", Seed: 42, Difficulty: io.DifficultyHard, SpawnPosition: math.NewVector3(1, 2, 3)}, 19133); err != nil {
		t.Fatal(err)
	}
	wd, err := LoadJavaWorldData(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	if wd.GetName() != "Java" || wd.GetSeed() != 42 || wd.GetGenerator() != "normal" || wd.GetDifficulty() != io.DifficultyHard || wd.GetSpawn() != math.NewVector3(1, 2, 3) {
		t.Fatalf("unexpected: %v", wd.GetCompoundTag())
	}
	wd.SetRainLevel(0.2) // PC stores a raining flag: ceil(0.2) = 1
	if err := wd.Save(); err != nil {
		t.Fatal(err)
	}
	reloaded, err := LoadJavaWorldData(filepath.Join(dir, "level.dat"))
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.GetRainLevel() != 1 {
		t.Errorf("rain level = %v, want 1", reloaded.GetRainLevel())
	}
}
