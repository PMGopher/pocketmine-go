package block

import (
	"fmt"
	"strings"
	"sync"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/item/enchantment"
	"pocketmine-go/pocketmine/math"
)

// This file is a port of pocketmine\block\VanillaBlocksInputs (the source PocketMine-MP generates
// VanillaBlocks from): every vanilla block, registered in the same order with the same names,
// break info and type tags. VanillaBlock(name) is VanillaBlocks::NAME(): it returns a new instance
// (a clone) of the registered block.

// vanillaHarvestLevel is a minimal ToolTier (GetHarvestLevel only): the real item.ToolTier lives in
// the item package, which this package can't import. The levels are item.ToolTier's.
type vanillaHarvestLevel int

func (h vanillaHarvestLevel) GetHarvestLevel() int { return int(h) }

const (
	vanillaToolTierWood    vanillaHarvestLevel = 1
	vanillaToolTierGold    vanillaHarvestLevel = 2
	vanillaToolTierStone   vanillaHarvestLevel = 3
	vanillaToolTierIron    vanillaHarvestLevel = 4
	vanillaToolTierDiamond vanillaHarvestLevel = 5
)

// vanillaBlocksRegistry is the RegistrySource VanillaBlocksInputs fills.
type vanillaBlocksRegistry struct {
	blocks  map[string]Behavior
	names   []string
	delayed []func()
	nextID  int
}

var (
	vanillaBlocks     *vanillaBlocksRegistry
	vanillaBlocksOnce sync.Once
)

func getVanillaBlocks() *vanillaBlocksRegistry {
	vanillaBlocksOnce.Do(func() {
		r := &vanillaBlocksRegistry{blocks: map[string]Behavior{}, nextID: FIRST_UNUSED_BLOCK_ID}
		r.setup()
		for _, register := range r.delayed {
			register()
		}
		r.delayed = nil
		vanillaBlocks = r
	})
	return vanillaBlocks
}

// VanillaBlock returns a new instance of the vanilla block registered under name (VanillaBlocks::NAME()).
// Panics if there is no such block.
func VanillaBlock(name string) Behavior {
	blk, ok := getVanillaBlocks().blocks[name]
	if !ok {
		panic(fmt.Sprintf("VanillaBlocks: no block registered as %q", name))
	}
	return blk.Clone()
}

// GetAllVanillaBlocks is VanillaBlocks::getAll(): a new instance of every vanilla block, keyed by
// registry name.
func GetAllVanillaBlocks() map[string]Behavior {
	r := getVanillaBlocks()
	result := make(map[string]Behavior, len(r.blocks))
	for name, blk := range r.blocks {
		result[name] = blk.Clone()
	}
	return result
}

// GetVanillaBlockNames returns the registry names in registration order.
func GetVanillaBlockNames() []string {
	return append([]string(nil), getVanillaBlocks().names...)
}

// makeBID is a port of VanillaBlocksInputs::makeBID: the type ID is the BlockTypeIds constant
// named after the block (a new ID is generated if there's none).
func (r *vanillaBlocksRegistry) makeBID(name string, newTile TileFactory) *BlockIdentifier {
	typeID, ok := typeIDsByName[strings.ToUpper(name)]
	if !ok {
		//this allows registering new stuff without adding new type ID constants
		typeID = r.nextID
		r.nextID++
	}
	id, err := NewBlockIdentifier(typeID, newTile)
	if err != nil {
		panic(err)
	}
	return id
}

// register is a port of VanillaBlocksInputs::register.
func (r *vanillaBlocksRegistry) register(name string, createBlock func(id *BlockIdentifier) Behavior, newTile ...TileFactory) Behavior {
	var factory TileFactory
	if len(newTile) > 0 {
		factory = newTile[0]
	}
	blk := createBlock(r.makeBID(name, factory))
	if _, exists := r.blocks[name]; exists {
		panic(fmt.Sprintf("VanillaBlocks: %q registered twice", name))
	}
	r.blocks[name] = blk
	r.names = append(r.names, name)
	return blk
}

// registerDelayed is RegistrySource::registerDelayed: the block is created after every other one,
// so it can refer to them (e.g. Nylium's vegetation).
func (r *vanillaBlocksRegistry) registerDelayed(name string, createBlock func(id *BlockIdentifier) Behavior, newTile ...TileFactory) {
	r.delayed = append(r.delayed, func() { r.register(name, createBlock, newTile...) })
}

// get returns the registered instance itself (only for use while setting up the registry).
func (r *vanillaBlocksRegistry) get(name string) Behavior {
	blk, ok := r.blocks[name]
	if !ok {
		panic(fmt.Sprintf("VanillaBlocks: %q isn't registered yet", name))
	}
	return blk
}

// BlockBreakInfo constructors with PHP's default parameters (Go has none).

func bNew(hardness float64, toolType ToolType, toolHarvestLevel int, blastResistance ...float64) *BlockBreakInfo {
	var br *float64
	if len(blastResistance) > 0 {
		br = &blastResistance[0]
	}
	return NewBlockBreakInfo(hardness, toolType, toolHarvestLevel, br, nil)
}

func optionalBlast(blastResistance []float64) *float64 {
	if len(blastResistance) > 0 {
		return &blastResistance[0]
	}
	return nil
}

func bPickaxe(hardness float64, toolTier ToolTier, blastResistance ...float64) *BlockBreakInfo {
	return BlockBreakInfoPickaxe(hardness, toolTier, optionalBlast(blastResistance))
}

func bAxe(hardness float64, toolTier ToolTier, blastResistance ...float64) *BlockBreakInfo {
	return BlockBreakInfoAxe(hardness, toolTier, optionalBlast(blastResistance))
}

func bShovel(hardness float64, toolTier ToolTier, blastResistance ...float64) *BlockBreakInfo {
	return BlockBreakInfoShovel(hardness, toolTier, optionalBlast(blastResistance))
}

func bInstant(toolType ToolType, toolHarvestLevel int) *BlockBreakInfo {
	return BlockBreakInfoInstant(toolType, toolHarvestLevel)
}

func bIndestructible(blastResistance ...float64) *BlockBreakInfo {
	if len(blastResistance) > 0 {
		return BlockBreakInfoIndestructible(blastResistance[0])
	}
	return BlockBreakInfoIndestructible(DefaultIndestructibleBlastResistance)
}

// newOpaque/newTransparent are `new Opaque(...)`/`new Transparent(...)`.
func newOpaque(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) Behavior {
	o := &Opaque{Block: NewBlock(idInfo, name, typeInfo)}
	o.Init(o)
	return o
}

func newTransparent(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) Behavior {
	t := &plainTransparent{Transparent{NewBlock(idInfo, name, typeInfo)}}
	t.Init(t)
	return t
}

// plainTransparent is a bare `new Transparent(...)` (barrier, invisible bedrock).
type plainTransparent struct {
	Transparent
}

func (t *plainTransparent) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

// Tile factories for the tile classes VanillaBlocksInputs passes (TileX::class).
func tileOf[T Tile](create func(world tile.World, pos math.Vector3) T) TileFactory {
	return func(world tile.World, pos math.Vector3) Tile { return create(world, pos) }
}

var (
	tileBanner            = tileOf(tile.NewBanner)
	tileBarrel            = tileOf(tile.NewBarrel)
	tileBeacon            = tileOf(tile.NewBeacon)
	tileBed               = tileOf(tile.NewBed)
	tileBell              = tileOf(tile.NewBell)
	tileBlastFurnace      = tileOf(tile.NewBlastFurnace)
	tileBrewingStand      = tileOf(tile.NewBrewingStand)
	tileCampfire          = tileOf(tile.NewCampfire)
	tileChest             = tileOf(tile.NewChest)
	tileChiseledBookshelf = tileOf(tile.NewChiseledBookshelf)
	tileComparator        = tileOf(tile.NewComparator)
	tileDaylightSensor    = tileOf(tile.NewDaylightSensor)
	tileEnchantingTable   = tileOf(tile.NewEnchantTable)
	tileEnderChest        = tileOf(tile.NewEnderChest)
	tileHangingSign       = tileOf(tile.NewHangingSign)
	tileHopper            = tileOf(tile.NewHopper)
	tileItemFrame         = tileOf(tile.NewItemFrame)
	tileJukebox           = tileOf(tile.NewJukebox)
	tileLectern           = tileOf(tile.NewLectern)
	tileMobHead           = tileOf(tile.NewMobHead)
	tileMonsterSpawner    = tileOf(tile.NewMonsterSpawner)
	tileNormalFurnace     = tileOf(tile.NewNormalFurnace)
	tileNote              = tileOf(tile.NewNote)
	tileShulkerBox        = tileOf(tile.NewShulkerBox)
	tileSign              = tileOf(tile.NewSign)
	tileSmoker            = tileOf(tile.NewSmoker)
	tileCauldron          = tileOf(tile.NewCauldron)
	tileFlowerPot         = tileOf(tile.NewFlowerPot)
	tileGlowingItemFrame  = tileOf(tile.NewGlowingItemFrame)
)

