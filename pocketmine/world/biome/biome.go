// Package biome is a port of a slice of pocketmine\world\biome: per-biome elevation, ground
// cover, climate and decoration, keyed by the Bedrock legacy biome ID scheme.
package biome

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/generator/populator"
)

// MaxBiomes is a port of Biome::MAX_BIOMES.
const MaxBiomes = 256

// Biome is a port of pocketmine\world\biome\Biome. PHP models each concrete biome (PlainBiome,
// OceanBiome, ...) as its own subclass whose constructor configures a shared base - since that
// inheritance exists purely to share construction logic, not polymorphic behaviour, this port
// uses one concrete struct plus plain constructor functions instead (see plains.go, ocean.go, ...
// for the "subclasses"), matching this codebase's established preference for composition/helpers
// over deep type hierarchies where nothing dispatches virtually.
type Biome struct {
	id         int
	registered bool
	name       string

	populators []populator.Populator

	minElevation, maxElevation int

	groundCover []block.Behavior

	rainfall    float64
	temperature float64
}

// SetID is a port of Biome::setId (registration is one-shot, matching the PHP original).
func (b *Biome) SetID(id int) {
	if !b.registered {
		b.registered = true
		b.id = id
	}
}

func (b *Biome) ID() int { return b.id }

func (b *Biome) Name() string { return b.name }

func (b *Biome) MinElevation() int { return b.minElevation }
func (b *Biome) MaxElevation() int { return b.maxElevation }

// SetElevation is a port of Biome::setElevation.
func (b *Biome) SetElevation(min, max int) {
	b.minElevation, b.maxElevation = min, max
}

// GroundCover is a port of Biome::getGroundCover: the Y-descending sequence of blocks a
// GroundCover populator lays over this biome's raw terrain (e.g. grass, then 4 layers of dirt).
func (b *Biome) GroundCover() []block.Behavior { return b.groundCover }

func (b *Biome) SetGroundCover(cover []block.Behavior) { b.groundCover = cover }

func (b *Biome) Temperature() float64 { return b.temperature }
func (b *Biome) Rainfall() float64    { return b.rainfall }

// AddPopulator is a port of Biome::addPopulator.
func (b *Biome) AddPopulator(p populator.Populator) {
	b.populators = append(b.populators, p)
}

func (b *Biome) Populators() []populator.Populator { return b.populators }

// PopulateChunk is a port of Biome::populateChunk.
func (b *Biome) PopulateChunk(world block.World, chunkX, chunkZ int, random *utils.Random) {
	for _, p := range b.populators {
		p.Populate(world, chunkX, chunkZ, random)
	}
}
