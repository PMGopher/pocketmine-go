package world

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestNewWorldCreationOptionsMatchesRealDefaults(t *testing.T) {
	o := NewWorldCreationOptions()

	if o.GeneratorName != "normal" {
		t.Errorf("GeneratorName = %q, want %q", o.GeneratorName, "normal")
	}
	if o.Difficulty != 2 { // DifficultyNormal
		t.Errorf("Difficulty = %d, want DifficultyNormal (2)", o.Difficulty)
	}
	if o.GeneratorOptions != "" {
		t.Errorf("GeneratorOptions = %q, want empty", o.GeneratorOptions)
	}
	if want := math.NewVector3(256, 70, 256); o.SpawnPosition != want {
		t.Errorf("SpawnPosition = %v, want %v", o.SpawnPosition, want)
	}
}

func TestNewWorldCreationOptionsPicksADifferentSeedEachTime(t *testing.T) {
	a := NewWorldCreationOptions()
	b := NewWorldCreationOptions()
	if a.Seed == b.Seed {
		t.Errorf("two calls to NewWorldCreationOptions produced the same seed (%d) - want randomized", a.Seed)
	}
}

func TestWorldCreationOptionsSettersChainAndApply(t *testing.T) {
	spawn := math.NewVector3(1, 2, 3)
	o := NewWorldCreationOptions().
		SetGeneratorName("hell").
		SetSeed(42).
		SetDifficulty(3).
		SetGeneratorOptions("some-options").
		SetSpawnPosition(spawn)

	if o.GeneratorName != "hell" {
		t.Errorf("GeneratorName = %q, want %q", o.GeneratorName, "hell")
	}
	if o.Seed != 42 {
		t.Errorf("Seed = %d, want 42", o.Seed)
	}
	if o.Difficulty != 3 {
		t.Errorf("Difficulty = %d, want 3", o.Difficulty)
	}
	if o.GeneratorOptions != "some-options" {
		t.Errorf("GeneratorOptions = %q, want %q", o.GeneratorOptions, "some-options")
	}
	if o.SpawnPosition != spawn {
		t.Errorf("SpawnPosition = %v, want %v", o.SpawnPosition, spawn)
	}
}