func (r *vanillaBlocksRegistry) setup() {
	r.register("air", func(id *BlockIdentifier) Behavior {
		return NewAir(id, "Air", NewBlockTypeInfo(bIndestructible(-1.0), nil, nil))
	})
	railBreakInfo := NewBlockTypeInfo(bNew(0.7, ToolTypeNone, 0), nil, nil)
	r.register("activator_rail", func(id *BlockIdentifier) Behavior { return NewActivatorRail(id, "Activator Rail", railBreakInfo) })
	r.register("anvil", func(id *BlockIdentifier) Behavior {
		return NewAnvil(id, "Anvil", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierWood, 6000.0), nil, nil))
	})
	r.register("azalea", func(id *BlockIdentifier) Behavior {
		return NewAzalea(id, "Azalea", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil))
	})
	r.register("flowering_azalea", func(id *BlockIdentifier) Behavior {
		return NewAzalea(id, "Flowering Azalea", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil))
	})
	r.register("bamboo", func(id *BlockIdentifier) Behavior {
		return NewBamboo(id, "Bamboo", NewBlockTypeInfo(bNew(1.0, ToolTypeAxe, 0).WithBreakTimeModifier(func(item Item, breakTime float64) float64 {
			if item.GetBlockToolType() == ToolTypeSword {
				return 0.0
			}
			return breakTime
		}), []string{BlockTypeTagsPottablePlants}, nil))
	})
	r.register("bamboo_sapling", func(id *BlockIdentifier) Behavior {
		return NewBambooSapling(id, "Bamboo Sapling", NewBlockTypeInfo(bNew(1.0, ToolTypeNone, 0), nil, nil))
	})
	bannerBreakInfo := NewBlockTypeInfo(bAxe(1.0, nil), nil, nil)
	r.register("banner", func(id *BlockIdentifier) Behavior { return NewFloorBanner(id, "Banner", bannerBreakInfo) }, tileBanner)
	r.register("wall_banner", func(id *BlockIdentifier) Behavior { return NewWallBanner(id, "Wall Banner", bannerBreakInfo) }, tileBanner)
	r.register("ominous_banner", func(id *BlockIdentifier) Behavior {
		return NewOminousFloorBanner(id, "Ominous Banner", bannerBreakInfo)
	}, tileBanner)
	r.register("ominous_wall_banner", func(id *BlockIdentifier) Behavior {
		return NewOminousWallBanner(id, "Ominous Wall Banner", bannerBreakInfo)
	}, tileBanner)
	r.register("barrel", func(id *BlockIdentifier) Behavior {
		return NewBarrel(id, "Barrel", NewBlockTypeInfo(bAxe(2.5, nil), nil, nil))
	}, tileBarrel)
	r.register("barrier", func(id *BlockIdentifier) Behavior {
		return newTransparent(id, "Barrier", NewBlockTypeInfo(bIndestructible(), nil, nil))
	})
	r.register("beacon", func(id *BlockIdentifier) Behavior {
		return NewBeacon(id, "Beacon", NewBlockTypeInfo(bNew(3.0, ToolTypeNone, 0), nil, nil))
	}, tileBeacon)
	r.register("bed", func(id *BlockIdentifier) Behavior {
		return NewBed(id, "Bed Block", NewBlockTypeInfo(bNew(0.2, ToolTypeNone, 0), nil, nil))
	}, tileBed)
	r.register("bedrock", func(id *BlockIdentifier) Behavior {
		return NewBedrock(id, "Bedrock", NewBlockTypeInfo(bIndestructible(18000000.0), nil, nil))
	})
	r.register("beetroots", func(id *BlockIdentifier) Behavior {
		return NewBeetroot(id, "Beetroot Block", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("bell", func(id *BlockIdentifier) Behavior {
		return NewBell(id, "Bell", NewBlockTypeInfo(bPickaxe(5.0, nil), nil, nil))
	}, tileBell)
	r.register("blue_ice", func(id *BlockIdentifier) Behavior {
		return NewBlueIce(id, "Blue Ice", NewBlockTypeInfo(bPickaxe(2.8, nil), nil, nil))
	})
	r.register("bone_block", func(id *BlockIdentifier) Behavior {
		return NewBoneBlock(id, "Bone Block", NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood), nil, nil))
	})
	r.register("bookshelf", func(id *BlockIdentifier) Behavior {
		return NewBookshelf(id, "Bookshelf", NewBlockTypeInfo(bAxe(1.5, nil), nil, nil))
	})
	r.register("chiseled_bookshelf", func(id *BlockIdentifier) Behavior {
		return NewChiseledBookshelf(id, "Chiseled Bookshelf", NewBlockTypeInfo(bAxe(1.5, nil), nil, nil))
	}, tileChiseledBookshelf)
	r.register("brewing_stand", func(id *BlockIdentifier) Behavior {
		return NewBrewingStand(id, "Brewing Stand", NewBlockTypeInfo(bPickaxe(0.5, nil), nil, nil))
	}, tileBrewingStand)
	bricksBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood, 30.0), nil, nil)
	r.register("brick_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Brick Stairs", bricksBreakInfo) })
	r.register("bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Bricks", bricksBreakInfo) })
	r.register("brown_mushroom", func(id *BlockIdentifier) Behavior {
		return NewBrownMushroom(id, "Brown Mushroom", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil))
	})
	r.register("cactus", func(id *BlockIdentifier) Behavior {
		return NewCactus(id, "Cactus", NewBlockTypeInfo(bNew(0.4, ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil))
	})
	r.register("cake", func(id *BlockIdentifier) Behavior {
		return NewCake(id, "Cake", NewBlockTypeInfo(bNew(0.5, ToolTypeNone, 0), nil, nil))
	})
	campfireBreakInfo := NewBlockTypeInfo(bAxe(2.0, nil), nil, nil)
	r.register("campfire", func(id *BlockIdentifier) Behavior { return NewCampfire(id, "Campfire", campfireBreakInfo) }, tileCampfire)
	r.register("soul_campfire", func(id *BlockIdentifier) Behavior { return NewSoulCampfire(id, "Soul Campfire", campfireBreakInfo) }, tileCampfire)
	r.register("carrots", func(id *BlockIdentifier) Behavior {
		return NewCarrot(id, "Carrot Block", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	chestBreakInfo := NewBlockTypeInfo(bAxe(2.5, nil), nil, nil)
	r.register("chest", func(id *BlockIdentifier) Behavior { return NewChest(id, "Chest", chestBreakInfo) }, tileChest)
	r.register("clay", func(id *BlockIdentifier) Behavior {
		return NewClay(id, "Clay Block", NewBlockTypeInfo(bShovel(0.6, nil), nil, nil))
	})
	r.register("coal", func(id *BlockIdentifier) Behavior {
		return NewCoal(id, "Coal Block", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierWood, 30.0), nil, nil))
	})
	cobblestoneBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood, 30.0), nil, nil)
	cobblestone := r.register("cobblestone", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Cobblestone", cobblestoneBreakInfo) })
	r.register("mossy_cobblestone", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Mossy Cobblestone", cobblestoneBreakInfo) })
	r.register("cobblestone_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Cobblestone Stairs", cobblestoneBreakInfo) })
	r.register("mossy_cobblestone_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Mossy Cobblestone Stairs", cobblestoneBreakInfo)
	})
	r.register("cobweb", func(id *BlockIdentifier) Behavior {
		return NewCobweb(id, "Cobweb", NewBlockTypeInfo(bNew(4.0, ToolTypeSword|ToolTypeShears, 1), nil, nil))
	})
	r.register("cocoa_pod", func(id *BlockIdentifier) Behavior {
		return NewCocoaBlock(id, "Cocoa Block", NewBlockTypeInfo(bAxe(0.2, nil, 15.0), nil, nil))
	})
	r.register("coral_block", func(id *BlockIdentifier) Behavior {
		return NewCoralBlock(id, "Coral Block", NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood, 30.0), nil, nil))
	})
	r.register("daylight_sensor", func(id *BlockIdentifier) Behavior {
		return NewDaylightSensor(id, "Daylight Sensor", NewBlockTypeInfo(bAxe(0.2, nil), nil, nil))
	}, tileDaylightSensor)
	r.register("dead_bush", func(id *BlockIdentifier) Behavior {
		return NewDeadBush(id, "Dead Bush", NewBlockTypeInfo(bInstant(ToolTypeShears, 1), []string{BlockTypeTagsPottablePlants}, nil))
	})
	r.register("detector_rail", func(id *BlockIdentifier) Behavior { return NewDetectorRail(id, "Detector Rail", railBreakInfo) })
	r.register("diamond", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Diamond Block", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierIron, 30.0), nil, nil))
	})
	r.register("dirt", func(id *BlockIdentifier) Behavior {
		return NewDirt(id, "Dirt", NewBlockTypeInfo(bShovel(0.5, nil), []string{BlockTypeTagsDirt}, nil))
	})
	r.register("sunflower", func(id *BlockIdentifier) Behavior {
		return NewDoublePlant(id, "Sunflower", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("lilac", func(id *BlockIdentifier) Behavior {
		return NewDoublePlant(id, "Lilac", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("rose_bush", func(id *BlockIdentifier) Behavior {
		return NewDoublePlant(id, "Rose Bush", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("peony", func(id *BlockIdentifier) Behavior {
		return NewDoublePlant(id, "Peony", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("pink_petals", func(id *BlockIdentifier) Behavior {
		return NewPinkPetals(id, "Pink Petals", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("double_tallgrass", func(id *BlockIdentifier) Behavior {
		return NewDoubleTallGrass(id, "Double Tallgrass", NewBlockTypeInfo(bInstant(ToolTypeShears, 1), nil, nil))
	})
	r.register("large_fern", func(id *BlockIdentifier) Behavior {
		return NewDoubleTallGrass(id, "Large Fern", NewBlockTypeInfo(bInstant(ToolTypeShears, 1), nil, nil))
	})
	r.register("pitcher_plant", func(id *BlockIdentifier) Behavior {
		return NewDoublePlant(id, "Pitcher Plant", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("pitcher_crop", func(id *BlockIdentifier) Behavior {
		return NewPitcherCrop(id, "Pitcher Crop", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("double_pitcher_crop", func(id *BlockIdentifier) Behavior {
		return NewDoublePitcherCrop(id, "Double Pitcher Crop", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("dragon_egg", func(id *BlockIdentifier) Behavior {
		return NewDragonEgg(id, "Dragon Egg", NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierWood, 45.0), nil, nil))
	})
	r.register("dried_kelp", func(id *BlockIdentifier) Behavior {
		return NewDriedKelp(id, "Dried Kelp Block", NewBlockTypeInfo(bNew(0.5, ToolTypeNone, 0, 12.5), nil, nil))
	})
	r.register("emerald", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Emerald Block", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierIron, 30.0), nil, nil))
	})
	r.register("enchanting_table", func(id *BlockIdentifier) Behavior {
		return NewEnchantingTable(id, "Enchanting Table", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierWood, 6000.0), nil, nil))
	}, tileEnchantingTable)
	r.register("end_portal_frame", func(id *BlockIdentifier) Behavior {
		return NewEndPortalFrame(id, "End Portal Frame", NewBlockTypeInfo(bIndestructible(18000000.0), nil, nil))
	})
	r.register("end_rod", func(id *BlockIdentifier) Behavior {
		return NewEndRod(id, "End Rod", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("end_stone", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "End Stone", NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierWood, 45.0), nil, nil))
	})
	endBrickBreakInfo := NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierWood, 45.0), nil, nil)
	r.register("end_stone_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "End Stone Bricks", endBrickBreakInfo) })
	r.register("end_stone_brick_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "End Stone Brick Stairs", endBrickBreakInfo) })
	r.register("ender_chest", func(id *BlockIdentifier) Behavior {
		return NewEnderChest(id, "Ender Chest", NewBlockTypeInfo(bPickaxe(22.5, nil, 3000.0), nil, nil))
	}, tileEnderChest)
	r.register("farmland", func(id *BlockIdentifier) Behavior {
		return NewFarmland(id, "Farmland", NewBlockTypeInfo(bShovel(0.6, nil), []string{BlockTypeTagsDirt}, nil))
	})
	r.register("fire", func(id *BlockIdentifier) Behavior {
		return NewFire(id, "Fire Block", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsFire}, nil))
	})
	flowerTypeInfo := NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil)
	r.register("dandelion", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Dandelion", flowerTypeInfo) })
	r.register("poppy", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Poppy", flowerTypeInfo) })
	r.register("allium", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Allium", flowerTypeInfo) })
	r.register("azure_bluet", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Azure Bluet", flowerTypeInfo) })
	r.register("blue_orchid", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Blue Orchid", flowerTypeInfo) })
	r.register("cornflower", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Cornflower", flowerTypeInfo) })
	r.register("lily_of_the_valley", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Lily of the Valley", flowerTypeInfo) })
	r.register("orange_tulip", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Orange Tulip", flowerTypeInfo) })
	r.register("oxeye_daisy", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Oxeye Daisy", flowerTypeInfo) })
	r.register("pink_tulip", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Pink Tulip", flowerTypeInfo) })
	r.register("red_tulip", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Red Tulip", flowerTypeInfo) })
	r.register("white_tulip", func(id *BlockIdentifier) Behavior { return NewFlower(id, "White Tulip", flowerTypeInfo) })
	r.register("torchflower", func(id *BlockIdentifier) Behavior { return NewFlower(id, "Torchflower", flowerTypeInfo) })
	r.register("torchflower_crop", func(id *BlockIdentifier) Behavior {
		return NewTorchflowerCrop(id, "Torchflower Crop", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("flower_pot", func(id *BlockIdentifier) Behavior {
		return NewFlowerPot(id, "Flower Pot", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	}, tileFlowerPot)
	r.register("frosted_ice", func(id *BlockIdentifier) Behavior {
		return NewFrostedIce(id, "Frosted Ice", NewBlockTypeInfo(bPickaxe(0.5, nil), nil, nil))
	})
	r.register("furnace", func(id *BlockIdentifier) Behavior {
		return NewFurnace(id, "Furnace", NewBlockTypeInfo(bPickaxe(3.5, vanillaToolTierWood), nil, nil), tile.FurnaceTypeFurnace)
	}, tileNormalFurnace)
	r.register("blast_furnace", func(id *BlockIdentifier) Behavior {
		return NewFurnace(id, "Blast Furnace", NewBlockTypeInfo(bPickaxe(3.5, vanillaToolTierWood), nil, nil), tile.FurnaceTypeBlastFurnace)
	}, tileBlastFurnace)
	r.register("smoker", func(id *BlockIdentifier) Behavior {
		return NewFurnace(id, "Smoker", NewBlockTypeInfo(bPickaxe(3.5, vanillaToolTierWood), nil, nil), tile.FurnaceTypeSmoker)
	}, tileSmoker)
	glassBreakInfo := NewBlockTypeInfo(bNew(0.3, ToolTypeNone, 0), nil, nil)
	r.register("glass", func(id *BlockIdentifier) Behavior { return NewGlass(id, "Glass", glassBreakInfo) })
	r.register("glass_pane", func(id *BlockIdentifier) Behavior { return NewGlassPane(id, "Glass Pane", glassBreakInfo) })
	r.register("glowing_obsidian", func(id *BlockIdentifier) Behavior {
		return NewGlowingObsidian(id, "Glowing Obsidian", NewBlockTypeInfo(bPickaxe(35.0, vanillaToolTierDiamond, 6000.0), nil, nil))
	})
	r.register("glowstone", func(id *BlockIdentifier) Behavior {
		return NewGlowstone(id, "Glowstone", NewBlockTypeInfo(bPickaxe(0.3, nil), nil, nil))
	})
	r.register("glow_lichen", func(id *BlockIdentifier) Behavior {
		return NewGlowLichen(id, "Glow Lichen", NewBlockTypeInfo(bAxe(0.2, nil), nil, nil))
	})
	r.register("gold", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Gold Block", NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierIron, 30.0), nil, nil))
	})
	r.register("grass", func(id *BlockIdentifier) Behavior {
		return NewGrass(id, "Grass", NewBlockTypeInfo(bShovel(0.6, nil), []string{BlockTypeTagsDirt}, nil))
	})
	r.register("grass_path", func(id *BlockIdentifier) Behavior {
		return NewGrassPath(id, "Grass Path", NewBlockTypeInfo(bShovel(0.65, nil), nil, nil))
	})
	r.register("gravel", func(id *BlockIdentifier) Behavior {
		return NewGravel(id, "Gravel", NewBlockTypeInfo(bShovel(0.6, nil), nil, nil))
	})
	r.register("hardened_clay", func(id *BlockIdentifier) Behavior {
		return NewHardenedClay(id, "Hardened Clay", NewBlockTypeInfo(bPickaxe(1.25, vanillaToolTierWood, 21.0), nil, nil))
	})
	hardenedGlassBreakInfo := NewBlockTypeInfo(bNew(10.0, ToolTypeNone, 0), nil, nil)
	r.register("hardened_glass", func(id *BlockIdentifier) Behavior {
		return NewHardenedGlass(id, "Hardened Glass", hardenedGlassBreakInfo)
	})
	r.register("hardened_glass_pane", func(id *BlockIdentifier) Behavior {
		return NewHardenedGlassPane(id, "Hardened Glass Pane", hardenedGlassBreakInfo)
	})
	r.register("hay_bale", func(id *BlockIdentifier) Behavior {
		return NewHayBale(id, "Hay Bale", NewBlockTypeInfo(bNew(0.5, ToolTypeNone, 0), nil, nil))
	})
	r.register("hopper", func(id *BlockIdentifier) Behavior {
		return NewHopper(id, "Hopper", NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierWood, 24.0), nil, nil))
	}, tileHopper)
	r.register("ice", func(id *BlockIdentifier) Behavior {
		return NewIce(id, "Ice", NewBlockTypeInfo(bPickaxe(0.5, nil), nil, nil))
	})
	updateBlockBreakInfo := NewBlockTypeInfo(bNew(1.0, ToolTypeNone, 0), nil, nil)
	r.register("info_update", func(id *BlockIdentifier) Behavior { return newOpaque(id, "update!", updateBlockBreakInfo) })
	r.register("info_update2", func(id *BlockIdentifier) Behavior { return newOpaque(id, "ate!upd", updateBlockBreakInfo) })
	r.register("invisible_bedrock", func(id *BlockIdentifier) Behavior {
		return newTransparent(id, "Invisible Bedrock", NewBlockTypeInfo(bIndestructible(18000000.0), nil, nil))
	})
	ironBreakInfo := NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierStone, 30.0), nil, nil)
	r.register("iron", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Iron Block", ironBreakInfo) })
	r.register("iron_bars", func(id *BlockIdentifier) Behavior { return NewThin(id, "Iron Bars", ironBreakInfo) })
	r.register("copper_bars", func(id *BlockIdentifier) Behavior { return NewCopperBars(id, "Copper Bars", ironBreakInfo) })
	r.register("iron_door", func(id *BlockIdentifier) Behavior {
		return NewDoor(id, "Iron Door", NewBlockTypeInfo(bPickaxe(5.0, nil), nil, nil))
	})
	r.register("iron_trapdoor", func(id *BlockIdentifier) Behavior {
		return NewTrapdoor(id, "Iron Trapdoor", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierWood), nil, nil))
	})
	itemFrameInfo := NewBlockTypeInfo(bNew(0.25, ToolTypeNone, 0), nil, nil)
	r.register("item_frame", func(id *BlockIdentifier) Behavior { return NewItemFrame(id, "Item Frame", itemFrameInfo) }, tileItemFrame)
	r.register("glowing_item_frame", func(id *BlockIdentifier) Behavior { return NewItemFrame(id, "Glow Item Frame", itemFrameInfo) }, tileGlowingItemFrame)
	r.register("jukebox", func(id *BlockIdentifier) Behavior {
		return NewJukebox(id, "Jukebox", NewBlockTypeInfo(bAxe(2.0, nil, 30.0), nil, nil))
	}, tileJukebox)
	r.register("ladder", func(id *BlockIdentifier) Behavior {
		return NewLadder(id, "Ladder", NewBlockTypeInfo(bAxe(0.4, nil), nil, nil))
	})
	lanternBreakInfo := NewBlockTypeInfo(bPickaxe(3.5, nil), nil, nil)
	r.register("lantern", func(id *BlockIdentifier) Behavior { return NewLantern(id, "Lantern", lanternBreakInfo, 15) })
	r.register("soul_lantern", func(id *BlockIdentifier) Behavior { return NewLantern(id, "Soul Lantern", lanternBreakInfo, 10) })
	r.register("copper_lantern", func(id *BlockIdentifier) Behavior {
		return NewCopperLantern(id, "Copper Lantern", lanternBreakInfo, 15)
	})
	r.register("lapis_lazuli", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Lapis Lazuli Block", NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierStone), nil, nil))
	})
	r.register("lava", func(id *BlockIdentifier) Behavior {
		return NewLava(id, "Lava", NewBlockTypeInfo(bIndestructible(500.0), nil, nil))
	})
	r.register("lectern", func(id *BlockIdentifier) Behavior {
		return NewLectern(id, "Lectern", NewBlockTypeInfo(bAxe(2.5, nil), nil, nil))
	}, tileLectern)
	r.register("lever", func(id *BlockIdentifier) Behavior {
		return NewLever(id, "Lever", NewBlockTypeInfo(bNew(0.5, ToolTypeNone, 0), nil, nil))
	})
	r.register("magma", func(id *BlockIdentifier) Behavior {
		return NewMagma(id, "Magma Block", NewBlockTypeInfo(bPickaxe(0.5, vanillaToolTierWood), nil, nil))
	})
	r.register("melon", func(id *BlockIdentifier) Behavior {
		return NewMelon(id, "Melon Block", NewBlockTypeInfo(bAxe(1.0, nil), nil, nil))
	})
	r.register("melon_stem", func(id *BlockIdentifier) Behavior {
		return NewMelonStem(id, "Melon Stem", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("monster_spawner", func(id *BlockIdentifier) Behavior {
		return NewMonsterSpawner(id, "Monster Spawner", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierWood), nil, nil))
	}, tileMonsterSpawner)
	r.register("mycelium", func(id *BlockIdentifier) Behavior {
		return NewMycelium(id, "Mycelium", NewBlockTypeInfo(bShovel(0.6, nil), []string{BlockTypeTagsDirt}, nil))
	})
	netherBrickBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood, 30.0), nil, nil)
	r.register("nether_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Nether Bricks", netherBrickBreakInfo) })
	r.register("red_nether_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Red Nether Bricks", netherBrickBreakInfo) })
	r.register("nether_brick_fence", func(id *BlockIdentifier) Behavior { return NewFence(id, "Nether Brick Fence", netherBrickBreakInfo) })
	r.register("nether_brick_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Nether Brick Stairs", netherBrickBreakInfo) })
	r.register("red_nether_brick_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Red Nether Brick Stairs", netherBrickBreakInfo)
	})
	r.register("chiseled_nether_bricks", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Chiseled Nether Bricks", netherBrickBreakInfo)
	})
	r.register("cracked_nether_bricks", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Cracked Nether Bricks", netherBrickBreakInfo)
	})
	r.register("nether_portal", func(id *BlockIdentifier) Behavior {
		return NewNetherPortal(id, "Nether Portal", NewBlockTypeInfo(bIndestructible(0.0), nil, nil))
	})
	r.register("nether_reactor_core", func(id *BlockIdentifier) Behavior {
		return NewNetherReactor(id, "Nether Reactor Core", NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierWood), nil, nil))
	})
	r.register("nether_wart_block", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Nether Wart Block", NewBlockTypeInfo(bNew(1.0, ToolTypeHoe, 0), nil, nil))
	})
	r.register("nether_wart", func(id *BlockIdentifier) Behavior {
		return NewNetherWartPlant(id, "Nether Wart", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("netherrack", func(id *BlockIdentifier) Behavior {
		return NewNetherrack(id, "Netherrack", NewBlockTypeInfo(bPickaxe(0.4, vanillaToolTierWood), nil, nil))
	})
	r.register("note_block", func(id *BlockIdentifier) Behavior {
		return NewNote(id, "Note Block", NewBlockTypeInfo(bAxe(0.8, nil), nil, nil))
	}, tileNote)
	r.register("obsidian", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Obsidian", NewBlockTypeInfo(bPickaxe(35.0, vanillaToolTierDiamond, 6000.0), nil, nil))
	})
	r.register("packed_ice", func(id *BlockIdentifier) Behavior {
		return NewPackedIce(id, "Packed Ice", NewBlockTypeInfo(bPickaxe(0.5, nil), nil, nil))
	})
	r.register("podzol", func(id *BlockIdentifier) Behavior {
		return NewPodzol(id, "Podzol", NewBlockTypeInfo(bShovel(0.5, nil), []string{BlockTypeTagsDirt}, nil))
	})
	r.register("potatoes", func(id *BlockIdentifier) Behavior {
		return NewPotato(id, "Potato Block", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("powered_rail", func(id *BlockIdentifier) Behavior { return NewPoweredRail(id, "Powered Rail", railBreakInfo) })
	prismarineBreakInfo := NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("prismarine", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Prismarine", prismarineBreakInfo) })
	r.register("dark_prismarine", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Dark Prismarine", prismarineBreakInfo) })
	r.register("prismarine_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Prismarine Bricks", prismarineBreakInfo) })
	r.register("prismarine_bricks_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Prismarine Bricks Stairs", prismarineBreakInfo)
	})
	r.register("dark_prismarine_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Dark Prismarine Stairs", prismarineBreakInfo) })
	r.register("prismarine_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Prismarine Stairs", prismarineBreakInfo) })
	pumpkinBreakInfo := NewBlockTypeInfo(bAxe(1.0, nil), nil, nil)
	r.register("pumpkin", func(id *BlockIdentifier) Behavior { return NewPumpkin(id, "Pumpkin", pumpkinBreakInfo) })
	r.register("carved_pumpkin", func(id *BlockIdentifier) Behavior {
		return NewCarvedPumpkin(id, "Carved Pumpkin", NewBlockTypeInfo(bAxe(1.0, nil), nil, []string{enchantment.TagMask}))
	})
	r.register("lit_pumpkin", func(id *BlockIdentifier) Behavior { return NewLitPumpkin(id, "Jack o'Lantern", pumpkinBreakInfo) })
	r.register("pumpkin_stem", func(id *BlockIdentifier) Behavior {
		return NewPumpkinStem(id, "Pumpkin Stem", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	purpurBreakInfo := NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("purpur", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Purpur Block", purpurBreakInfo) })
	r.register("purpur_pillar", func(id *BlockIdentifier) Behavior { return NewSimplePillar(id, "Purpur Pillar", purpurBreakInfo) })
	r.register("purpur_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Purpur Stairs", purpurBreakInfo) })
	quartzBreakInfo := NewBlockTypeInfo(bPickaxe(0.8, vanillaToolTierWood), nil, nil)
	smoothQuartzBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood, 30.0), nil, nil)
	r.register("quartz", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Quartz Block", quartzBreakInfo) })
	r.register("chiseled_quartz", func(id *BlockIdentifier) Behavior {
		return NewSimplePillar(id, "Chiseled Quartz Block", quartzBreakInfo)
	})
	r.register("quartz_pillar", func(id *BlockIdentifier) Behavior { return NewSimplePillar(id, "Quartz Pillar", quartzBreakInfo) })
	r.register("smooth_quartz", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Smooth Quartz Block", smoothQuartzBreakInfo) })
	r.register("quartz_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Quartz Bricks", quartzBreakInfo) })
	r.register("quartz_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Quartz Stairs", quartzBreakInfo) })
	r.register("smooth_quartz_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Smooth Quartz Stairs", smoothQuartzBreakInfo) })
	r.register("rail", func(id *BlockIdentifier) Behavior { return NewRail(id, "Rail", railBreakInfo) })
	r.register("red_mushroom", func(id *BlockIdentifier) Behavior {
		return NewRedMushroom(id, "Red Mushroom", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil))
	})
	r.register("redstone", func(id *BlockIdentifier) Behavior {
		return NewRedstone(id, "Redstone Block", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierWood, 30.0), nil, nil))
	})
	r.register("redstone_comparator", func(id *BlockIdentifier) Behavior {
		return NewRedstoneComparator(id, "Redstone Comparator", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	}, tileComparator)
	r.register("redstone_lamp", func(id *BlockIdentifier) Behavior {
		return NewRedstoneLamp(id, "Redstone Lamp", NewBlockTypeInfo(bNew(0.3, ToolTypeNone, 0), nil, nil))
	})
	r.register("redstone_repeater", func(id *BlockIdentifier) Behavior {
		return NewRedstoneRepeater(id, "Redstone Repeater", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("redstone_torch", func(id *BlockIdentifier) Behavior {
		return NewRedstoneTorch(id, "Redstone Torch", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("redstone_wire", func(id *BlockIdentifier) Behavior {
		return NewRedstoneWire(id, "Redstone", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("reserved6", func(id *BlockIdentifier) Behavior {
		return NewReserved6(id, "reserved6", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	sandTypeInfo := NewBlockTypeInfo(bShovel(0.5, nil), []string{BlockTypeTagsSand}, nil)
	r.register("sand", func(id *BlockIdentifier) Behavior { return NewSand(id, "Sand", sandTypeInfo) })
	r.register("red_sand", func(id *BlockIdentifier) Behavior { return NewSand(id, "Red Sand", sandTypeInfo) })
	r.register("sea_lantern", func(id *BlockIdentifier) Behavior {
		return NewSeaLantern(id, "Sea Lantern", NewBlockTypeInfo(bNew(0.3, ToolTypeNone, 0), nil, nil))
	})
	r.register("sea_pickle", func(id *BlockIdentifier) Behavior {
		return NewSeaPickle(id, "Sea Pickle", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("mob_head", func(id *BlockIdentifier) Behavior {
		return NewMobHead(id, "Mob Head", NewBlockTypeInfo(bNew(1.0, ToolTypeNone, 0), nil, []string{enchantment.TagMask}))
	}, tileMobHead)
	r.register("slime", func(id *BlockIdentifier) Behavior {
		return NewSlime(id, "Slime Block", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("snow", func(id *BlockIdentifier) Behavior {
		return NewSnow(id, "Snow Block", NewBlockTypeInfo(bShovel(0.2, vanillaToolTierWood), nil, nil))
	})
	r.register("snow_layer", func(id *BlockIdentifier) Behavior {
		return NewSnowLayer(id, "Snow Layer", NewBlockTypeInfo(bShovel(0.1, vanillaToolTierWood), nil, nil))
	})
	r.register("soul_sand", func(id *BlockIdentifier) Behavior {
		return NewSoulSand(id, "Soul Sand", NewBlockTypeInfo(bShovel(0.5, nil), nil, nil))
	})
	r.register("sponge", func(id *BlockIdentifier) Behavior {
		return NewSponge(id, "Sponge", NewBlockTypeInfo(bNew(0.6, ToolTypeHoe, 0), nil, nil))
	})
	shulkerBoxBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, nil), nil, nil)
	r.register("shulker_box", func(id *BlockIdentifier) Behavior { return NewShulkerBox(id, "Shulker Box", shulkerBoxBreakInfo) }, tileShulkerBox)
	stoneBreakInfo := NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood, 30.0), nil, nil)
	stone := r.register("stone", func(id *BlockIdentifier) Behavior { return NewStone(id, "Stone", stoneBreakInfo) })
	r.register("andesite", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Andesite", stoneBreakInfo) })
	r.register("diorite", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Diorite", stoneBreakInfo) })
	r.register("granite", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Granite", stoneBreakInfo) })
	r.register("polished_andesite", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Polished Andesite", stoneBreakInfo) })
	r.register("polished_diorite", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Polished Diorite", stoneBreakInfo) })
	r.register("polished_granite", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Polished Granite", stoneBreakInfo) })
	stoneBrick := r.register("stone_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Stone Bricks", stoneBreakInfo) })
	mossyStoneBrick := r.register("mossy_stone_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Mossy Stone Bricks", stoneBreakInfo) })
	crackedStoneBrick := r.register("cracked_stone_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Cracked Stone Bricks", stoneBreakInfo) })
	chiseledStoneBrick := r.register("chiseled_stone_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Chiseled Stone Bricks", stoneBreakInfo) })
	infestedStoneBreakInfo := NewBlockTypeInfo(bPickaxe(0.75, nil), nil, nil)
	r.register("infested_stone", func(id *BlockIdentifier) Behavior {
		return NewInfestedStone(id, "Infested Stone", infestedStoneBreakInfo, stone)
	})
	r.register("infested_stone_brick", func(id *BlockIdentifier) Behavior {
		return NewInfestedStone(id, "Infested Stone Brick", infestedStoneBreakInfo, stoneBrick)
	})
	r.register("infested_cobblestone", func(id *BlockIdentifier) Behavior {
		return NewInfestedStone(id, "Infested Cobblestone", NewBlockTypeInfo(bPickaxe(1.0, nil, 3.75), nil, nil), cobblestone)
	})
	r.register("infested_mossy_stone_brick", func(id *BlockIdentifier) Behavior {
		return NewInfestedStone(id, "Infested Mossy Stone Brick", infestedStoneBreakInfo, mossyStoneBrick)
	})
	r.register("infested_cracked_stone_brick", func(id *BlockIdentifier) Behavior {
		return NewInfestedStone(id, "Infested Cracked Stone Brick", infestedStoneBreakInfo, crackedStoneBrick)
	})
	r.register("infested_chiseled_stone_brick", func(id *BlockIdentifier) Behavior {
		return NewInfestedStone(id, "Infested Chiseled Stone Brick", infestedStoneBreakInfo, chiseledStoneBrick)
	})
	r.register("stone_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Stone Stairs", stoneBreakInfo) })
	r.register("smooth_stone", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Smooth Stone", NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood, 30.0), nil, nil))
	})
	r.register("andesite_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Andesite Stairs", stoneBreakInfo) })
	r.register("diorite_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Diorite Stairs", stoneBreakInfo) })
	r.register("granite_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Granite Stairs", stoneBreakInfo) })
	r.register("polished_andesite_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Polished Andesite Stairs", stoneBreakInfo) })
	r.register("polished_diorite_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Polished Diorite Stairs", stoneBreakInfo) })
	r.register("polished_granite_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Polished Granite Stairs", stoneBreakInfo) })
	r.register("stone_brick_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Stone Brick Stairs", stoneBreakInfo) })
	r.register("mossy_stone_brick_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Mossy Stone Brick Stairs", stoneBreakInfo) })
	r.register("stone_button", func(id *BlockIdentifier) Behavior {
		return NewStoneButton(id, "Stone Button", NewBlockTypeInfo(bPickaxe(0.5, nil), nil, nil))
	})
	r.register("stonecutter", func(id *BlockIdentifier) Behavior {
		return NewStonecutter(id, "Stonecutter", NewBlockTypeInfo(bPickaxe(3.5, nil), nil, nil))
	})
	r.register("stone_pressure_plate", func(id *BlockIdentifier) Behavior {
		return NewStonePressurePlate(id, "Stone Pressure Plate", NewBlockTypeInfo(bPickaxe(0.5, nil), nil, nil), 20)
	})
	stoneSlabBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood, 30.0), nil, nil)
	r.register("brick_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Brick", stoneSlabBreakInfo) })
	r.register("cobblestone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Cobblestone", stoneSlabBreakInfo) })
	r.register("fake_wooden_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Fake Wooden", stoneSlabBreakInfo) })
	r.register("nether_brick_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Nether Brick", stoneSlabBreakInfo) })
	r.register("quartz_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Quartz", stoneSlabBreakInfo) })
	r.register("sandstone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Sandstone", stoneSlabBreakInfo) })
	r.register("smooth_stone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Smooth Stone", stoneSlabBreakInfo) })
	r.register("stone_brick_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Stone Brick", stoneSlabBreakInfo) })
	r.register("red_nether_brick_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Red Nether Brick", stoneSlabBreakInfo) })
	r.register("red_sandstone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Red Sandstone", stoneSlabBreakInfo) })
	r.register("smooth_sandstone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Smooth Sandstone", stoneSlabBreakInfo) })
	r.register("cut_red_sandstone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Cut Red Sandstone", stoneSlabBreakInfo) })
	r.register("cut_sandstone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Cut Sandstone", stoneSlabBreakInfo) })
	r.register("mossy_cobblestone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Mossy Cobblestone", stoneSlabBreakInfo) })
	r.register("purpur_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Purpur", stoneSlabBreakInfo) })
	r.register("smooth_red_sandstone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Smooth Red Sandstone", stoneSlabBreakInfo) })
	r.register("smooth_quartz_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Smooth Quartz", stoneSlabBreakInfo) })
	r.register("stone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Stone", stoneSlabBreakInfo) })
	r.register("end_stone_brick_slab", func(id *BlockIdentifier) Behavior {
		return NewSlab(id, "End Stone Brick", NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierWood, 30.0), nil, nil))
	})
	lightStoneSlabBreakInfo := NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("dark_prismarine_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Dark Prismarine", lightStoneSlabBreakInfo) })
	r.register("prismarine_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Prismarine", lightStoneSlabBreakInfo) })
	r.register("prismarine_bricks_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Prismarine Bricks", lightStoneSlabBreakInfo) })
	r.register("andesite_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Andesite", lightStoneSlabBreakInfo) })
	r.register("diorite_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Diorite", lightStoneSlabBreakInfo) })
	r.register("granite_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Granite", lightStoneSlabBreakInfo) })
	r.register("polished_andesite_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Polished Andesite", lightStoneSlabBreakInfo) })
	r.register("polished_diorite_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Polished Diorite", lightStoneSlabBreakInfo) })
	r.register("polished_granite_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Polished Granite", lightStoneSlabBreakInfo) })
	r.register("mossy_stone_brick_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Mossy Stone Brick", lightStoneSlabBreakInfo) })
	r.register("legacy_stonecutter", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Legacy Stonecutter", NewBlockTypeInfo(bPickaxe(3.5, vanillaToolTierWood), nil, nil))
	})
	r.register("sugarcane", func(id *BlockIdentifier) Behavior {
		return NewSugarcane(id, "Sugarcane", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("sweet_berry_bush", func(id *BlockIdentifier) Behavior {
		return NewSweetBerryBush(id, "Sweet Berry Bush", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("tnt", func(id *BlockIdentifier) Behavior {
		return NewTNT(id, "TNT", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("fern", func(id *BlockIdentifier) Behavior {
		return NewTallGrass(id, "Fern", NewBlockTypeInfo(bInstant(ToolTypeShears, 1), []string{BlockTypeTagsPottablePlants}, nil), func() Behavior { return VanillaBlock("large_fern") })
	})
	r.register("tall_grass", func(id *BlockIdentifier) Behavior {
		return NewTallGrass(id, "Tall Grass", NewBlockTypeInfo(bInstant(ToolTypeShears, 1), nil, nil), func() Behavior { return VanillaBlock("double_tallgrass") })
	})
	r.register("blue_torch", func(id *BlockIdentifier) Behavior {
		return NewTorch(id, "Blue Torch", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("copper_torch", func(id *BlockIdentifier) Behavior {
		return NewTorch(id, "Copper Torch", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("purple_torch", func(id *BlockIdentifier) Behavior {
		return NewTorch(id, "Purple Torch", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("red_torch", func(id *BlockIdentifier) Behavior {
		return NewTorch(id, "Red Torch", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("green_torch", func(id *BlockIdentifier) Behavior {
		return NewTorch(id, "Green Torch", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("torch", func(id *BlockIdentifier) Behavior {
		return NewTorch(id, "Torch", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("trapped_chest", func(id *BlockIdentifier) Behavior { return NewTrappedChest(id, "Trapped Chest", chestBreakInfo) }, tileChest)
	r.register("tripwire", func(id *BlockIdentifier) Behavior {
		return NewTripwire(id, "Tripwire", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("tripwire_hook", func(id *BlockIdentifier) Behavior {
		return NewTripwireHook(id, "Tripwire Hook", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("underwater_torch", func(id *BlockIdentifier) Behavior {
		return NewUnderwaterTorch(id, "Underwater Torch", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("vines", func(id *BlockIdentifier) Behavior {
		return NewVine(id, "Vines", NewBlockTypeInfo(bAxe(0.2, nil), nil, nil))
	})
	r.register("water", func(id *BlockIdentifier) Behavior {
		return NewWater(id, "Water", NewBlockTypeInfo(bIndestructible(500.0), nil, nil))
	})
	r.register("lily_pad", func(id *BlockIdentifier) Behavior {
		return NewWaterLily(id, "Lily Pad", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	weightedPressurePlateBreakInfo := NewBlockTypeInfo(bPickaxe(0.5, nil), nil, nil)
	r.register("weighted_pressure_plate_heavy", func(id *BlockIdentifier) Behavior {
		return NewWeightedPressurePlateHeavy(id, "Weighted Pressure Plate Heavy", weightedPressurePlateBreakInfo, 10, 0.1)
	})
	r.register("weighted_pressure_plate_light", func(id *BlockIdentifier) Behavior {
		return NewWeightedPressurePlateLight(id, "Weighted Pressure Plate Light", weightedPressurePlateBreakInfo, 10, 1.0)
	})
	r.register("wheat", func(id *BlockIdentifier) Behavior {
		return NewWheat(id, "Wheat Block", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	leavesBreakInfo := NewBlockTypeInfo(bNew(0.2, ToolTypeHoe, 0).WithBreakTimeModifier(func(item Item, breakTime float64) float64 {
		if item.GetBlockToolType() == ToolTypeShears {
			return 0.0
		}
		return breakTime
	}), nil, nil)
	saplingTypeInfo := NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil)
	for _, saplingType := range blockutils.AllSaplingTypes {
		name := saplingType.GetDisplayName()
		r.register(saplingType.IDName()+"_sapling", func(id *BlockIdentifier) Behavior {
			return NewSapling(id, name+" Sapling", saplingTypeInfo, saplingType)
		})
	}
	for _, leavesType := range blockutils.AllLeavesTypes {
		name := leavesType.GetDisplayName()
		r.register(leavesType.IDName()+"_leaves", func(id *BlockIdentifier) Behavior { return NewLeaves(id, name+" Leaves", leavesBreakInfo, leavesType) })
	}
	sandstoneBreakInfo := NewBlockTypeInfo(bPickaxe(0.8, vanillaToolTierWood), nil, nil)
	smoothSandstoneBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood, 30.0), nil, nil)
	r.register("red_sandstone_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Red Sandstone Stairs", sandstoneBreakInfo) })
	r.register("smooth_red_sandstone_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Smooth Red Sandstone Stairs", smoothSandstoneBreakInfo)
	})
	r.register("red_sandstone", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Red Sandstone", sandstoneBreakInfo) })
	r.register("chiseled_red_sandstone", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Chiseled Red Sandstone", sandstoneBreakInfo) })
	r.register("cut_red_sandstone", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Cut Red Sandstone", sandstoneBreakInfo) })
	r.register("smooth_red_sandstone", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Smooth Red Sandstone", smoothSandstoneBreakInfo)
	})
	r.register("sandstone_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Sandstone Stairs", sandstoneBreakInfo) })
	r.register("smooth_sandstone_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Smooth Sandstone Stairs", smoothSandstoneBreakInfo)
	})
	r.register("sandstone", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Sandstone", sandstoneBreakInfo) })
	r.register("chiseled_sandstone", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Chiseled Sandstone", sandstoneBreakInfo) })
	r.register("cut_sandstone", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Cut Sandstone", sandstoneBreakInfo) })
	r.register("smooth_sandstone", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Smooth Sandstone", smoothSandstoneBreakInfo) })
	r.register("glazed_terracotta", func(id *BlockIdentifier) Behavior {
		return NewGlazedTerracotta(id, "Glazed Terracotta", NewBlockTypeInfo(bPickaxe(1.4, vanillaToolTierWood), nil, nil))
	})
	r.register("dyed_shulker_box", func(id *BlockIdentifier) Behavior {
		return NewDyedShulkerBox(id, "Dyed Shulker Box", shulkerBoxBreakInfo)
	}, tileShulkerBox)
	r.register("stained_glass", func(id *BlockIdentifier) Behavior { return NewStainedGlass(id, "Stained Glass", glassBreakInfo) })
	r.register("stained_glass_pane", func(id *BlockIdentifier) Behavior {
		return NewStainedGlassPane(id, "Stained Glass Pane", glassBreakInfo)
	})
	r.register("stained_clay", func(id *BlockIdentifier) Behavior {
		return NewStainedHardenedClay(id, "Stained Clay", NewBlockTypeInfo(bPickaxe(1.25, vanillaToolTierWood, 6.25), nil, nil))
	})
	r.register("stained_hardened_glass", func(id *BlockIdentifier) Behavior {
		return NewStainedHardenedGlass(id, "Stained Hardened Glass", hardenedGlassBreakInfo)
	})
	r.register("stained_hardened_glass_pane", func(id *BlockIdentifier) Behavior {
		return NewStainedHardenedGlassPane(id, "Stained Hardened Glass Pane", hardenedGlassBreakInfo)
	})
	r.register("carpet", func(id *BlockIdentifier) Behavior {
		return NewCarpet(id, "Carpet", NewBlockTypeInfo(bNew(0.1, ToolTypeNone, 0), nil, nil))
	})
	r.register("concrete", func(id *BlockIdentifier) Behavior {
		return NewConcrete(id, "Concrete", NewBlockTypeInfo(bPickaxe(1.8, vanillaToolTierWood), nil, nil))
	})
	r.register("concrete_powder", func(id *BlockIdentifier) Behavior {
		return NewConcretePowder(id, "Concrete Powder", NewBlockTypeInfo(bShovel(0.5, nil), nil, nil))
	})
	r.register("wool", func(id *BlockIdentifier) Behavior {
		return NewWool(id, "Wool", NewBlockTypeInfo(bNew(0.8, ToolTypeShears, 0).WithBreakTimeModifier(func(item Item, breakTime float64) float64 {
			if item.GetBlockToolType() == ToolTypeShears {
				breakTime *= 3 //shears break compatible blocks 15x faster, but wool 5x
			}
			return breakTime
		}), nil, nil))
	})
	r.register("end_stone_brick_wall", func(id *BlockIdentifier) Behavior {
		return NewWall(id, "End Stone Brick Wall", NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierWood, 45.0), nil, nil))
	})
	brickWallBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood, 30.0), nil, nil)
	r.register("cobblestone_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Cobblestone Wall", brickWallBreakInfo) })
	r.register("brick_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Brick Wall", brickWallBreakInfo) })
	r.register("mossy_cobblestone_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Mossy Cobblestone Wall", brickWallBreakInfo) })
	r.register("nether_brick_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Nether Brick Wall", brickWallBreakInfo) })
	r.register("red_nether_brick_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Red Nether Brick Wall", brickWallBreakInfo) })
	stoneWallBreakInfo := NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("stone_brick_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Stone Brick Wall", stoneWallBreakInfo) })
	r.register("mossy_stone_brick_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Mossy Stone Brick Wall", stoneWallBreakInfo) })
	r.register("granite_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Granite Wall", stoneWallBreakInfo) })
	r.register("diorite_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Diorite Wall", stoneWallBreakInfo) })
	r.register("andesite_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Andesite Wall", stoneWallBreakInfo) })
	r.register("prismarine_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Prismarine Wall", stoneWallBreakInfo) })
	sandstoneWallBreakInfo := NewBlockTypeInfo(bPickaxe(0.8, vanillaToolTierWood, 4.0), nil, nil)
	r.register("red_sandstone_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Red Sandstone Wall", sandstoneWallBreakInfo) })
	r.register("sandstone_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Sandstone Wall", sandstoneWallBreakInfo) })
	r.registerElements()
	chemistryTableBreakInfo := NewBlockTypeInfo(bPickaxe(2.5, vanillaToolTierWood), nil, nil)
	r.register("compound_creator", func(id *BlockIdentifier) Behavior {
		return NewChemistryTable(id, "Compound Creator", chemistryTableBreakInfo)
	})
	r.register("element_constructor", func(id *BlockIdentifier) Behavior {
		return NewChemistryTable(id, "Element Constructor", chemistryTableBreakInfo)
	})
	r.register("lab_table", func(id *BlockIdentifier) Behavior { return NewChemistryTable(id, "Lab Table", chemistryTableBreakInfo) })
	r.register("material_reducer", func(id *BlockIdentifier) Behavior {
		return NewChemistryTable(id, "Material Reducer", chemistryTableBreakInfo)
	})
	r.register("chemical_heat", func(id *BlockIdentifier) Behavior { return NewChemicalHeat(id, "Heat Block", chemistryTableBreakInfo) })
	r.registerMushroomBlocks()
	r.register("coral", func(id *BlockIdentifier) Behavior {
		return NewCoral(id, "Coral", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("coral_fan", func(id *BlockIdentifier) Behavior {
		return NewFloorCoralFan(id, "Coral Fan", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("wall_coral_fan", func(id *BlockIdentifier) Behavior {
		return NewWallCoralFan(id, "Wall Coral Fan", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("mangrove_roots", func(id *BlockIdentifier) Behavior {
		return NewMangroveRoots(id, "Mangrove Roots", NewBlockTypeInfo(bAxe(0.7, nil), nil, nil))
	})
	r.register("muddy_mangrove_roots", func(id *BlockIdentifier) Behavior {
		return NewSimplePillar(id, "Muddy Mangrove Roots", NewBlockTypeInfo(bShovel(0.7, nil), []string{BlockTypeTagsMud}, nil))
	})
	r.register("froglight", func(id *BlockIdentifier) Behavior {
		return NewFroglight(id, "Froglight", NewBlockTypeInfo(bNew(0.3, ToolTypeNone, 0), nil, nil))
	})
	r.register("sculk", func(id *BlockIdentifier) Behavior {
		return NewSculk(id, "Sculk", NewBlockTypeInfo(bNew(0.2, ToolTypeHoe, 0), nil, nil))
	})
	r.register("reinforced_deepslate", func(id *BlockIdentifier) Behavior {
		return NewReinforcedDeepslate(id, "Reinforced Deepslate", NewBlockTypeInfo(bNew(55.0, ToolTypeNone, 0, 6000.0), nil, nil))
	})
	r.register("cactus_flower", func(id *BlockIdentifier) Behavior {
		return NewCactusFlower(id, "Cactus Flower", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.registerBlocksR13()
	r.registerBlocksR14()
	r.registerBlocksR16()
	r.registerBlocksR17()
	r.registerBlocksR18()
	r.registerMudBlocks()
	r.registerResinBlocks()
	r.registerTuffBlocks()
	r.registerCraftingTables()
	r.registerChorusBlocks()
	r.registerOres()
	r.registerWoodenBlocks()
	r.registerCauldronBlocks()
}

func (r *vanillaBlocksRegistry) registerWoodenBlocks() {
	planksBreakInfo := NewBlockTypeInfo(bAxe(2.0, nil, 15.0), nil, nil)
	signBreakInfo := NewBlockTypeInfo(bAxe(1.0, nil), nil, nil)
	hangingSignBreakInfo := NewBlockTypeInfo(bAxe(1.0, nil), []string{BlockTypeTagsHangingSign}, nil)
	logBreakInfo := NewBlockTypeInfo(bAxe(2.0, nil), nil, nil)
	woodenDoorBreakInfo := NewBlockTypeInfo(bAxe(3.0, nil, 15.0), nil, nil)
	woodenButtonBreakInfo := NewBlockTypeInfo(bAxe(0.5, nil), nil, nil)
	woodenPressurePlateBreakInfo := NewBlockTypeInfo(bAxe(0.5, nil), nil, nil)
	for _, woodType := range blockutils.AllWoodTypes {
		name := woodType.GetDisplayName()
		idName := func(suffix string) string { return woodType.IDName() + "_" + suffix }

		logSuffix, ok := woodType.GetStandardLogSuffix()
		if !ok {
			logSuffix = "Log"
		}
		r.register(idName(strings.ToLower(logSuffix)), func(id *BlockIdentifier) Behavior { return NewWood(id, name+" "+logSuffix, logBreakInfo, woodType) })
		if woodType != blockutils.WoodTypeBamboo {
			//TODO: kinda sus hack - there's no all-sided log for bamboo
			//maybe log type and wood type need to be separated
			//we won't be able to do an overloaded accessor for wood until this is addressed
			woodSuffix, ok := woodType.GetAllSidedLogSuffix()
			if !ok {
				woodSuffix = "Wood"
			}
			r.register(idName(strings.ToLower(woodSuffix)), func(id *BlockIdentifier) Behavior { return NewWood(id, name+" "+woodSuffix, logBreakInfo, woodType) })
		}

		r.register(idName("planks"), func(id *BlockIdentifier) Behavior { return NewPlanks(id, name+" Planks", planksBreakInfo, woodType) })
		r.register(idName("fence"), func(id *BlockIdentifier) Behavior {
			return NewWoodenFence(id, name+" Fence", planksBreakInfo, woodType)
		})
		r.register(idName("slab"), func(id *BlockIdentifier) Behavior { return NewWoodenSlab(id, name, planksBreakInfo, woodType) })

		r.register(idName("fence_gate"), func(id *BlockIdentifier) Behavior {
			return NewFenceGate(id, name+" Fence Gate", planksBreakInfo, woodType)
		})
		r.register(idName("stairs"), func(id *BlockIdentifier) Behavior {
			return NewWoodenStairs(id, name+" Stairs", planksBreakInfo, woodType)
		})
		r.register(idName("door"), func(id *BlockIdentifier) Behavior {
			return NewWoodenDoor(id, name+" Door", woodenDoorBreakInfo, woodType)
		})

		r.register(idName("button"), func(id *BlockIdentifier) Behavior {
			return NewWoodenButton(id, name+" Button", woodenButtonBreakInfo, woodType)
		})
		r.register(idName("pressure_plate"), func(id *BlockIdentifier) Behavior {
			return NewWoodenPressurePlate(id, name+" Pressure Plate", woodenPressurePlateBreakInfo, woodType, 20)
		})
		r.register(idName("trapdoor"), func(id *BlockIdentifier) Behavior {
			return NewWoodenTrapdoor(id, name+" Trapdoor", woodenDoorBreakInfo, woodType)
		})

		// The sign item callbacks (getSignItemCallback/getHangingSignItemCallback) aren't passed: the Go
		// sign constructors don't take them (sign items aren't ported yet).
		r.registerDelayed(idName("sign"), func(id *BlockIdentifier) Behavior { return NewFloorSign(id, name+" Sign", signBreakInfo, woodType) }, tileSign)
		r.registerDelayed(idName("wall_sign"), func(id *BlockIdentifier) Behavior { return NewWallSign(id, name+" Wall Sign", signBreakInfo, woodType) }, tileSign)

		r.registerDelayed(idName("ceiling_center_hanging_sign"), func(id *BlockIdentifier) Behavior {
			return NewCeilingCenterHangingSign(id, name+" Center Hanging Sign", hangingSignBreakInfo, woodType)
		}, tileHangingSign)
		r.registerDelayed(idName("ceiling_edges_hanging_sign"), func(id *BlockIdentifier) Behavior {
			return NewCeilingEdgesHangingSign(id, name+" Edges Hanging Sign", hangingSignBreakInfo, woodType)
		}, tileHangingSign)
		r.registerDelayed(idName("wall_hanging_sign"), func(id *BlockIdentifier) Behavior {
			return NewWallHangingSign(id, name+" Wall Hanging Sign", hangingSignBreakInfo, woodType)
		}, tileHangingSign)
	}

	mosaicBreakInfo := NewBlockTypeInfo(bAxe(2.0, nil, 15.0), []string{BlockTypeTagsBambooMosaic}, nil)
	r.register("bamboo_mosaic", func(id *BlockIdentifier) Behavior {
		return NewPlanks(id, "Bamboo Mosaic", mosaicBreakInfo, blockutils.WoodTypeBamboo)
	})
	r.register("bamboo_mosaic_slab", func(id *BlockIdentifier) Behavior {
		return NewWoodenSlab(id, "Bamboo Mosaic", mosaicBreakInfo, blockutils.WoodTypeBamboo)
	})
	r.register("bamboo_mosaic_stairs", func(id *BlockIdentifier) Behavior {
		return NewWoodenStairs(id, "Bamboo Mosaic Stairs", mosaicBreakInfo, blockutils.WoodTypeBamboo)
	})
}

func (r *vanillaBlocksRegistry) registerMushroomBlocks() {
	mushroomBlockBreakInfo := NewBlockTypeInfo(bAxe(0.2, nil), nil, nil)
	r.register("brown_mushroom_block", func(id *BlockIdentifier) Behavior {
		return NewBrownMushroomBlock(id, "Brown Mushroom Block", mushroomBlockBreakInfo)
	})
	r.register("red_mushroom_block", func(id *BlockIdentifier) Behavior {
		return NewRedMushroomBlock(id, "Red Mushroom Block", mushroomBlockBreakInfo)
	})
	// //finally, the stems
	r.register("mushroom_stem", func(id *BlockIdentifier) Behavior {
		return NewMushroomStem(id, "Mushroom Stem", mushroomBlockBreakInfo)
	})
	r.register("all_sided_mushroom_stem", func(id *BlockIdentifier) Behavior {
		return NewMushroomStem(id, "All Sided Mushroom Stem", mushroomBlockBreakInfo)
	})
}

func (r *vanillaBlocksRegistry) registerElements() {
	instaBreak := NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil)
	r.register("element_zero", func(id *BlockIdentifier) Behavior { return newOpaque(id, "???", instaBreak) })

	register := func(name, displayName, symbol string, atomicWeight, group int) {
		r.register("element_"+name, func(id *BlockIdentifier) Behavior {
			return NewElement(id, displayName, instaBreak, symbol, atomicWeight, group)
		})
	}

	register("hydrogen", "Hydrogen", "h", 1, 5)
	register("helium", "Helium", "he", 2, 7)
	register("lithium", "Lithium", "li", 3, 0)
	register("beryllium", "Beryllium", "be", 4, 1)
	register("boron", "Boron", "b", 5, 4)
	register("carbon", "Carbon", "c", 6, 5)
	register("nitrogen", "Nitrogen", "n", 7, 5)
	register("oxygen", "Oxygen", "o", 8, 5)
	register("fluorine", "Fluorine", "f", 9, 6)
	register("neon", "Neon", "ne", 10, 7)
	register("sodium", "Sodium", "na", 11, 0)
	register("magnesium", "Magnesium", "mg", 12, 1)
	register("aluminum", "Aluminum", "al", 13, 3)
	register("silicon", "Silicon", "si", 14, 4)
	register("phosphorus", "Phosphorus", "p", 15, 5)
	register("sulfur", "Sulfur", "s", 16, 5)
	register("chlorine", "Chlorine", "cl", 17, 6)
	register("argon", "Argon", "ar", 18, 7)
	register("potassium", "Potassium", "k", 19, 0)
	register("calcium", "Calcium", "ca", 20, 1)
	register("scandium", "Scandium", "sc", 21, 2)
	register("titanium", "Titanium", "ti", 22, 2)
	register("vanadium", "Vanadium", "v", 23, 2)
	register("chromium", "Chromium", "cr", 24, 2)
	register("manganese", "Manganese", "mn", 25, 2)
	register("iron", "Iron", "fe", 26, 2)
	register("cobalt", "Cobalt", "co", 27, 2)
	register("nickel", "Nickel", "ni", 28, 2)
	register("copper", "Copper", "cu", 29, 2)
	register("zinc", "Zinc", "zn", 30, 2)
	register("gallium", "Gallium", "ga", 31, 3)
	register("germanium", "Germanium", "ge", 32, 4)
	register("arsenic", "Arsenic", "as", 33, 4)
	register("selenium", "Selenium", "se", 34, 5)
	register("bromine", "Bromine", "br", 35, 6)
	register("krypton", "Krypton", "kr", 36, 7)
	register("rubidium", "Rubidium", "rb", 37, 0)
	register("strontium", "Strontium", "sr", 38, 1)
	register("yttrium", "Yttrium", "y", 39, 2)
	register("zirconium", "Zirconium", "zr", 40, 2)
	register("niobium", "Niobium", "nb", 41, 2)
	register("molybdenum", "Molybdenum", "mo", 42, 2)
	register("technetium", "Technetium", "tc", 43, 2)
	register("ruthenium", "Ruthenium", "ru", 44, 2)
	register("rhodium", "Rhodium", "rh", 45, 2)
	register("palladium", "Palladium", "pd", 46, 2)
	register("silver", "Silver", "ag", 47, 2)
	register("cadmium", "Cadmium", "cd", 48, 2)
	register("indium", "Indium", "in", 49, 3)
	register("tin", "Tin", "sn", 50, 3)
	register("antimony", "Antimony", "sb", 51, 4)
	register("tellurium", "Tellurium", "te", 52, 4)
	register("iodine", "Iodine", "i", 53, 6)
	register("xenon", "Xenon", "xe", 54, 7)
	register("cesium", "Cesium", "cs", 55, 0)
	register("barium", "Barium", "ba", 56, 1)
	register("lanthanum", "Lanthanum", "la", 57, 8)
	register("cerium", "Cerium", "ce", 58, 8)
	register("praseodymium", "Praseodymium", "pr", 59, 8)
	register("neodymium", "Neodymium", "nd", 60, 8)
	register("promethium", "Promethium", "pm", 61, 8)
	register("samarium", "Samarium", "sm", 62, 8)
	register("europium", "Europium", "eu", 63, 8)
	register("gadolinium", "Gadolinium", "gd", 64, 8)
	register("terbium", "Terbium", "tb", 65, 8)
	register("dysprosium", "Dysprosium", "dy", 66, 8)
	register("holmium", "Holmium", "ho", 67, 8)
	register("erbium", "Erbium", "er", 68, 8)
	register("thulium", "Thulium", "tm", 69, 8)
	register("ytterbium", "Ytterbium", "yb", 70, 8)
	register("lutetium", "Lutetium", "lu", 71, 8)
	register("hafnium", "Hafnium", "hf", 72, 2)
	register("tantalum", "Tantalum", "ta", 73, 2)
	register("tungsten", "Tungsten", "w", 74, 2)
	register("rhenium", "Rhenium", "re", 75, 2)
	register("osmium", "Osmium", "os", 76, 2)
	register("iridium", "Iridium", "ir", 77, 2)
	register("platinum", "Platinum", "pt", 78, 2)
	register("gold", "Gold", "au", 79, 2)
	register("mercury", "Mercury", "hg", 80, 2)
	register("thallium", "Thallium", "tl", 81, 3)
	register("lead", "Lead", "pb", 82, 3)
	register("bismuth", "Bismuth", "bi", 83, 3)
	register("polonium", "Polonium", "po", 84, 4)
	register("astatine", "Astatine", "at", 85, 6)
	register("radon", "Radon", "rn", 86, 7)
	register("francium", "Francium", "fr", 87, 0)
	register("radium", "Radium", "ra", 88, 1)
	register("actinium", "Actinium", "ac", 89, 9)
	register("thorium", "Thorium", "th", 90, 9)
	register("protactinium", "Protactinium", "pa", 91, 9)
	register("uranium", "Uranium", "u", 92, 9)
	register("neptunium", "Neptunium", "np", 93, 9)
	register("plutonium", "Plutonium", "pu", 94, 9)
	register("americium", "Americium", "am", 95, 9)
	register("curium", "Curium", "cm", 96, 9)
	register("berkelium", "Berkelium", "bk", 97, 9)
	register("californium", "Californium", "cf", 98, 9)
	register("einsteinium", "Einsteinium", "es", 99, 9)
	register("fermium", "Fermium", "fm", 100, 9)
	register("mendelevium", "Mendelevium", "md", 101, 9)
	register("nobelium", "Nobelium", "no", 102, 9)
	register("lawrencium", "Lawrencium", "lr", 103, 9)
	register("rutherfordium", "Rutherfordium", "rf", 104, 2)
	register("dubnium", "Dubnium", "db", 105, 2)
	register("seaborgium", "Seaborgium", "sg", 106, 2)
	register("bohrium", "Bohrium", "bh", 107, 2)
	register("hassium", "Hassium", "hs", 108, 2)
	register("meitnerium", "Meitnerium", "mt", 109, 2)
	register("darmstadtium", "Darmstadtium", "ds", 110, 2)
	register("roentgenium", "Roentgenium", "rg", 111, 2)
	register("copernicium", "Copernicium", "cn", 112, 2)
	register("nihonium", "Nihonium", "nh", 113, 3)
	register("flerovium", "Flerovium", "fl", 114, 3)
	register("moscovium", "Moscovium", "mc", 115, 3)
	register("livermorium", "Livermorium", "lv", 116, 3)
	register("tennessine", "Tennessine", "ts", 117, 6)
	register("oganesson", "Oganesson", "og", 118, 7)

}

func (r *vanillaBlocksRegistry) registerOres() {
	stoneOreBreakInfo := func(toolTier ToolTier) *BlockTypeInfo { return NewBlockTypeInfo(bPickaxe(3.0, toolTier), nil, nil) }
	r.register("coal_ore", func(id *BlockIdentifier) Behavior {
		return NewCoalOre(id, "Coal Ore", stoneOreBreakInfo(vanillaToolTierWood))
	})
	r.register("copper_ore", func(id *BlockIdentifier) Behavior {
		return NewCopperOre(id, "Copper Ore", stoneOreBreakInfo(vanillaToolTierStone))
	})
	r.register("diamond_ore", func(id *BlockIdentifier) Behavior {
		return NewDiamondOre(id, "Diamond Ore", stoneOreBreakInfo(vanillaToolTierIron))
	})
	r.register("emerald_ore", func(id *BlockIdentifier) Behavior {
		return NewEmeraldOre(id, "Emerald Ore", stoneOreBreakInfo(vanillaToolTierIron))
	})
	r.register("gold_ore", func(id *BlockIdentifier) Behavior {
		return NewGoldOre(id, "Gold Ore", stoneOreBreakInfo(vanillaToolTierIron))
	})
	r.register("iron_ore", func(id *BlockIdentifier) Behavior {
		return NewIronOre(id, "Iron Ore", stoneOreBreakInfo(vanillaToolTierStone))
	})
	r.register("lapis_lazuli_ore", func(id *BlockIdentifier) Behavior {
		return NewLapisOre(id, "Lapis Lazuli Ore", stoneOreBreakInfo(vanillaToolTierStone))
	})
	r.register("redstone_ore", func(id *BlockIdentifier) Behavior {
		return NewRedstoneOre(id, "Redstone Ore", stoneOreBreakInfo(vanillaToolTierIron))
	})
	deepslateOreBreakInfo := func(toolTier ToolTier) *BlockTypeInfo {
		return NewBlockTypeInfo(bPickaxe(4.5, toolTier, 15.0), nil, nil)
	}
	r.register("deepslate_coal_ore", func(id *BlockIdentifier) Behavior {
		return NewCoalOre(id, "Deepslate Coal Ore", deepslateOreBreakInfo(vanillaToolTierWood))
	})
	r.register("deepslate_copper_ore", func(id *BlockIdentifier) Behavior {
		return NewCopperOre(id, "Deepslate Copper Ore", deepslateOreBreakInfo(vanillaToolTierStone))
	})
	r.register("deepslate_diamond_ore", func(id *BlockIdentifier) Behavior {
		return NewDiamondOre(id, "Deepslate Diamond Ore", deepslateOreBreakInfo(vanillaToolTierIron))
	})
	r.register("deepslate_emerald_ore", func(id *BlockIdentifier) Behavior {
		return NewEmeraldOre(id, "Deepslate Emerald Ore", deepslateOreBreakInfo(vanillaToolTierIron))
	})
	r.register("deepslate_gold_ore", func(id *BlockIdentifier) Behavior {
		return NewGoldOre(id, "Deepslate Gold Ore", deepslateOreBreakInfo(vanillaToolTierIron))
	})
	r.register("deepslate_iron_ore", func(id *BlockIdentifier) Behavior {
		return NewIronOre(id, "Deepslate Iron Ore", deepslateOreBreakInfo(vanillaToolTierStone))
	})
	r.register("deepslate_lapis_lazuli_ore", func(id *BlockIdentifier) Behavior {
		return NewLapisOre(id, "Deepslate Lapis Lazuli Ore", deepslateOreBreakInfo(vanillaToolTierStone))
	})
	r.register("deepslate_redstone_ore", func(id *BlockIdentifier) Behavior {
		return NewRedstoneOre(id, "Deepslate Redstone Ore", deepslateOreBreakInfo(vanillaToolTierIron))
	})
	netherrackOreBreakInfo := NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierWood), nil, nil)
	r.register("nether_quartz_ore", func(id *BlockIdentifier) Behavior {
		return NewNetherQuartzOre(id, "Nether Quartz Ore", netherrackOreBreakInfo)
	})
	r.register("nether_gold_ore", func(id *BlockIdentifier) Behavior {
		return NewNetherGoldOre(id, "Nether Gold Ore", netherrackOreBreakInfo)
	})
}

func (r *vanillaBlocksRegistry) registerCraftingTables() {
	// //TODO: this is the same for all wooden crafting blocks
	craftingBlockBreakInfo := NewBlockTypeInfo(bAxe(2.5, nil), nil, nil)
	r.register("cartography_table", func(id *BlockIdentifier) Behavior {
		return NewCartographyTable(id, "Cartography Table", craftingBlockBreakInfo)
	})
	r.register("crafting_table", func(id *BlockIdentifier) Behavior {
		return NewCraftingTable(id, "Crafting Table", craftingBlockBreakInfo)
	})
	r.register("fletching_table", func(id *BlockIdentifier) Behavior {
		return NewFletchingTable(id, "Fletching Table", craftingBlockBreakInfo)
	})
	r.register("loom", func(id *BlockIdentifier) Behavior { return NewLoom(id, "Loom", craftingBlockBreakInfo) })
	r.register("smithing_table", func(id *BlockIdentifier) Behavior {
		return NewSmithingTable(id, "Smithing Table", craftingBlockBreakInfo)
	})
}

func (r *vanillaBlocksRegistry) registerChorusBlocks() {
	chorusBlockBreakInfo := NewBlockTypeInfo(bAxe(0.4, nil), nil, nil)
	r.register("chorus_plant", func(id *BlockIdentifier) Behavior { return NewChorusPlant(id, "Chorus Plant", chorusBlockBreakInfo) })
	r.register("chorus_flower", func(id *BlockIdentifier) Behavior { return NewChorusFlower(id, "Chorus Flower", chorusBlockBreakInfo) })
}

func (r *vanillaBlocksRegistry) registerBlocksR13() {
	r.register("light", func(id *BlockIdentifier) Behavior {
		return NewLight(id, "Light Block", NewBlockTypeInfo(bIndestructible(), nil, nil))
	})
	r.register("structure_void", func(id *BlockIdentifier) Behavior {
		return NewStructureVoid(id, "Structure Void", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("wither_rose", func(id *BlockIdentifier) Behavior {
		return NewWitherRose(id, "Wither Rose", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil))
	})
}

func (r *vanillaBlocksRegistry) registerBlocksR14() {
	r.register("honeycomb", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Honeycomb Block", NewBlockTypeInfo(bNew(0.6, ToolTypeNone, 0), nil, nil))
	})
}

func (r *vanillaBlocksRegistry) registerBlocksR16() {
	// //for some reason, slabs have weird hardness like the legacy ones
	slabBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood, 30.0), nil, nil)
	r.register("ancient_debris", func(id *BlockIdentifier) Behavior {
		return newFireProofOpaque(id, "Ancient Debris", NewBlockTypeInfo(bPickaxe(30.0, vanillaToolTierDiamond, 6000.0), nil, nil))
	})
	netheriteBreakInfo := NewBlockTypeInfo(bPickaxe(50.0, vanillaToolTierDiamond, 6000.0), nil, nil)
	r.register("netherite", func(id *BlockIdentifier) Behavior {
		return newFireProofOpaque(id, "Netherite Block", netheriteBreakInfo)
	})
	basaltBreakInfo := NewBlockTypeInfo(bPickaxe(1.25, vanillaToolTierWood, 21.0), nil, nil)
	r.register("basalt", func(id *BlockIdentifier) Behavior { return NewSimplePillar(id, "Basalt", basaltBreakInfo) })
	r.register("polished_basalt", func(id *BlockIdentifier) Behavior { return NewSimplePillar(id, "Polished Basalt", basaltBreakInfo) })
	r.register("smooth_basalt", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Smooth Basalt", basaltBreakInfo) })
	blackstoneBreakInfo := NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("blackstone", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Blackstone", blackstoneBreakInfo) })
	r.register("blackstone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Blackstone", slabBreakInfo) })
	r.register("blackstone_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Blackstone Stairs", blackstoneBreakInfo) })
	r.register("blackstone_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Blackstone Wall", blackstoneBreakInfo) })
	r.register("gilded_blackstone", func(id *BlockIdentifier) Behavior {
		return NewGildedBlackstone(id, "Gilded Blackstone", blackstoneBreakInfo)
	})
	polishedBlackstoneBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood, 30.0), nil, nil)
	r.register("polished_blackstone", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Polished Blackstone", polishedBlackstoneBreakInfo)
	})
	r.register("polished_blackstone_button", func(id *BlockIdentifier) Behavior {
		return NewStoneButton(id, "Polished Blackstone Button", NewBlockTypeInfo(bPickaxe(0.5, nil), nil, nil))
	})
	r.register("polished_blackstone_pressure_plate", func(id *BlockIdentifier) Behavior {
		return NewStonePressurePlate(id, "Polished Blackstone Pressure Plate", NewBlockTypeInfo(bPickaxe(0.5, nil), nil, nil), 20)
	})
	r.register("polished_blackstone_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Polished Blackstone", slabBreakInfo) })
	r.register("polished_blackstone_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Polished Blackstone Stairs", polishedBlackstoneBreakInfo)
	})
	r.register("polished_blackstone_wall", func(id *BlockIdentifier) Behavior {
		return NewWall(id, "Polished Blackstone Wall", polishedBlackstoneBreakInfo)
	})
	r.register("chiseled_polished_blackstone", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Chiseled Polished Blackstone", blackstoneBreakInfo)
	})
	r.register("polished_blackstone_bricks", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Polished Blackstone Bricks", blackstoneBreakInfo)
	})
	r.register("polished_blackstone_brick_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Polished Blackstone Brick", slabBreakInfo) })
	r.register("polished_blackstone_brick_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Polished Blackstone Brick Stairs", blackstoneBreakInfo)
	})
	r.register("polished_blackstone_brick_wall", func(id *BlockIdentifier) Behavior {
		return NewWall(id, "Polished Blackstone Brick Wall", blackstoneBreakInfo)
	})
	r.register("cracked_polished_blackstone_bricks", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Cracked Polished Blackstone Bricks", blackstoneBreakInfo)
	})
	r.register("soul_torch", func(id *BlockIdentifier) Behavior {
		return NewTorch(id, "Soul Torch", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("soul_fire", func(id *BlockIdentifier) Behavior {
		return NewSoulFire(id, "Soul Fire", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsFire}, nil))
	})
	r.register("soul_soil", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Soul Soil", NewBlockTypeInfo(bShovel(0.5, nil), nil, nil))
	})
	r.register("shroomlight", func(id *BlockIdentifier) Behavior {
		return newLightOpaque(id, "Shroomlight", NewBlockTypeInfo(bNew(1.0, ToolTypeHoe, 0), nil, nil), 15)
	})
	r.register("warped_wart_block", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Warped Wart Block", NewBlockTypeInfo(bNew(1.0, ToolTypeHoe, 0), nil, nil))
	})
	r.register("crying_obsidian", func(id *BlockIdentifier) Behavior {
		return newLightOpaque(id, "Crying Obsidian", NewBlockTypeInfo(bPickaxe(35.0 /* 50 in Java */, vanillaToolTierDiamond, 6000.0), nil, nil), 10)
	})
	r.register("twisting_vines", func(id *BlockIdentifier) Behavior {
		return NewNetherVines(id, "Twisting Vines", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil), math.Up)
	})
	r.register("weeping_vines", func(id *BlockIdentifier) Behavior {
		return NewNetherVines(id, "Weeping Vines", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil), math.Down)
	})
	netherRootsInfo := NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil)
	r.register("crimson_roots", func(id *BlockIdentifier) Behavior { return NewNetherRoots(id, "Crimson Roots", netherRootsInfo) })
	r.register("warped_roots", func(id *BlockIdentifier) Behavior { return NewNetherRoots(id, "Warped Roots", netherRootsInfo) })
	r.register("chain", func(id *BlockIdentifier) Behavior {
		return NewChain(id, "Chain", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierWood, 30.0), nil, nil))
	})
	r.register("copper_chain", func(id *BlockIdentifier) Behavior {
		return NewCopperChain(id, "Copper Chain", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierWood, 30.0), nil, nil))
	})
	r.register("respawn_anchor", func(id *BlockIdentifier) Behavior {
		return NewRespawnAnchor(id, "Respawn Anchor", NewBlockTypeInfo(bPickaxe(50.0, vanillaToolTierDiamond, 6000.0), nil, nil))
	})
	netherFungusInfo := NewBlockTypeInfo(bInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants, BlockTypeTagsHugeFungusReplaceable}, nil)
	r.register("crimson_fungus", func(id *BlockIdentifier) Behavior {
		return NewNetherFungus(id, "Crimson Fungus", netherFungusInfo, TreeTypeCrimson, CRIMSON_NYLIUM)
	})
	r.register("warped_fungus", func(id *BlockIdentifier) Behavior {
		return NewNetherFungus(id, "Warped Fungus", netherFungusInfo, TreeTypeWarped, WARPED_NYLIUM)
	})
	r.register("nether_sprouts", func(id *BlockIdentifier) Behavior {
		return NewNetherSprouts(id, "Nether Sprouts", NewBlockTypeInfo(bInstant(ToolTypeShears, 1), nil, nil))
	})
	nyliumBreakInfo := NewBlockTypeInfo(bPickaxe(0.4, vanillaToolTierWood), []string{BlockTypeTagsNylium}, nil)
	r.register("crimson_nylium", func(id *BlockIdentifier) Behavior {
		return NewNylium(id, "Crimson Nylium", nyliumBreakInfo, []Behavior{r.get("crimson_fungus"), r.get("crimson_roots")})
	})
	r.register("warped_nylium", func(id *BlockIdentifier) Behavior {
		return NewNylium(id, "Warped Nylium", nyliumBreakInfo, []Behavior{r.get("warped_fungus"), r.get("warped_roots"), r.get("nether_sprouts")})
	})
}

