package tile

import (
	"fmt"
	"sync"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// Constructor builds a tile at a position, before its save data is read (PHP's
// `new $class($world, $pos)`).
type Constructor func(world World, pos math.Vector3) Tile

// TileFactory is a port of pocketmine\block\tile\TileFactory: the save IDs of every tile type, for
// loading tiles from chunk NBT. Go has no class names, so a tile's own save ID comes from its
// SaveID() (the first name PHP registers for its class); the registry maps every accepted save ID
// (including legacy aliases) to a constructor.
type TileFactory struct {
	knownTiles map[string]Constructor
}

var (
	tileFactoryOnce sync.Once
	tileFactory     *TileFactory
)

// GetTileFactory is TileFactory::getInstance().
func GetTileFactory() *TileFactory {
	tileFactoryOnce.Do(func() { tileFactory = newTileFactory() })
	return tileFactory
}

func ctor[T Tile](f func(World, math.Vector3) T) Constructor {
	return func(world World, pos math.Vector3) Tile { return f(world, pos) }
}

// newTileFactory is a port of TileFactory::__construct. PHP also registers each class's short
// name, which is the same as one of the listed names for every tile except those added below.
func newTileFactory() *TileFactory {
	f := &TileFactory{knownTiles: map[string]Constructor{}}
	f.Register(ctor(NewBarrel), "Barrel", "minecraft:barrel")
	f.Register(ctor(NewBanner), "Banner", "minecraft:banner")
	f.Register(ctor(NewBeacon), "Beacon", "minecraft:beacon")
	f.Register(ctor(NewBed), "Bed", "minecraft:bed")
	f.Register(ctor(NewBell), "Bell", "minecraft:bell")
	f.Register(ctor(NewBlastFurnace), "BlastFurnace", "minecraft:blast_furnace")
	f.Register(ctor(NewBrewingStand), "BrewingStand", "minecraft:brewing_stand")
	f.Register(ctor(NewCampfire), "Campfire", "minecraft:campfire")
	f.Register(ctor(NewCauldron), "Cauldron", "minecraft:cauldron")
	f.Register(ctor(NewChest), "Chest", "minecraft:chest")
	f.Register(ctor(NewChiseledBookshelf), "ChiseledBookshelf", "minecraft:chiseled_bookshelf")
	f.Register(ctor(NewComparator), "Comparator", "minecraft:comparator")
	f.Register(ctor(NewDaylightSensor), "DaylightDetector", "minecraft:daylight_detector", "DaylightSensor")
	f.Register(ctor(NewEnchantTable), "EnchantTable", "minecraft:enchanting_table")
	f.Register(ctor(NewEnderChest), "EnderChest", "minecraft:ender_chest")
	f.Register(ctor(NewFlowerPot), "FlowerPot", "minecraft:flower_pot")
	f.Register(ctor(NewNormalFurnace), "Furnace", "minecraft:furnace", "NormalFurnace")
	f.Register(ctor(NewHopper), "Hopper", "minecraft:hopper")
	f.Register(ctor(NewItemFrame), "ItemFrame") // this is an entity in PC
	f.Register(ctor(NewJukebox), "Jukebox", "RecordPlayer", "minecraft:jukebox")
	f.Register(ctor(NewLectern), "Lectern", "minecraft:lectern")
	f.Register(ctor(NewMonsterSpawner), "MobSpawner", "minecraft:mob_spawner", "MonsterSpawner")
	f.Register(ctor(NewNote), "Music", "minecraft:noteblock", "Note")
	f.Register(ctor(NewShulkerBox), "ShulkerBox", "minecraft:shulker_box")
	f.Register(ctor(NewSign), "Sign", "minecraft:sign")
	f.Register(ctor(NewSmoker), "Smoker", "minecraft:smoker")
	f.Register(ctor(NewSporeBlossom), "SporeBlossom", "minecraft:spore_blossom")
	f.Register(ctor(NewMobHead), "Skull", "minecraft:skull", "MobHead")
	f.Register(ctor(NewGlowingItemFrame), "GlowItemFrame", "GlowingItemFrame")
	f.Register(ctor(NewHangingSign), "HangingSign", "minecraft:hanging_sign")

	//TODO: ChalkboardBlock
	//TODO: ChemistryTable
	//TODO: CommandBlock
	//TODO: Conduit
	//TODO: Dispenser
	//TODO: Dropper
	//TODO: EndGateway
	//TODO: EndPortal
	//TODO: JigsawBlock
	//TODO: MovingBlock
	//TODO: NetherReactor
	//TODO: PistonArm
	//TODO: StructureBlock
	return f
}

// Register is a port of TileFactory::register: every save name maps to the constructor.
func (f *TileFactory) Register(constructor Constructor, saveNames ...string) {
	for _, name := range saveNames {
		f.knownTiles[name] = constructor
	}
}

// IsRegistered reports whether a save ID is known.
func (f *TileFactory) IsRegistered(saveID string) bool {
	_, ok := f.knownTiles[saveID]
	return ok
}

// SavedDataLoadingError is pocketmine\data\SavedDataLoadingException for tile data (this package
// can't import pocketmine/data's users).
type SavedDataLoadingError struct{ Message string }

func (e *SavedDataLoadingError) Error() string { return e.Message }

// CreateFromData is a port of TileFactory::createFromData: the tile saved in tag, or nil (and no
// error) if its save ID is unknown. Bad data is a *SavedDataLoadingError.
func (f *TileFactory) CreateFromData(world World, tag *nbt.CompoundTag) (Tile, error) {
	typ := string(tag.GetStringOr(TagID, ""))
	constructor, ok := f.knownTiles[typ]
	if !ok {
		return nil, nil
	}
	x, errX := tag.GetInt(TagX)
	y, errY := tag.GetInt(TagY)
	z, errZ := tag.GetInt(TagZ)
	for _, err := range []error{errX, errY, errZ} {
		if err != nil {
			return nil, &SavedDataLoadingError{err.Error()}
		}
	}
	t := constructor(world, math.NewVector3(float64(x), float64(y), float64(z)))
	if err := t.ReadSaveData(tag); err != nil {
		return nil, &SavedDataLoadingError{fmt.Sprintf("%v", err)}
	}
	return t, nil
}
