package biome

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/world/generator/object"
	"pocketmine-go/pocketmine/world/generator/populator"
)

// grassyGroundCover is a port of GrassyBiome's constructor.
func grassyGroundCover() []block.Behavior {
	return []block.Behavior{block.VanillaGrass(), block.VanillaDirt(), block.VanillaDirt(), block.VanillaDirt(), block.VanillaDirt()}
}

// sandyGroundCover is a port of SandyBiome's constructor.
func sandyGroundCover() []block.Behavior {
	return []block.Behavior{block.VanillaSand(), block.VanillaSand(), block.VanillaSandstone(), block.VanillaSandstone(), block.VanillaSandstone()}
}

// snowyGroundCover is a port of SnowyBiome's constructor.
func snowyGroundCover() []block.Behavior {
	return []block.Behavior{block.VanillaSnowLayer(), block.VanillaGrass(), block.VanillaDirt(), block.VanillaDirt(), block.VanillaDirt()}
}

func tallGrass(baseAmount int) populator.Populator {
	tg := populator.NewTallGrass()
	tg.BaseAmount = baseAmount
	return tg
}

func trees(t object.TreeType, baseAmount int) populator.Populator {
	tr := populator.NewTree(t)
	tr.BaseAmount = baseAmount
	return tr
}

// NewOceanBiome is a port of OceanBiome.
func NewOceanBiome() *Biome {
	b := &Biome{name: "Ocean", groundCover: []block.Behavior{block.VanillaGravel(), block.VanillaGravel(), block.VanillaGravel(), block.VanillaGravel(), block.VanillaGravel()}}
	b.AddPopulator(tallGrass(5))
	b.SetElevation(46, 58)
	b.temperature, b.rainfall = 0.5, 0.5
	return b
}

// NewPlainBiome is a port of PlainBiome.
func NewPlainBiome() *Biome {
	b := &Biome{name: "Plains", groundCover: grassyGroundCover()}
	tg := populator.NewTallGrass()
	tg.BaseAmount = 12
	b.AddPopulator(tg)
	b.SetElevation(63, 68)
	b.temperature, b.rainfall = 0.8, 0.4
	return b
}

// NewDesertBiome is a port of DesertBiome (extends SandyBiome).
func NewDesertBiome() *Biome {
	b := &Biome{name: "Desert", groundCover: sandyGroundCover()}
	b.SetElevation(63, 74)
	b.temperature, b.rainfall = 2, 0
	return b
}

// NewMountainsBiome is a port of MountainsBiome (extends GrassyBiome).
func NewMountainsBiome() *Biome {
	b := &Biome{name: "Mountains", groundCover: grassyGroundCover()}
	b.AddPopulator(trees(object.TreeTypeOak, 1))
	b.AddPopulator(tallGrass(1))

	ore := &populator.Ore{}
	ore.SetOreTypes([]*object.OreType{
		object.NewOreType(block.VanillaEmeraldOre(), block.VanillaStone(), 11, 1, 0, 32),
	})
	b.AddPopulator(ore)

	b.SetElevation(63, 127)
	b.temperature, b.rainfall = 0.4, 0.5
	return b
}

// NewSmallMountainsBiome is a port of SmallMountainsBiome (extends MountainsBiome).
func NewSmallMountainsBiome() *Biome {
	b := NewMountainsBiome()
	b.name = "Small Mountains"
	b.SetElevation(63, 97)
	return b
}

// NewForestBiome is a port of ForestBiome (extends GrassyBiome). birch selects the Birch Forest
// variant (BiomeRegistry registers both a plain ForestBiome and a birch one at different IDs).
func NewForestBiome(birch bool) *Biome {
	name := "Oak Forest"
	temperature, rainfall := 0.7, 0.8
	treeType := object.TreeTypeOak
	if birch {
		name = "Birch Forest"
		temperature, rainfall = 0.6, 0.5
		treeType = object.TreeTypeBirch
	}

	b := &Biome{name: name, groundCover: grassyGroundCover()}
	b.AddPopulator(trees(treeType, 5))
	b.AddPopulator(tallGrass(3))
	b.SetElevation(63, 81)
	b.temperature, b.rainfall = temperature, rainfall
	return b
}

// NewTaigaBiome is a port of TaigaBiome (extends SnowyBiome).
func NewTaigaBiome() *Biome {
	b := &Biome{name: "Taiga", groundCover: snowyGroundCover()}
	b.AddPopulator(trees(object.TreeTypeSpruce, 10))
	b.AddPopulator(tallGrass(1))
	b.SetElevation(63, 81)
	b.temperature, b.rainfall = 0.05, 0.8
	return b
}

// NewSwampBiome is a port of SwampBiome (extends GrassyBiome).
func NewSwampBiome() *Biome {
	b := &Biome{name: "Swamp", groundCover: grassyGroundCover()}
	b.SetElevation(62, 63)
	b.temperature, b.rainfall = 0.8, 0.9
	return b
}

// NewRiverBiome is a port of RiverBiome.
func NewRiverBiome() *Biome {
	b := &Biome{name: "River", groundCover: []block.Behavior{block.VanillaDirt(), block.VanillaDirt(), block.VanillaDirt(), block.VanillaDirt(), block.VanillaDirt()}}
	b.AddPopulator(tallGrass(5))
	b.SetElevation(58, 62)
	b.temperature, b.rainfall = 0.5, 0.7
	return b
}

// NewHellBiome is a port of HellBiome.
func NewHellBiome() *Biome {
	return &Biome{name: "Hell"}
}

// NewIcePlainsBiome is a port of IcePlainsBiome (extends SnowyBiome).
func NewIcePlainsBiome() *Biome {
	b := &Biome{name: "Ice Plains", groundCover: snowyGroundCover()}
	b.AddPopulator(tallGrass(5))
	b.SetElevation(63, 74)
	b.temperature, b.rainfall = 0.05, 0.8
	return b
}

// NewUnknownBiome is a port of UnknownBiome: the polyfill BiomeRegistry.GetBiome auto-registers
// for any ID nothing else has claimed.
func NewUnknownBiome() *Biome {
	return &Biome{name: "Unknown"}
}