func (r *vanillaBlocksRegistry) registerBlocksR17() {
	// //in java this can be acquired using any tool - seems to be a parity issue in bedrock
	amethystInfo := NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood), nil, nil)
	r.register("amethyst", func(id *BlockIdentifier) Behavior { return NewAmethyst(id, "Amethyst", amethystInfo) })
	r.register("budding_amethyst", func(id *BlockIdentifier) Behavior { return NewBuddingAmethyst(id, "Budding Amethyst", amethystInfo) })
	r.register("amethyst_cluster", func(id *BlockIdentifier) Behavior { return NewAmethystCluster(id, "Amethyst Cluster", amethystInfo) })
	r.register("calcite", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Calcite", NewBlockTypeInfo(bPickaxe(0.75, vanillaToolTierWood), nil, nil))
	})
	r.register("raw_copper", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Raw Copper Block", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierStone, 30.0), nil, nil))
	})
	r.register("raw_gold", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Raw Gold Block", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierIron, 30.0), nil, nil))
	})
	r.register("raw_iron", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Raw Iron Block", NewBlockTypeInfo(bPickaxe(5.0, vanillaToolTierStone, 30.0), nil, nil))
	})
	deepslateBreakInfo := NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierWood, 30.0), nil, nil)
	deepslate := r.register("deepslate", func(id *BlockIdentifier) Behavior { return NewDeepslate(id, "Deepslate", deepslateBreakInfo) })
	// //TODO: parity issue here - in Java this has a hardness of 3.0, but in bedrock it's 3.5
	r.register("chiseled_deepslate", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Chiseled Deepslate", NewBlockTypeInfo(bPickaxe(3.5, vanillaToolTierWood, 30.0), nil, nil))
	})
	deepslateBrickBreakInfo := NewBlockTypeInfo(bPickaxe(3.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("deepslate_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Deepslate Bricks", deepslateBrickBreakInfo) })
	r.register("deepslate_brick_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Deepslate Brick", deepslateBrickBreakInfo) })
	r.register("deepslate_brick_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Deepslate Brick Stairs", deepslateBrickBreakInfo)
	})
	r.register("deepslate_brick_wall", func(id *BlockIdentifier) Behavior {
		return NewWall(id, "Deepslate Brick Wall", deepslateBrickBreakInfo)
	})
	r.register("cracked_deepslate_bricks", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Cracked Deepslate Bricks", deepslateBrickBreakInfo)
	})
	deepslateTilesBreakInfo := NewBlockTypeInfo(bPickaxe(3.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("deepslate_tiles", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Deepslate Tiles", deepslateTilesBreakInfo) })
	r.register("deepslate_tile_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Deepslate Tile", deepslateTilesBreakInfo) })
	r.register("deepslate_tile_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Deepslate Tile Stairs", deepslateTilesBreakInfo)
	})
	r.register("deepslate_tile_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Deepslate Tile Wall", deepslateTilesBreakInfo) })
	r.register("cracked_deepslate_tiles", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Cracked Deepslate Tiles", deepslateTilesBreakInfo)
	})
	cobbledDeepslateBreakInfo := NewBlockTypeInfo(bPickaxe(3.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("cobbled_deepslate", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Cobbled Deepslate", cobbledDeepslateBreakInfo)
	})
	r.register("cobbled_deepslate_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Cobbled Deepslate", cobbledDeepslateBreakInfo) })
	r.register("cobbled_deepslate_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Cobbled Deepslate Stairs", cobbledDeepslateBreakInfo)
	})
	r.register("cobbled_deepslate_wall", func(id *BlockIdentifier) Behavior {
		return NewWall(id, "Cobbled Deepslate Wall", cobbledDeepslateBreakInfo)
	})
	polishedDeepslateBreakInfo := NewBlockTypeInfo(bPickaxe(3.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("polished_deepslate", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Polished Deepslate", polishedDeepslateBreakInfo)
	})
	r.register("polished_deepslate_slab", func(id *BlockIdentifier) Behavior {
		return NewSlab(id, "Polished Deepslate", polishedDeepslateBreakInfo)
	})
	r.register("polished_deepslate_stairs", func(id *BlockIdentifier) Behavior {
		return NewStair(id, "Polished Deepslate Stairs", polishedDeepslateBreakInfo)
	})
	r.register("polished_deepslate_wall", func(id *BlockIdentifier) Behavior {
		return NewWall(id, "Polished Deepslate Wall", polishedDeepslateBreakInfo)
	})
	r.register("tinted_glass", func(id *BlockIdentifier) Behavior {
		return NewTintedGlass(id, "Tinted Glass", NewBlockTypeInfo(bNew(0.3, ToolTypeNone, 0), nil, nil))
	})
	copperBreakInfo := NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierStone, 30.0), nil, nil)
	r.register("lightning_rod", func(id *BlockIdentifier) Behavior { return NewLightningRod(id, "Lightning Rod", copperBreakInfo) })
	r.register("copper", func(id *BlockIdentifier) Behavior { return NewCopper(id, "Copper Block", copperBreakInfo) })
	r.register("chiseled_copper", func(id *BlockIdentifier) Behavior { return NewCopper(id, "Chiseled Copper", copperBreakInfo) })
	r.register("copper_grate", func(id *BlockIdentifier) Behavior { return NewCopperGrate(id, "Copper Grate", copperBreakInfo) })
	r.register("cut_copper", func(id *BlockIdentifier) Behavior { return NewCopper(id, "Cut Copper Block", copperBreakInfo) })
	r.register("cut_copper_slab", func(id *BlockIdentifier) Behavior { return NewCopperSlab(id, "Cut Copper Slab", copperBreakInfo) })
	r.register("cut_copper_stairs", func(id *BlockIdentifier) Behavior { return NewCopperStairs(id, "Cut Copper Stairs", copperBreakInfo) })
	r.register("copper_bulb", func(id *BlockIdentifier) Behavior { return NewCopperBulb(id, "Copper Bulb", copperBreakInfo) })
	r.register("copper_door", func(id *BlockIdentifier) Behavior {
		return NewCopperDoor(id, "Copper Door", NewBlockTypeInfo(bPickaxe(3.0, nil, 30.0), nil, nil))
	})
	r.register("copper_trapdoor", func(id *BlockIdentifier) Behavior {
		return NewCopperTrapdoor(id, "Copper Trapdoor", NewBlockTypeInfo(bPickaxe(3.0, vanillaToolTierStone, 30.0), nil, nil))
	})
	candleBreakInfo := NewBlockTypeInfo(bNew(0.1, ToolTypeNone, 0), nil, nil)
	r.register("candle", func(id *BlockIdentifier) Behavior { return NewCandle(id, "Candle", candleBreakInfo) })
	r.register("dyed_candle", func(id *BlockIdentifier) Behavior { return NewDyedCandle(id, "Dyed Candle", candleBreakInfo) })
	// //TODO: duplicated break info :(
	cakeBreakInfo := NewBlockTypeInfo(bNew(0.5, ToolTypeNone, 0), nil, nil)
	r.register("cake_with_candle", func(id *BlockIdentifier) Behavior { return NewCakeWithCandle(id, "Cake With Candle", cakeBreakInfo) })
	r.register("cake_with_dyed_candle", func(id *BlockIdentifier) Behavior {
		return NewCakeWithDyedCandle(id, "Cake With Dyed Candle", cakeBreakInfo)
	})
	r.register("hanging_roots", func(id *BlockIdentifier) Behavior {
		return NewHangingRoots(id, "Hanging Roots", NewBlockTypeInfo(bInstant(ToolTypeShears, 1), nil, nil))
	})
	r.register("cave_vines", func(id *BlockIdentifier) Behavior {
		return NewCaveVines(id, "Cave Vines", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("small_dripleaf", func(id *BlockIdentifier) Behavior {
		return NewSmallDripleaf(id, "Small Dripleaf", NewBlockTypeInfo(bInstant(ToolTypeShears, 1), nil, nil))
	})
	r.register("big_dripleaf_head", func(id *BlockIdentifier) Behavior {
		return NewBigDripleafHead(id, "Big Dripleaf", NewBlockTypeInfo(bNew(0.1, ToolTypeNone, 0), nil, nil))
	})
	r.register("big_dripleaf_stem", func(id *BlockIdentifier) Behavior {
		return NewBigDripleafStem(id, "Big Dripleaf Stem", NewBlockTypeInfo(bNew(0.1, ToolTypeNone, 0), nil, nil))
	})
	r.register("infested_deepslate", func(id *BlockIdentifier) Behavior {
		return NewInfestedPillar(id, "Infested Deepslate", NewBlockTypeInfo(bPickaxe(1.5, nil, 3.75), nil, nil), deepslate)
	})
}

func (r *vanillaBlocksRegistry) registerBlocksR18() {
	r.register("spore_blossom", func(id *BlockIdentifier) Behavior {
		return NewSporeBlossom(id, "Spore Blossom", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
}

func (r *vanillaBlocksRegistry) registerMudBlocks() {
	r.register("mud", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Mud", NewBlockTypeInfo(bShovel(0.5, nil), []string{BlockTypeTagsMud}, nil))
	})
	r.register("packed_mud", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Packed Mud", NewBlockTypeInfo(bPickaxe(1.0, nil, 15.0), nil, nil))
	})
	mudBricksBreakInfo := NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood, 15.0), nil, nil)
	r.register("mud_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Mud Bricks", mudBricksBreakInfo) })
	r.register("mud_brick_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Mud Brick", mudBricksBreakInfo) })
	r.register("mud_brick_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Mud Brick Stairs", mudBricksBreakInfo) })
	r.register("mud_brick_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Mud Brick Wall", mudBricksBreakInfo) })
}

