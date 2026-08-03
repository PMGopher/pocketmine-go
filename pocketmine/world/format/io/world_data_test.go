package io

import (
	"os"
	"path/filepath"
	"testing"

	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

func TestGenerateThenLoadWorldDataRoundTripsEverySetField(t *testing.T) {
	dir := t.TempDir()
	spawn := math.NewVector3(12, 70, -34)

	generated, err := GenerateWorldData(dir, "My World", 123456789, GeneratorFlat, "flat", "2;7,3,3,2;1", spawn)
	if err != nil {
		t.Fatalf("GenerateWorldData: %v", err)
	}

	if got := generated.GetName(); got != "My World" {
		t.Errorf("GetName() = %q, want %q", got, "My World")
	}
	if got := generated.GetSeed(); got != 123456789 {
		t.Errorf("GetSeed() = %d, want 123456789", got)
	}
	if got := generated.GetGenerator(); got != "flat" {
		t.Errorf("GetGenerator() = %q, want %q", got, "flat")
	}
	if got := generated.GetGeneratorOptions(); got != "2;7,3,3,2;1" {
		t.Errorf("GetGeneratorOptions() = %q, want %q", got, "2;7,3,3,2;1")
	}
	if got := generated.GetSpawn(); got != spawn {
		t.Errorf("GetSpawn() = %v, want %v", got, spawn)
	}
	if got := generated.GetTime(); got != 0 {
		t.Errorf("GetTime() on a freshly generated world = %d, want 0", got)
	}

	loaded, err := LoadWorldData(dir)
	if err != nil {
		t.Fatalf("LoadWorldData: %v", err)
	}
	if got := loaded.GetName(); got != "My World" {
		t.Errorf("after reload, GetName() = %q, want %q", got, "My World")
	}
	if got := loaded.GetSeed(); got != 123456789 {
		t.Errorf("after reload, GetSeed() = %d, want 123456789", got)
	}
	if got := loaded.GetGenerator(); got != "flat" {
		t.Errorf("after reload, GetGenerator() = %q, want %q", got, "flat")
	}
	if got := loaded.GetSpawn(); got != spawn {
		t.Errorf("after reload, GetSpawn() = %v, want %v", got, spawn)
	}
}

func TestWorldDataSetTimeSpawnAndWeatherPersistAcrossSave(t *testing.T) {
	dir := t.TempDir()
	wd, err := GenerateWorldData(dir, "W", 1, GeneratorInfinite, "default", "", math.NewVector3(0, 64, 0))
	if err != nil {
		t.Fatal(err)
	}

	wd.SetTime(123456)
	wd.SetSpawn(math.NewVector3(5, 80, -5))
	wd.SetDifficulty(DifficultyHard)
	wd.SetRainLevel(0.75)
	wd.SetRainTime(500)
	wd.SetLightningLevel(0.5)
	wd.SetLightningTime(200)

	if err := wd.Save(dir); err != nil {
		t.Fatalf("Save: %v", err)
	}

	reloaded, err := LoadWorldData(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.GetTime(); got != 123456 {
		t.Errorf("GetTime() after reload = %d, want 123456", got)
	}
	if got := reloaded.GetSpawn(); got != math.NewVector3(5, 80, -5) {
		t.Errorf("GetSpawn() after reload = %v, want (5,80,-5)", got)
	}
	if got := reloaded.GetDifficulty(); got != DifficultyHard {
		t.Errorf("GetDifficulty() after reload = %d, want %d", got, DifficultyHard)
	}
	if got := reloaded.GetRainLevel(); got != 0.75 {
		t.Errorf("GetRainLevel() after reload = %v, want 0.75", got)
	}
	if got := reloaded.GetRainTime(); got != 500 {
		t.Errorf("GetRainTime() after reload = %d, want 500", got)
	}
	if got := reloaded.GetLightningLevel(); got != 0.5 {
		t.Errorf("GetLightningLevel() after reload = %v, want 0.5", got)
	}
	if got := reloaded.GetLightningTime(); got != 200 {
		t.Errorf("GetLightningTime() after reload = %d, want 200", got)
	}
}

func TestWorldDataDefaultsToNormalDifficultyWhenUnset(t *testing.T) {
	dir := t.TempDir()
	wd, err := GenerateWorldData(dir, "W", 1, GeneratorInfinite, "default", "", math.NewVector3(0, 64, 0))
	if err != nil {
		t.Fatal(err)
	}
	if got := wd.GetDifficulty(); got != DifficultyNormal {
		t.Errorf("GetDifficulty() on a freshly generated world = %d, want DifficultyNormal (%d)", got, DifficultyNormal)
	}
}

func TestLoadWorldDataRejectsATooNewStorageVersion(t *testing.T) {
	dir := t.TempDir()

	// WorldData.Save always re-stamps the current storage version, so an actually-too-new
	// level.dat (as a real newer client/server would leave behind) can only be exercised by
	// writing the raw bytes directly, bypassing Save entirely.
	tag := nbt.NewCompoundTag().
		SetInt(tagStorageVersion, nbt.IntTag(currentStorageVersion+1)).
		SetInt(tagNetworkVersion, currentStorageNetworkVersion).
		SetString(tagLevelName, "W").
		SetLong(tagRandomSeed, 1)
	root, err := nbt.NewTreeRoot(tag, "")
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := nbt.NewLittleEndianSerializer().Write(root)
	if err != nil {
		t.Fatal(err)
	}
	header := append(binaryutils.WriteLInt(int32(currentStorageVersion+1)), binaryutils.WriteLInt(int32(len(buffer)))...)
	if err := os.WriteFile(filepath.Join(dir, levelDatFileName), append(header, buffer...), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadWorldData(dir); err == nil {
		t.Error("LoadWorldData with a too-new StorageVersion = nil error, want an error")
	}
}

func TestWorldDataFixFillsInGeneratorNameFromLegacyGeneratorTagForFlat(t *testing.T) {
	dir := t.TempDir()
	tag := nbt.NewCompoundTag().
		SetInt(tagStorageVersion, currentStorageVersion).
		SetInt(tagNetworkVersion, currentStorageNetworkVersion).
		SetString(tagLevelName, "Legacy").
		SetLong(tagRandomSeed, 42).
		SetInt(tagGenerator, GeneratorFlat)
	wd := &WorldData{tag: tag}
	if err := wd.Save(dir); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadWorldData(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.GetGenerator(); got != "flat" {
		t.Errorf("GetGenerator() after fix() on a legacy Generator=FLAT tag = %q, want %q", got, "flat")
	}
	if got := loaded.GetGeneratorOptions(); got != "2;7,3,3,2;1" {
		t.Errorf("GetGeneratorOptions() after fix() = %q, want %q", got, "2;7,3,3,2;1")
	}
}
