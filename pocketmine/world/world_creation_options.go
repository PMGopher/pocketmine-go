package world

import (
	"math/rand"

	"pocketmine-go/pocketmine/math"
	worldio "pocketmine-go/pocketmine/world/format/io"
)

// WorldCreationOptions is a port of pocketmine\world\WorldCreationOptions: user-customizable
// settings for world creation.
//
// GeneratorClass isn't ported (real PHP's class-string<Generator>, validated via
// Utils::testValidInstance and later resolved through GeneratorManager) - WorldManager.GenerateWorld
// already takes an already-constructed generator.Generator directly instead (see its own doc
// comment on why: not every Generator this port has can be built from a bare name/options string
// yet). GeneratorName is kept here as the string recorded into level.dat, so a later LoadWorld can
// resolve an equivalent generator via generator.GetFactory - the same string GenerateWorld already
// needed, just carried on this options value now instead of as a bare parameter.
type WorldCreationOptions struct {
	GeneratorName    string
	Seed             int64
	Difficulty       int
	GeneratorOptions string
	SpawnPosition    math.Vector3
}

// NewWorldCreationOptions is a port of WorldCreationOptions::create()/__construct: a random
// int32-range seed, spawn (256,70,256), NORMAL difficulty, no generator options, "normal" generator
// - all exactly matching the real constructor's own defaults.
func NewWorldCreationOptions() *WorldCreationOptions {
	return &WorldCreationOptions{
		GeneratorName: "normal",
		Seed:          int64(int32(rand.Uint32())), // random_int(Limits::INT32_MIN, Limits::INT32_MAX)
		Difficulty:    worldio.DifficultyNormal,
		SpawnPosition: math.NewVector3(256, 70, 256),
	}
}

// SetGeneratorName is a port of WorldCreationOptions::setGeneratorClass, adapted to this port's own
// name-based generator resolution (see WorldCreationOptions' own doc comment).
func (o *WorldCreationOptions) SetGeneratorName(name string) *WorldCreationOptions {
	o.GeneratorName = name
	return o
}

// SetSeed is a port of WorldCreationOptions::setSeed.
func (o *WorldCreationOptions) SetSeed(seed int64) *WorldCreationOptions {
	o.Seed = seed
	return o
}

// SetDifficulty is a port of WorldCreationOptions::setDifficulty.
func (o *WorldCreationOptions) SetDifficulty(difficulty int) *WorldCreationOptions {
	o.Difficulty = difficulty
	return o
}

// SetGeneratorOptions is a port of WorldCreationOptions::setGeneratorOptions.
func (o *WorldCreationOptions) SetGeneratorOptions(options string) *WorldCreationOptions {
	o.GeneratorOptions = options
	return o
}

// SetSpawnPosition is a port of WorldCreationOptions::setSpawnPosition.
func (o *WorldCreationOptions) SetSpawnPosition(pos math.Vector3) *WorldCreationOptions {
	o.SpawnPosition = pos
	return o
}