func (r *vanillaBlocksRegistry) registerResinBlocks() {
	r.register("resin", func(id *BlockIdentifier) Behavior {
		return newOpaque(id, "Block of Resin", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	r.register("resin_clump", func(id *BlockIdentifier) Behavior {
		return NewResinClump(id, "Resin Clump", NewBlockTypeInfo(bInstant(ToolTypeNone, 0), nil, nil))
	})
	resinBricksInfo := NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("resin_brick_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Resin Brick", resinBricksInfo) })
	r.register("resin_brick_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Resin Brick Stairs", resinBricksInfo) })
	r.register("resin_brick_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Resin Brick Wall", resinBricksInfo) })
	r.register("resin_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Resin Bricks", resinBricksInfo) })
	r.register("chiseled_resin_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Chiseled Resin Bricks", resinBricksInfo) })
}

func (r *vanillaBlocksRegistry) registerTuffBlocks() {
	tuffBreakInfo := NewBlockTypeInfo(bPickaxe(1.5, vanillaToolTierWood, 30.0), nil, nil)
	r.register("tuff", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Tuff", tuffBreakInfo) })
	r.register("tuff_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Tuff", tuffBreakInfo) })
	r.register("tuff_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Tuff Stairs", tuffBreakInfo) })
	r.register("tuff_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Tuff Wall", tuffBreakInfo) })
	r.register("chiseled_tuff", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Chiseled Tuff", tuffBreakInfo) })
	r.register("tuff_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Tuff Bricks", tuffBreakInfo) })
	r.register("tuff_brick_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Tuff Brick", tuffBreakInfo) })
	r.register("tuff_brick_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Tuff Brick Stairs", tuffBreakInfo) })
	r.register("tuff_brick_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Tuff Brick Wall", tuffBreakInfo) })
	r.register("chiseled_tuff_bricks", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Chiseled Tuff Bricks", tuffBreakInfo) })
	r.register("polished_tuff", func(id *BlockIdentifier) Behavior { return newOpaque(id, "Polished Tuff", tuffBreakInfo) })
	r.register("polished_tuff_slab", func(id *BlockIdentifier) Behavior { return NewSlab(id, "Polished Tuff", tuffBreakInfo) })
	r.register("polished_tuff_stairs", func(id *BlockIdentifier) Behavior { return NewStair(id, "Polished Tuff Stairs", tuffBreakInfo) })
	r.register("polished_tuff_wall", func(id *BlockIdentifier) Behavior { return NewWall(id, "Polished Tuff Wall", tuffBreakInfo) })
}

