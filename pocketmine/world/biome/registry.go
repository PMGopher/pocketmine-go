package biome

// Registry is a port of pocketmine\world\biome\BiomeRegistry. Unlike the PHP original (a
// SingletonTrait global), this is an explicit, constructed value - matching this port's
// established preference for passing dependencies in (World, BlockTranslator, ...) over reaching
// for package-level singletons, since only the not-yet-ported Normal generator ever needs one.
type Registry struct {
	biomes [MaxBiomes]*Biome
}

// NewRegistry is a port of BiomeRegistry::__construct: registers the same 11 concrete biomes at
// the same IDs as real PocketMine-MP.
func NewRegistry() *Registry {
	r := &Registry{}

	r.Register(IDOcean, NewOceanBiome())
	r.Register(IDPlains, NewPlainBiome())
	r.Register(IDDesert, NewDesertBiome())
	r.Register(IDExtremeHills, NewMountainsBiome())
	r.Register(IDForest, NewForestBiome(false))
	r.Register(IDTaiga, NewTaigaBiome())
	r.Register(IDSwampland, NewSwampBiome())
	r.Register(IDRiver, NewRiverBiome())

	r.Register(IDHell, NewHellBiome())

	r.Register(IDIcePlains, NewIcePlainsBiome())

	r.Register(IDExtremeHillsEdge, NewSmallMountainsBiome())

	r.Register(IDBirchForest, NewForestBiome(true))

	return r
}

// Register is a port of BiomeRegistry::register.
func (r *Registry) Register(id int, b *Biome) {
	r.biomes[id] = b
	b.SetID(id)
}

// GetBiome is a port of BiomeRegistry::getBiome: auto-registers UnknownBiome for any ID nothing
// else has claimed, matching the PHP original.
func (r *Registry) GetBiome(id int) *Biome {
	if id < 0 || id >= MaxBiomes {
		id = 0
	}
	if r.biomes[id] == nil {
		r.Register(id, NewUnknownBiome())
	}
	return r.biomes[id]
}
