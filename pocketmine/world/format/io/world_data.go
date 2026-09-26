// Package io is a port of pocketmine\world\format\io: world providers (LevelDB, and read-only
// region formats in io/region), the WorldData interface for a world's level.dat (implemented in
// io/data), the global block/item data handlers, the world format converter and the provider
// manager.
package io

import "pocketmine-go/pocketmine/math"

// WorldData is a port of pocketmine\world\format\io\WorldData: a world's metadata (level.dat).
type WorldData interface {
	// Save is a port of WorldData::save: writes the world state (weather, time, ...) to disk.
	Save() error

	GetName() string
	SetName(value string)
	// GetGenerator returns the generator name.
	GetGenerator() string
	GetGeneratorOptions() string
	GetSeed() int64
	GetTime() int64
	SetTime(value int64)
	GetSpawn() math.Vector3
	SetSpawn(pos math.Vector3)
	// GetDifficulty returns the world difficulty (one of the World::DIFFICULTY_* constants).
	GetDifficulty() int
	SetDifficulty(difficulty int)
	// GetRainTime returns the time in ticks to the next rain level change.
	GetRainTime() int
	SetRainTime(ticks int)
	// GetRainLevel returns 0.0 - 1.0.
	GetRainLevel() float64
	SetRainLevel(level float64)
	// GetLightningTime returns the time in ticks to the next lightning level change.
	GetLightningTime() int
	SetLightningTime(ticks int)
	// GetLightningLevel returns 0.0 - 1.0.
	GetLightningLevel() float64
	SetLightningLevel(level float64)
}

// WorldCreationOptions is what a provider's generate() reads from
// pocketmine\world\WorldCreationOptions (the world package's type can't be imported here: world
// imports this package). GeneratorName is GeneratorManager::getGeneratorName of the options'
// generator class.
type WorldCreationOptions struct {
	GeneratorName    string
	GeneratorOptions string
	Seed             int64
	Difficulty       int
	SpawnPosition    math.Vector3
}

// Difficulty constants, a port of World::DIFFICULTY_* (the world package re-exports them; they're
// needed here for level.dat's default difficulty).
const (
	DifficultyPeaceful = 0
	DifficultyEasy     = 1
	DifficultyNormal   = 2
	DifficultyHard     = 3
)