func (r *vanillaBlocksRegistry) registerCauldronBlocks() {
	cauldronBreakInfo := NewBlockTypeInfo(bPickaxe(2.0, vanillaToolTierWood), nil, nil)
	r.register("cauldron", func(id *BlockIdentifier) Behavior { return NewCauldron(id, "Cauldron", cauldronBreakInfo) }, tileCauldron)
	r.register("water_cauldron", func(id *BlockIdentifier) Behavior { return NewWaterCauldron(id, "Water Cauldron", cauldronBreakInfo) }, tileCauldron)
	r.register("lava_cauldron", func(id *BlockIdentifier) Behavior { return NewLavaCauldron(id, "Lava Cauldron", cauldronBreakInfo) }, tileCauldron)
	r.register("potion_cauldron", func(id *BlockIdentifier) Behavior { return NewPotionCauldron(id, "Potion Cauldron", cauldronBreakInfo) }, tileCauldron)
}

// VanillaAir is VanillaBlocks::AIR().
func VanillaAir() Behavior { return VanillaBlock("air") }

// VanillaDirt is VanillaBlocks::DIRT().
func VanillaDirt() Behavior { return VanillaBlock("dirt") }

// VanillaWater is VanillaBlocks::WATER().
func VanillaWater() Behavior { return VanillaBlock("water") }

