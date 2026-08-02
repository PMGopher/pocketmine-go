// Package generator is a port of a slice of pocketmine\world\generator.
package generator

import "pocketmine-go/pocketmine/world/format"

// Generator is a port of pocketmine\world\generator\Generator's core contract: GenerateChunk.
// PopulateChunk (and the whole Populator system it drives - Ore veins, trees, ground cover, tall
// grass) isn't ported: it's a separate, larger undertaking (each populator is its own
// world-generation algorithm) unrelated to getting a base chunk generated at all, which is what
// this port needs first. ConvertSeed (parsing a string world seed into an int) isn't ported either
// - nothing here takes a user-supplied string seed yet.
type Generator interface {
	// GenerateChunk is a port of Generator::generateChunk, minus the ChunkManager write: PHP calls
	// `$world->setChunk($chunkX, $chunkZ, ...)` itself; this port doesn't have a ChunkManager/World
	// type calling into it yet, so it returns the generated chunk and lets the caller (World, once
	// built) store it instead.
	GenerateChunk(chunkX, chunkZ int) *format.Chunk
}
