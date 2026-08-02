// Package generator is a port of a slice of pocketmine\world\generator.
package generator

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/world/format"
)

// Generator is a port of pocketmine\world\generator\Generator's core contract: GenerateChunk and
// PopulateChunk. ConvertSeed (parsing a string world seed into an int) isn't ported - nothing
// here takes a user-supplied string seed yet.
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