// VanillaNetherrack is VanillaBlocks::NETHERRACK().
func VanillaNetherrack() Behavior { return VanillaBlock("netherrack") }

// VanillaObsidian is VanillaBlocks::OBSIDIAN().
func VanillaObsidian() Behavior { return VanillaBlock("obsidian") }

// VanillaSoulSoil is VanillaBlocks::SOUL_SOIL().
func VanillaSoulSoil() Behavior { return VanillaBlock("soul_soil") }

// VanillaCake is VanillaBlocks::CAKE().
func VanillaCake() Behavior { return VanillaBlock("cake") }

// VanillaConcrete is VanillaBlocks::CONCRETE().
func VanillaConcrete() Behavior { return VanillaBlock("concrete") }

// VanillaMelon is VanillaBlocks::MELON().
func VanillaMelon() Behavior { return VanillaBlock("melon") }

// VanillaPumpkin is VanillaBlocks::PUMPKIN().
func VanillaPumpkin() Behavior { return VanillaBlock("pumpkin") }

// VanillaSugarcane is VanillaBlocks::SUGARCANE().
func VanillaSugarcane() Behavior { return VanillaBlock("sugarcane") }

// VanillaCaveVines is VanillaBlocks::CAVE_VINES().
func VanillaCaveVines() Behavior { return VanillaBlock("cave_vines") }

