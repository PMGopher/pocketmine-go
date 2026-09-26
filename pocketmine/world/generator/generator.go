// Package generator is a port of a slice of pocketmine\world\generator.
package generator

import (
	stdmath "math"
	"regexp"
	"strconv"
	"strings"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/format"
)

// Generator is a port of pocketmine\world\generator\Generator's core contract: GenerateChunk and
// PopulateChunk (ConvertSeed is below).
type Generator interface {
	// GenerateChunk is a port of Generator::generateChunk, minus the ChunkManager write: PHP calls
	// `$world->setChunk($chunkX, $chunkZ, ...)` itself; this port doesn't have a ChunkManager/World
	// type calling into it yet, so it returns the generated chunk and lets the caller (World, once
	// built) store it instead.
	GenerateChunk(chunkX, chunkZ int) *format.Chunk
	// PopulateChunk is a port of Generator::populateChunk - runs this generator's Populators
	// (e.g. Flat's optional Ore decoration) against an already-generated, already-stored chunk, so
	// they can read/write neighbouring blocks through world (a populator placing an ore vein at the
	// edge of a chunk needs its neighbour to already be reachable via World.GetBlockAt/SetBlock).
	PopulateChunk(world block.World, chunkX, chunkZ int)
}

// seedIntegerPattern is convertSeed's /^-?\d+$/: this avoids treating seeds like "404.4" as
// integer seeds.
var seedIntegerPattern = regexp.MustCompile(`^-?\d+$`)

// ConvertSeed is a port of Generator::convertSeed: a seed from configuration as a number. ok is
// false for an empty seed, which should cause a random seed to be selected - can't use 0 here
// because 0 is a valid seed.
func ConvertSeed(seed string) (converted int64, ok bool) {
	if seed == "" {
		return 0, false
	}
	if seedIntegerPattern.MatchString(seed) {
		// (int) "99999999999999999999" saturates to PHP_INT_MAX; ParseInt does the same.
		n, err := strconv.ParseInt(seed, 10, 64)
		if err != nil {
			if strings.HasPrefix(seed, "-") {
				return stdmath.MinInt64, true
			}
			return stdmath.MaxInt64, true
		}
		return n, true
	}
	return int64(utils.JavaStringHash(seed)), true
}
