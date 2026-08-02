// Package biomeselector is a port of pocketmine\world\generator\biome (just BiomeSelector) -
// named differently from pocketmine/world/biome (the Biome/Registry types it selects between) so
// callers that need both don't have to alias one at every import site, unlike PHP's fully
// qualified namespaces.
package biomeselector

import (
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/biome"
	"pocketmine-go/pocketmine/world/generator/noise"
)

const mapSize = 64

// LookupFunc is a port of BiomeSelector::lookup. PHP makes this an abstract method implemented by
// an anonymous subclass at each call site (Normal's constructor is the only caller); Go has no
// anonymous subclassing, so this port takes the lookup table as a plain function value instead -
// the same "function value instead of a one-off abstract subtype" pattern already used elsewhere
// in this port (e.g. block.NewItemBlockFunc).
type LookupFunc func(temperature, rainfall float64) int

// Selector is a port of pocketmine\world\generator\biome\BiomeSelector.
type Selector struct {
	temperature *noise.Simplex
	rainfall    *noise.Simplex

	lookup   LookupFunc
	registry *biome.Registry

	table [mapSize * mapSize]*biome.Biome
}

// New is a port of BiomeSelector::__construct. Call Recalculate once lookup is ready to use -
// mirroring the PHP original's separate construct-then-recalculate() two-step (Normal.php calls
// recalculate() itself right after constructing its anonymous subclass).
func New(random *utils.Random, registry *biome.Registry, lookup LookupFunc) *Selector {
	return &Selector{
		temperature: noise.NewSimplex(random, 2, 1.0/16, 1.0/512),
		rainfall:    noise.NewSimplex(random, 2, 1.0/16, 1.0/512),
		lookup:      lookup,
		registry:    registry,
	}
}

// Recalculate is a port of BiomeSelector::recalculate.
func (s *Selector) Recalculate() {
	for i := 0; i < mapSize; i++ {
		for j := 0; j < mapSize; j++ {
			id := s.lookup(float64(i)/63, float64(j)/63)
			s.table[i+(j<<6)] = s.registry.GetBiome(id)
		}
	}
}

// GetTemperature is a port of BiomeSelector::getTemperature.
func (s *Selector) GetTemperature(x, z float64) float64 {
	return (s.temperature.Noise2D(x, z, true) + 1) / 2
}

// GetRainfall is a port of BiomeSelector::getRainfall.
func (s *Selector) GetRainfall(x, z float64) float64 {
	return (s.rainfall.Noise2D(x, z, true) + 1) / 2
}

// PickBiome is a port of BiomeSelector::pickBiome.
func (s *Selector) PickBiome(x, z float64) *biome.Biome {
	temperature := int(s.GetTemperature(x, z) * 63)
	rainfall := int(s.GetRainfall(x, z) * 63)
	return s.table[temperature+(rainfall<<6)]
}