// VanillaTorchflower is VanillaBlocks::TORCHFLOWER().
func VanillaTorchflower() Behavior { return VanillaBlock("torchflower") }

// VanillaTorchflowerCrop is VanillaBlocks::TORCHFLOWER_CROP().
func VanillaTorchflowerCrop() Behavior { return VanillaBlock("torchflower_crop") }

// VanillaCrimsonNylium is VanillaBlocks::CRIMSON_NYLIUM().
func VanillaCrimsonNylium() Behavior { return VanillaBlock("crimson_nylium") }

// VanillaWarpedNylium is VanillaBlocks::WARPED_NYLIUM().
func VanillaWarpedNylium() Behavior { return VanillaBlock("warped_nylium") }

// VanillaAmethystCluster is VanillaBlocks::AMETHYST_CLUSTER().
func VanillaAmethystCluster() Behavior { return VanillaBlock("amethyst_cluster") }

// VanillaCobblestone is VanillaBlocks::COBBLESTONE().
func VanillaCobblestone() Behavior { return VanillaBlock("cobblestone") }

// VanillaBasalt is VanillaBlocks::BASALT().
func VanillaBasalt() Behavior { return VanillaBlock("basalt") }

// VanillaGrass is VanillaBlocks::GRASS().
func VanillaGrass() Behavior { return VanillaBlock("grass") }

