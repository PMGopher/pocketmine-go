package biomeselector

import (
	"testing"

	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/biome"
)

func fourWayLookup(temperature, rainfall float64) int {
	switch {
	case rainfall < 0.5 && temperature < 0.5:
		return biome.IDOcean
	case rainfall < 0.5:
		return biome.IDDesert
	case temperature < 0.5:
		return biome.IDForest
	default:
		return biome.IDPlains
	}
}

func TestPickBiomeIsDeterministicForAFixedSeed(t *testing.T) {
	registry := biome.NewRegistry()

	a := New(utils.NewRandom(7), registry, fourWayLookup)
	a.Recalculate()
	b := New(utils.NewRandom(7), registry, fourWayLookup)
	b.Recalculate()

	for _, pos := range [][2]float64{{0, 0}, {500, -200}, {-1234, 4321}} {
		got, want := a.PickBiome(pos[0], pos[1]), b.PickBiome(pos[0], pos[1])
		if got.ID() != want.ID() {
			t.Errorf("PickBiome(%v) = %s, want %s (same seed must give identical output)", pos, got.Name(), want.Name())
		}
	}
}

func TestPickBiomeOnlyReturnsBiomesTheLookupFuncCanProduce(t *testing.T) {
	registry := biome.NewRegistry()
	s := New(utils.NewRandom(1), registry, fourWayLookup)
	s.Recalculate()

	allowed := map[int]bool{biome.IDOcean: true, biome.IDDesert: true, biome.IDForest: true, biome.IDPlains: true}

	for x := -200.0; x <= 200; x += 17.3 {
		for z := -200.0; z <= 200; z += 23.1 {
			b := s.PickBiome(x, z)
			if !allowed[b.ID()] {
				t.Fatalf("PickBiome(%v,%v) = %s (id %d), want one of Ocean/Desert/Forest/Plains", x, z, b.Name(), b.ID())
			}
		}
	}
}

func TestGetTemperatureAndRainfallStayWithinUnitRange(t *testing.T) {
	s := New(utils.NewRandom(3), biome.NewRegistry(), fourWayLookup)

	for x := -100.0; x <= 100; x += 11 {
		for z := -100.0; z <= 100; z += 13 {
			temp := s.GetTemperature(x, z)
			rain := s.GetRainfall(x, z)
			if temp < 0 || temp > 1 {
				t.Fatalf("GetTemperature(%v,%v) = %v, want within [0,1]", x, z, temp)
			}
			if rain < 0 || rain > 1 {
				t.Fatalf("GetRainfall(%v,%v) = %v, want within [0,1]", x, z, rain)
			}
		}
	}
}