// VanillaMycelium is VanillaBlocks::MYCELIUM().
func VanillaMycelium() Behavior { return VanillaBlock("mycelium") }

// VanillaCactus is VanillaBlocks::CACTUS().
func VanillaCactus() Behavior { return VanillaBlock("cactus") }

// VanillaCactusFlower is VanillaBlocks::CACTUS_FLOWER().
func VanillaCactusFlower() Behavior { return VanillaBlock("cactus_flower") }

// VanillaOminousBanner is VanillaBlocks::OMINOUS_BANNER().
func VanillaOminousBanner() Behavior { return VanillaBlock("ominous_banner") }

// VanillaOminousWallBanner is VanillaBlocks::OMINOUS_WALL_BANNER().
func VanillaOminousWallBanner() Behavior { return VanillaBlock("ominous_wall_banner") }

// VanillaBamboo is VanillaBlocks::BAMBOO().
func VanillaBamboo() Behavior { return VanillaBlock("bamboo") }

// VanillaChorusPlant is VanillaBlocks::CHORUS_PLANT().
func VanillaChorusPlant() Behavior { return VanillaBlock("chorus_plant") }

// VanillaDoublePitcherCrop is VanillaBlocks::DOUBLE_PITCHER_CROP().
func VanillaDoublePitcherCrop() Behavior { return VanillaBlock("double_pitcher_crop") }

// VanillaBigDripleafStem is VanillaBlocks::BIG_DRIPLEAF_STEM().
func VanillaBigDripleafStem() Behavior { return VanillaBlock("big_dripleaf_stem") }

// VanillaSmallDripleaf is VanillaBlocks::SMALL_DRIPLEAF().
func VanillaSmallDripleaf() Behavior { return VanillaBlock("small_dripleaf") }

// VanillaBigDripleafHead is VanillaBlocks::BIG_DRIPLEAF_HEAD().
func VanillaBigDripleafHead() Behavior { return VanillaBlock("big_dripleaf_head") }

// VanillaStone is VanillaBlocks::STONE().
func VanillaStone() Behavior { return VanillaBlock("stone") }

// VanillaBedrock is VanillaBlocks::BEDROCK().
func VanillaBedrock() Behavior { return VanillaBlock("bedrock") }

// VanillaTallGrass is VanillaBlocks::TALL_GRASS().
func VanillaTallGrass() Behavior { return VanillaBlock("tall_grass") }

// VanillaGravel is VanillaBlocks::GRAVEL().
func VanillaGravel() Behavior { return VanillaBlock("gravel") }

// VanillaCoalOre is VanillaBlocks::COAL_ORE().
func VanillaCoalOre() Behavior { return VanillaBlock("coal_ore") }

// VanillaDiamondOre is VanillaBlocks::DIAMOND_ORE().
func VanillaDiamondOre() Behavior { return VanillaBlock("diamond_ore") }

// VanillaGoldOre is VanillaBlocks::GOLD_ORE().
func VanillaGoldOre() Behavior { return VanillaBlock("gold_ore") }

// VanillaIronOre is VanillaBlocks::IRON_ORE().
func VanillaIronOre() Behavior { return VanillaBlock("iron_ore") }

// VanillaLapisLazuliOre is VanillaBlocks::LAPIS_LAZULI_ORE().
func VanillaLapisLazuliOre() Behavior { return VanillaBlock("lapis_lazuli_ore") }

// VanillaRedstoneOre is VanillaBlocks::REDSTONE_ORE().
func VanillaRedstoneOre() Behavior { return VanillaBlock("redstone_ore") }

// VanillaEmeraldOre is VanillaBlocks::EMERALD_ORE().
func VanillaEmeraldOre() Behavior { return VanillaBlock("emerald_ore") }

// VanillaSand is VanillaBlocks::SAND().
func VanillaSand() Behavior { return VanillaBlock("sand") }

// VanillaSandstone is VanillaBlocks::SANDSTONE().
func VanillaSandstone() Behavior { return VanillaBlock("sandstone") }

// VanillaSnowLayer is VanillaBlocks::SNOW_LAYER().
func VanillaSnowLayer() Behavior { return VanillaBlock("snow_layer") }

// VanillaOakLog is VanillaBlocks::OAK_LOG().
func VanillaOakLog() Behavior { return VanillaBlock("oak_log") }

// VanillaOakLeaves is VanillaBlocks::OAK_LEAVES().
func VanillaOakLeaves() Behavior { return VanillaBlock("oak_leaves") }

// VanillaSpruceLog is VanillaBlocks::SPRUCE_LOG().
func VanillaSpruceLog() Behavior { return VanillaBlock("spruce_log") }

// VanillaSpruceLeaves is VanillaBlocks::SPRUCE_LEAVES().
func VanillaSpruceLeaves() Behavior { return VanillaBlock("spruce_leaves") }

// VanillaBirchLog is VanillaBlocks::BIRCH_LOG().
func VanillaBirchLog() Behavior { return VanillaBlock("birch_log") }

// VanillaBirchLeaves is VanillaBlocks::BIRCH_LEAVES().
func VanillaBirchLeaves() Behavior { return VanillaBlock("birch_leaves") }

// VanillaFire is VanillaBlocks::FIRE().
func VanillaFire() Behavior { return VanillaBlock("fire") }

// VanillaTNT is VanillaBlocks::TNT().
func VanillaTNT() Behavior { return VanillaBlock("tnt") }

// VanillaLava is VanillaBlocks::LAVA().
func VanillaLava() Behavior { return VanillaBlock("lava") }

// VanillaNetherQuartzOre is VanillaBlocks::NETHER_QUARTZ_ORE().
func VanillaNetherQuartzOre() Behavior { return VanillaBlock("nether_quartz_ore") }

// VanillaIce is VanillaBlocks::ICE().
func VanillaIce() Behavior { return VanillaBlock("ice") }

// VanillaFrostedIce is VanillaBlocks::FROSTED_ICE().
func VanillaFrostedIce() Behavior { return VanillaBlock("frosted_ice") }

// VanillaOakPlanks is VanillaBlocks::OAK_PLANKS().
func VanillaOakPlanks() Behavior { return VanillaBlock("oak_planks") }
