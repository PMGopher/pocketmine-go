package block

import (
	"fmt"
	"sync"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

// VanillaBlocks is a port of a small slice of pocketmine\block\VanillaBlocks (itself generated
// from VanillaBlocksInputs.php, ~1100 source lines / ~5700 generated lines - porting it in full is
// its own multi-session undertaking). This starts with just the handful of singletons already
// needed by gap comments scattered across this package (Farmland/GrassPath/Ice/FrostedIce's block-
// swap-on-decay logic) - populated incrementally as more blocks need one. Each getter matches the
// real generated getter's shape: lazily construct a cached singleton on first use, then return a
// Clone() of it every call (blocks are mutable per-instance, so callers must never share the
// singleton itself).
//
// Real BreakInfo/tag configuration for each entry is copied from VanillaBlocksInputs.php, not
// guessed - see each var's comment for the exact source line.
var (
	vanillaAir               Behavior
	vanillaDirt              Behavior
	vanillaWater             Behavior
	vanillaNetherrack        Behavior
	vanillaObsidian          Behavior
	vanillaSoulSoil          Behavior
	vanillaCake              Behavior
	vanillaConcrete          Behavior
	vanillaMelon             Behavior
	vanillaPumpkin           Behavior
	vanillaSugarcane         Behavior
	vanillaCaveVines         Behavior
	vanillaTorchflower       Behavior
	vanillaTorchflowerCrop   Behavior
	vanillaCrimsonNylium     Behavior
	vanillaWarpedNylium      Behavior
	vanillaAmethystCluster   Behavior
	vanillaCobblestone       Behavior
	vanillaBasalt            Behavior
	vanillaGrass             Behavior
	vanillaMycelium          Behavior
	vanillaCactus            Behavior
	vanillaCactusFlower      Behavior
	vanillaOminousBanner     Behavior
	vanillaOminousWallBanner Behavior
	vanillaBamboo            Behavior
	vanillaChorusPlant       Behavior
	vanillaDoublePitcherCrop Behavior
	vanillaBigDripleafStem   Behavior
	vanillaSmallDripleaf     Behavior
	vanillaBigDripleafHead   Behavior
	vanillaStone             Behavior
	vanillaBedrock           Behavior
	vanillaTallGrass         Behavior
	vanillaGravel            Behavior
	vanillaCoalOre           Behavior
	vanillaDiamondOre        Behavior
	vanillaGoldOre           Behavior
	vanillaIronOre           Behavior
	vanillaLapisLazuliOre    Behavior
	vanillaRedstoneOre       Behavior
	vanillaSand              Behavior
	vanillaSandstone         Behavior
	vanillaSnowLayer         Behavior
	vanillaEmeraldOre        Behavior
	vanillaOakLog            Behavior
	vanillaOakLeaves         Behavior
	vanillaSpruceLog         Behavior
	vanillaSpruceLeaves      Behavior
	vanillaBirchLog          Behavior
	vanillaBirchLeaves       Behavior
	vanillaFire              Behavior
	vanillaTNT               Behavior
	vanillaIce               Behavior
	vanillaFrostedIce        Behavior
	vanillaOakPlanks         Behavior
	vanillaLava              Behavior
	vanillaNetherQuartzOre   Behavior
)

// vanillaHarvestLevel is a minimal block.ToolTier implementation (GetHarvestLevel only) for use
// within this file - the real per-tier metadata (item.ToolTier) lives in the item package, which
// this package can't import (item already imports block). The harvest-level numbers themselves
// are copied from item.ToolTier's real table (WOOD=1, DIAMOND=5, etc.), not guessed.
type vanillaHarvestLevel int

func (h vanillaHarvestLevel) GetHarvestLevel() int { return int(h) }

const (
	vanillaToolTierWood    vanillaHarvestLevel = 1
	vanillaToolTierStone   vanillaHarvestLevel = 3
	vanillaToolTierIron    vanillaHarvestLevel = 4
	vanillaToolTierDiamond vanillaHarvestLevel = 5
)

func mustVanillaBlockIdentifier(blockTypeID int) *BlockIdentifier {
	idInfo, err := NewBlockIdentifier(blockTypeID, nil)
	if err != nil {
		panic(fmt.Sprintf("VanillaBlocks: invalid block type ID %d: %v", blockTypeID, err))
	}
	return idInfo
}

// VanillaAir is a port of VanillaBlocks::AIR() - see VanillaBlocksInputs.php's
// register("air", ...): BreakInfo::indestructible(-1.0).
var vanillaAirOnce sync.Once

func VanillaAir() Behavior {
	vanillaAirOnce.Do(func() {
		vanillaAir = NewAir(mustVanillaBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoIndestructible(-1.0), nil, nil))
	})
	return vanillaAir.Clone()
}

// VanillaDirt is a port of VanillaBlocks::DIRT() - see VanillaBlocksInputs.php's
// register("dirt", ...): BreakInfo::shovel(0.5), [Tags::DIRT].
var vanillaDirtOnce sync.Once

func VanillaDirt() Behavior {
	vanillaDirtOnce.Do(func() {
		vanillaDirt = NewDirt(mustVanillaBlockIdentifier(DIRT), "Dirt", NewBlockTypeInfo(BlockBreakInfoShovel(0.5, nil, nil), []string{BlockTypeTagsDirt}, nil))
	})
	return vanillaDirt.Clone()
}

// VanillaWater is a port of VanillaBlocks::WATER() - see VanillaBlocksInputs.php's
// register("water", ...): BreakInfo::indestructible(500.0).
var vanillaWaterOnce sync.Once

func VanillaWater() Behavior {
	vanillaWaterOnce.Do(func() {
		vanillaWater = NewWater(mustVanillaBlockIdentifier(WATER), "Water", NewBlockTypeInfo(BlockBreakInfoIndestructible(500.0), nil, nil))
	})
	return vanillaWater.Clone()
}

// VanillaNetherrack is a port of VanillaBlocks::NETHERRACK() - see VanillaBlocksInputs.php's
// register("netherrack", ...): BreakInfo::pickaxe(0.4, ToolTier::WOOD).
var vanillaNetherrackOnce sync.Once

func VanillaNetherrack() Behavior {
	vanillaNetherrackOnce.Do(func() {
		vanillaNetherrack = NewNetherrack(mustVanillaBlockIdentifier(NETHERRACK), "Netherrack", NewBlockTypeInfo(BlockBreakInfoPickaxe(0.4, vanillaToolTierWood, nil), nil, nil))
	})
	return vanillaNetherrack.Clone()
}

// VanillaObsidian is a port of VanillaBlocks::OBSIDIAN() - see VanillaBlocksInputs.php's
// register("obsidian", ...): BreakInfo::pickaxe(35.0, ToolTier::DIAMOND, 6000.0).
var vanillaObsidianOnce sync.Once

func VanillaObsidian() Behavior {
	vanillaObsidianOnce.Do(func() {
		blastResistance := 6000.0
		obsidian := &Opaque{Block: NewBlock(mustVanillaBlockIdentifier(OBSIDIAN), "Obsidian", NewBlockTypeInfo(BlockBreakInfoPickaxe(35.0, vanillaToolTierDiamond, &blastResistance), nil, nil))}
		obsidian.Init(obsidian)
		vanillaObsidian = obsidian
	})
	return vanillaObsidian.Clone()
}

// VanillaSoulSoil is a port of VanillaBlocks::SOUL_SOIL() - see VanillaBlocksInputs.php's
// register("soul_soil", ...): BreakInfo::shovel(0.5).
var vanillaSoulSoilOnce sync.Once

func VanillaSoulSoil() Behavior {
	vanillaSoulSoilOnce.Do(func() {
		soil := &Opaque{Block: NewBlock(mustVanillaBlockIdentifier(SOUL_SOIL), "Soul Soil", NewBlockTypeInfo(BlockBreakInfoShovel(0.5, nil, nil), nil, nil))}
		soil.Init(soil)
		vanillaSoulSoil = soil
	})
	return vanillaSoulSoil.Clone()
}

// VanillaCake is a port of VanillaBlocks::CAKE() - see VanillaBlocksInputs.php's
// register("cake", ...): new BreakInfo(0.5) (no tool type, no tier).
var vanillaCakeOnce sync.Once

func VanillaCake() Behavior {
	vanillaCakeOnce.Do(func() {
		vanillaCake = NewCake(mustVanillaBlockIdentifier(CAKE), "Cake", NewBlockTypeInfo(NewBlockBreakInfo(0.5, ToolTypeNone, 0, nil, nil), nil, nil))
	})
	return vanillaCake.Clone()
}

// VanillaConcrete is a port of VanillaBlocks::CONCRETE() - see VanillaBlocksInputs.php's
// register("concrete", ...): BreakInfo::pickaxe(1.8, ToolTier::WOOD). Defaults to White, same as
// the real singleton - callers needing a specific color call SetColor on the result, same pattern
// as VanillaCake()/CakeWithCandle.GetResidue.
var vanillaConcreteOnce sync.Once

func VanillaConcrete() Behavior {
	vanillaConcreteOnce.Do(func() {
		vanillaConcrete = NewConcrete(mustVanillaBlockIdentifier(CONCRETE), "Concrete", NewBlockTypeInfo(BlockBreakInfoPickaxe(1.8, vanillaToolTierWood, nil), nil, nil))
	})
	return vanillaConcrete.Clone()
}

// VanillaMelon is a port of VanillaBlocks::MELON() - see VanillaBlocksInputs.php's
// register("melon", ...): BreakInfo::axe(1.0).
var vanillaMelonOnce sync.Once

func VanillaMelon() Behavior {
	vanillaMelonOnce.Do(func() {
		vanillaMelon = NewMelon(mustVanillaBlockIdentifier(MELON), "Melon Block", NewBlockTypeInfo(BlockBreakInfoAxe(1.0, nil, nil), nil, nil))
	})
	return vanillaMelon.Clone()
}

// VanillaPumpkin is a port of VanillaBlocks::PUMPKIN() - see VanillaBlocksInputs.php's
// register("pumpkin", ...): BreakInfo::axe(1.0).
var vanillaPumpkinOnce sync.Once

func VanillaPumpkin() Behavior {
	vanillaPumpkinOnce.Do(func() {
		vanillaPumpkin = NewPumpkin(mustVanillaBlockIdentifier(PUMPKIN), "Pumpkin", NewBlockTypeInfo(BlockBreakInfoAxe(1.0, nil, nil), nil, nil))
	})
	return vanillaPumpkin.Clone()
}

// VanillaSugarcane is a port of VanillaBlocks::SUGARCANE() - see VanillaBlocksInputs.php's
// register("sugarcane", ...): BreakInfo::instant().
var vanillaSugarcaneOnce sync.Once

func VanillaSugarcane() Behavior {
	vanillaSugarcaneOnce.Do(func() {
		vanillaSugarcane = NewSugarcane(mustVanillaBlockIdentifier(SUGARCANE), "Sugarcane", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	})
	return vanillaSugarcane.Clone()
}

// VanillaCaveVines is a port of VanillaBlocks::CAVE_VINES() - see VanillaBlocksInputs.php's
// register("cave_vines", ...): BreakInfo::instant().
var vanillaCaveVinesOnce sync.Once

func VanillaCaveVines() Behavior {
	vanillaCaveVinesOnce.Do(func() {
		vanillaCaveVines = NewCaveVines(mustVanillaBlockIdentifier(CAVE_VINES), "Cave Vines", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	})
	return vanillaCaveVines.Clone()
}

// VanillaTorchflower is a port of VanillaBlocks::TORCHFLOWER() - see VanillaBlocksInputs.php's
// register("torchflower", ...): BreakInfo::instant(), [Tags::POTTABLE_PLANTS] ($flowerTypeInfo).
var vanillaTorchflowerOnce sync.Once

func VanillaTorchflower() Behavior {
	vanillaTorchflowerOnce.Do(func() {
		vanillaTorchflower = NewFlower(mustVanillaBlockIdentifier(TORCHFLOWER), "Torchflower", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil))
	})
	return vanillaTorchflower.Clone()
}

// VanillaTorchflowerCrop is a port of VanillaBlocks::TORCHFLOWER_CROP() - see
// VanillaBlocksInputs.php's register("torchflower_crop", ...): BreakInfo::instant().
var vanillaTorchflowerCropOnce sync.Once

func VanillaTorchflowerCrop() Behavior {
	vanillaTorchflowerCropOnce.Do(func() {
		vanillaTorchflowerCrop = NewTorchflowerCrop(mustVanillaBlockIdentifier(TORCHFLOWER_CROP), "Torchflower Crop", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	})
	return vanillaTorchflowerCrop.Clone()
}

// VanillaCrimsonNylium is a port of VanillaBlocks::CRIMSON_NYLIUM() - see VanillaBlocksInputs.php's
// registerDelayed("crimson_nylium", ...): BreakInfo::pickaxe(0.4, ToolTier::WOOD), [Tags::NYLIUM].
// Vegetation is nil (the real singleton lists CRIMSON_FUNGUS/CRIMSON_ROOTS, neither ported yet) -
// harmless for Netherrack.tryTransform, the only current caller, which never reads it.
var vanillaCrimsonNyliumOnce sync.Once

func VanillaCrimsonNylium() Behavior {
	vanillaCrimsonNyliumOnce.Do(func() {
		vanillaCrimsonNylium = NewNylium(mustVanillaBlockIdentifier(CRIMSON_NYLIUM), "Crimson Nylium", NewBlockTypeInfo(BlockBreakInfoPickaxe(0.4, vanillaToolTierWood, nil), []string{BlockTypeTagsNylium}, nil), nil)
	})
	return vanillaCrimsonNylium.Clone()
}

// VanillaWarpedNylium is a port of VanillaBlocks::WARPED_NYLIUM() - see VanillaBlocksInputs.php's
// registerDelayed("warped_nylium", ...): BreakInfo::pickaxe(0.4, ToolTier::WOOD), [Tags::NYLIUM].
// Vegetation is nil, same reasoning as VanillaCrimsonNylium.
var vanillaWarpedNyliumOnce sync.Once

func VanillaWarpedNylium() Behavior {
	vanillaWarpedNyliumOnce.Do(func() {
		vanillaWarpedNylium = NewNylium(mustVanillaBlockIdentifier(WARPED_NYLIUM), "Warped Nylium", NewBlockTypeInfo(BlockBreakInfoPickaxe(0.4, vanillaToolTierWood, nil), []string{BlockTypeTagsNylium}, nil), nil)
	})
	return vanillaWarpedNylium.Clone()
}

// VanillaAmethystCluster is a port of VanillaBlocks::AMETHYST_CLUSTER() - see
// VanillaBlocksInputs.php's register("amethyst_cluster", ...): BreakInfo::pickaxe(1.5, ToolTier::WOOD).
var vanillaAmethystClusterOnce sync.Once

func VanillaAmethystCluster() Behavior {
	vanillaAmethystClusterOnce.Do(func() {
		vanillaAmethystCluster = NewAmethystCluster(mustVanillaBlockIdentifier(AMETHYST_CLUSTER), "Amethyst Cluster", NewBlockTypeInfo(BlockBreakInfoPickaxe(1.5, vanillaToolTierWood, nil), nil, nil))
	})
	return vanillaAmethystCluster.Clone()
}

// VanillaCobblestone is a port of VanillaBlocks::COBBLESTONE() - see VanillaBlocksInputs.php's
// register("cobblestone", ...): BreakInfo::pickaxe(2.0, ToolTier::WOOD, 30.0).
var vanillaCobblestoneOnce sync.Once

func VanillaCobblestone() Behavior {
	vanillaCobblestoneOnce.Do(func() {
		blastResistance := 30.0
		cobble := &Opaque{Block: NewBlock(mustVanillaBlockIdentifier(COBBLESTONE), "Cobblestone", NewBlockTypeInfo(BlockBreakInfoPickaxe(2.0, vanillaToolTierWood, &blastResistance), nil, nil))}
		cobble.Init(cobble)
		vanillaCobblestone = cobble
	})
	return vanillaCobblestone.Clone()
}

// VanillaBasalt is a port of VanillaBlocks::BASALT() - see VanillaBlocksInputs.php's
// register("basalt", ...): BreakInfo::pickaxe(1.25, ToolTier::WOOD, 21.0).
var vanillaBasaltOnce sync.Once

func VanillaBasalt() Behavior {
	vanillaBasaltOnce.Do(func() {
		blastResistance := 21.0
		vanillaBasalt = NewSimplePillar(mustVanillaBlockIdentifier(BASALT), "Basalt", NewBlockTypeInfo(BlockBreakInfoPickaxe(1.25, vanillaToolTierWood, &blastResistance), nil, nil))
	})
	return vanillaBasalt.Clone()
}

// VanillaGrass is a port of VanillaBlocks::GRASS() - see VanillaBlocksInputs.php's
// register("grass", ...): BreakInfo::shovel(0.6), [Tags::DIRT].
var vanillaGrassOnce sync.Once

func VanillaGrass() Behavior {
	vanillaGrassOnce.Do(func() {
		vanillaGrass = NewGrass(mustVanillaBlockIdentifier(GRASS), "Grass", NewBlockTypeInfo(BlockBreakInfoShovel(0.6, nil, nil), []string{BlockTypeTagsDirt}, nil))
	})
	return vanillaGrass.Clone()
}

// VanillaMycelium is a port of VanillaBlocks::MYCELIUM() - see VanillaBlocksInputs.php's
// register("mycelium", ...): BreakInfo::shovel(0.6), [Tags::DIRT].
var vanillaMyceliumOnce sync.Once

func VanillaMycelium() Behavior {
	vanillaMyceliumOnce.Do(func() {
		vanillaMycelium = NewMycelium(mustVanillaBlockIdentifier(MYCELIUM), "Mycelium", NewBlockTypeInfo(BlockBreakInfoShovel(0.6, nil, nil), []string{BlockTypeTagsDirt}, nil))
	})
	return vanillaMycelium.Clone()
}

// VanillaCactus is a port of VanillaBlocks::CACTUS() - see VanillaBlocksInputs.php's
// register("cactus", ...): new BreakInfo(0.4) (no tool type, no tier), [Tags::POTTABLE_PLANTS].
var vanillaCactusOnce sync.Once

func VanillaCactus() Behavior {
	vanillaCactusOnce.Do(func() {
		vanillaCactus = NewCactus(mustVanillaBlockIdentifier(CACTUS), "Cactus", NewBlockTypeInfo(NewBlockBreakInfo(0.4, ToolTypeNone, 0, nil, nil), []string{BlockTypeTagsPottablePlants}, nil))
	})
	return vanillaCactus.Clone()
}

// VanillaCactusFlower is a port of VanillaBlocks::CACTUS_FLOWER() - see VanillaBlocksInputs.php's
// register("cactus_flower", ...): BreakInfo::instant().
var vanillaCactusFlowerOnce sync.Once

func VanillaCactusFlower() Behavior {
	vanillaCactusFlowerOnce.Do(func() {
		vanillaCactusFlower = NewCactusFlower(mustVanillaBlockIdentifier(CACTUS_FLOWER), "Cactus Flower", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	})
	return vanillaCactusFlower.Clone()
}

// VanillaOminousBanner is a port of VanillaBlocks::OMINOUS_BANNER() - see
// VanillaBlocksInputs.php's register("ominous_banner", ...): BreakInfo::axe(1.0) (no tool tier).
var vanillaOminousBannerOnce sync.Once

func VanillaOminousBanner() Behavior {
	vanillaOminousBannerOnce.Do(func() {
		vanillaOminousBanner = NewOminousFloorBanner(mustVanillaBlockIdentifier(OMINOUS_BANNER), "Ominous Banner", NewBlockTypeInfo(BlockBreakInfoAxe(1.0, nil, nil), nil, nil))
	})
	return vanillaOminousBanner.Clone()
}

// VanillaOminousWallBanner is a port of VanillaBlocks::OMINOUS_WALL_BANNER() - see
// VanillaBlocksInputs.php's register("ominous_wall_banner", ...): BreakInfo::axe(1.0) (no tool
// tier, same $bannerBreakInfo as VanillaOminousBanner).
var vanillaOminousWallBannerOnce sync.Once

func VanillaOminousWallBanner() Behavior {
	vanillaOminousWallBannerOnce.Do(func() {
		vanillaOminousWallBanner = NewOminousWallBanner(mustVanillaBlockIdentifier(OMINOUS_WALL_BANNER), "Ominous Wall Banner", NewBlockTypeInfo(BlockBreakInfoAxe(1.0, nil, nil), nil, nil))
	})
	return vanillaOminousWallBanner.Clone()
}

// VanillaBamboo is a port of VanillaBlocks::BAMBOO() - see VanillaBlocksInputs.php's
// register("bamboo", ...): new BreakInfo(1.0, ToolType::AXE), [Tags::POTTABLE_PLANTS]. The real
// BreakInfo is an anonymous subclass that also makes a sword break bamboo instantly
// (getBreakTime() returns 0.0 for ToolType::SWORD) - BlockBreakInfo isn't self-dispatched in this
// port (no concrete type overrides its methods anywhere else either), so that one behavioral
// nuance is dropped here; every other field is copied exactly.
var vanillaBambooOnce sync.Once

func VanillaBamboo() Behavior {
	vanillaBambooOnce.Do(func() {
		vanillaBamboo = NewBamboo(mustVanillaBlockIdentifier(BAMBOO), "Bamboo", NewBlockTypeInfo(NewBlockBreakInfo(1.0, ToolTypeAxe, 0, nil, nil), []string{BlockTypeTagsPottablePlants}, nil))
	})
	return vanillaBamboo.Clone()
}

// VanillaChorusPlant is a port of VanillaBlocks::CHORUS_PLANT() - see VanillaBlocksInputs.php's
// register("chorus_plant", ...): BreakInfo::axe(0.4).
var vanillaChorusPlantOnce sync.Once

func VanillaChorusPlant() Behavior {
	vanillaChorusPlantOnce.Do(func() {
		vanillaChorusPlant = NewChorusPlant(mustVanillaBlockIdentifier(CHORUS_PLANT), "Chorus Plant", NewBlockTypeInfo(BlockBreakInfoAxe(0.4, nil, nil), nil, nil))
	})
	return vanillaChorusPlant.Clone()
}

// VanillaDoublePitcherCrop is a port of VanillaBlocks::DOUBLE_PITCHER_CROP() - see
// VanillaBlocksInputs.php's register("double_pitcher_crop", ...): BreakInfo::instant().
var vanillaDoublePitcherCropOnce sync.Once

func VanillaDoublePitcherCrop() Behavior {
	vanillaDoublePitcherCropOnce.Do(func() {
		vanillaDoublePitcherCrop = NewDoublePitcherCrop(mustVanillaBlockIdentifier(DOUBLE_PITCHER_CROP), "Double Pitcher Crop", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	})
	return vanillaDoublePitcherCrop.Clone()
}

// VanillaBigDripleafStem is a port of VanillaBlocks::BIG_DRIPLEAF_STEM() - see
// VanillaBlocksInputs.php's register("big_dripleaf_stem", ...): new BreakInfo(0.1) (no tool type,
// no tier).
var vanillaBigDripleafStemOnce sync.Once

func VanillaBigDripleafStem() Behavior {
	vanillaBigDripleafStemOnce.Do(func() {
		vanillaBigDripleafStem = NewBigDripleafStem(mustVanillaBlockIdentifier(BIG_DRIPLEAF_STEM), "Big Dripleaf Stem", NewBlockTypeInfo(NewBlockBreakInfo(0.1, ToolTypeNone, 0, nil, nil), nil, nil))
	})
	return vanillaBigDripleafStem.Clone()
}

// VanillaSmallDripleaf is a port of VanillaBlocks::SMALL_DRIPLEAF() - see
// VanillaBlocksInputs.php's register("small_dripleaf", ...): BreakInfo::instant(ToolType::SHEARS,
// toolHarvestLevel: 1).
var vanillaSmallDripleafOnce sync.Once

func VanillaSmallDripleaf() Behavior {
	vanillaSmallDripleafOnce.Do(func() {
		vanillaSmallDripleaf = NewSmallDripleaf(mustVanillaBlockIdentifier(SMALL_DRIPLEAF), "Small Dripleaf", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeShears, 1), nil, nil))
	})
	return vanillaSmallDripleaf.Clone()
}

// VanillaBigDripleafHead is a port of VanillaBlocks::BIG_DRIPLEAF_HEAD() - see
// VanillaBlocksInputs.php's register("big_dripleaf_head", ...): new BreakInfo(0.1) (no tool type,
// no tier).
var vanillaBigDripleafHeadOnce sync.Once

func VanillaBigDripleafHead() Behavior {
	vanillaBigDripleafHeadOnce.Do(func() {
		vanillaBigDripleafHead = NewBigDripleafHead(mustVanillaBlockIdentifier(BIG_DRIPLEAF_HEAD), "Big Dripleaf", NewBlockTypeInfo(NewBlockBreakInfo(0.1, ToolTypeNone, 0, nil, nil), nil, nil))
	})
	return vanillaBigDripleafHead.Clone()
}

// VanillaStone is a port of VanillaBlocks::STONE() - see VanillaBlocksInputs.php's
// register("stone", ...): BreakInfo::pickaxe(1.5, ToolTier::WOOD, 30.0).
var vanillaStoneOnce sync.Once

func VanillaStone() Behavior {
	vanillaStoneOnce.Do(func() {
		blastResistance := 30.0
		vanillaStone = NewStone(mustVanillaBlockIdentifier(STONE), "Stone", NewBlockTypeInfo(BlockBreakInfoPickaxe(1.5, vanillaToolTierWood, &blastResistance), nil, nil))
	})
	return vanillaStone.Clone()
}

// VanillaBedrock is a port of VanillaBlocks::BEDROCK() - see VanillaBlocksInputs.php's
// register("bedrock", ...): BreakInfo::indestructible(18000000.0).
var vanillaBedrockOnce sync.Once

func VanillaBedrock() Behavior {
	vanillaBedrockOnce.Do(func() {
		vanillaBedrock = NewBedrock(mustVanillaBlockIdentifier(BEDROCK), "Bedrock", NewBlockTypeInfo(BlockBreakInfoIndestructible(18000000.0), nil, nil))
	})
	return vanillaBedrock.Clone()
}

// VanillaTallGrass is a port of VanillaBlocks::TALL_GRASS() - see VanillaBlocksInputs.php's
// register("tall_grass", ...): BreakInfo::instant(ToolType::SHEARS, 1). No DoublePlantVariant
// (nil) since VanillaDoubleTallGrass isn't ported yet.
var vanillaTallGrassOnce sync.Once

func VanillaTallGrass() Behavior {
	vanillaTallGrassOnce.Do(func() {
		vanillaTallGrass = NewTallGrass(mustVanillaBlockIdentifier(TALL_GRASS), "Tall Grass", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeShears, 1), nil, nil), nil)
	})
	return vanillaTallGrass.Clone()
}

// VanillaGravel is a port of VanillaBlocks::GRAVEL() - see VanillaBlocksInputs.php's
// register("gravel", ...): BreakInfo::shovel(0.6).
var vanillaGravelOnce sync.Once

func VanillaGravel() Behavior {
	vanillaGravelOnce.Do(func() {
		vanillaGravel = NewGravel(mustVanillaBlockIdentifier(GRAVEL), "Gravel", NewBlockTypeInfo(BlockBreakInfoShovel(0.6, vanillaToolTierWood, nil), nil, nil))
	})
	return vanillaGravel.Clone()
}

// vanillaStoneOreBreakInfo mirrors VanillaBlocksInputs::registerOres's local
// $stoneOreBreakInfo = fn(ToolTier $toolTier) => new Info(BreakInfo::pickaxe(3.0, $toolTier)).
func vanillaStoneOreBreakInfo(toolTier ToolTier) *BlockBreakInfo {
	return BlockBreakInfoPickaxe(3.0, toolTier, nil)
}

// VanillaCoalOre is a port of VanillaBlocks::COAL_ORE() - see VanillaBlocksInputs.php's
// register("coal_ore", ...): $stoneOreBreakInfo(ToolTier::WOOD).
var vanillaCoalOreOnce sync.Once

func VanillaCoalOre() Behavior {
	vanillaCoalOreOnce.Do(func() {
		vanillaCoalOre = NewCoalOre(mustVanillaBlockIdentifier(COAL_ORE), "Coal Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierWood), nil, nil))
	})
	return vanillaCoalOre.Clone()
}

// VanillaDiamondOre is a port of VanillaBlocks::DIAMOND_ORE() - see VanillaBlocksInputs.php's
// register("diamond_ore", ...): $stoneOreBreakInfo(ToolTier::IRON).
var vanillaDiamondOreOnce sync.Once

func VanillaDiamondOre() Behavior {
	vanillaDiamondOreOnce.Do(func() {
		vanillaDiamondOre = NewDiamondOre(mustVanillaBlockIdentifier(DIAMOND_ORE), "Diamond Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierIron), nil, nil))
	})
	return vanillaDiamondOre.Clone()
}

// VanillaGoldOre is a port of VanillaBlocks::GOLD_ORE() - see VanillaBlocksInputs.php's
// register("gold_ore", ...): $stoneOreBreakInfo(ToolTier::IRON).
var vanillaGoldOreOnce sync.Once

func VanillaGoldOre() Behavior {
	vanillaGoldOreOnce.Do(func() {
		vanillaGoldOre = NewGoldOre(mustVanillaBlockIdentifier(GOLD_ORE), "Gold Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierIron), nil, nil))
	})
	return vanillaGoldOre.Clone()
}

// VanillaIronOre is a port of VanillaBlocks::IRON_ORE() - see VanillaBlocksInputs.php's
// register("iron_ore", ...): $stoneOreBreakInfo(ToolTier::STONE).
var vanillaIronOreOnce sync.Once

func VanillaIronOre() Behavior {
	vanillaIronOreOnce.Do(func() {
		vanillaIronOre = NewIronOre(mustVanillaBlockIdentifier(IRON_ORE), "Iron Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierStone), nil, nil))
	})
	return vanillaIronOre.Clone()
}

// VanillaLapisLazuliOre is a port of VanillaBlocks::LAPIS_LAZULI_ORE() - see
// VanillaBlocksInputs.php's register("lapis_lazuli_ore", ...): $stoneOreBreakInfo(ToolTier::STONE).
var vanillaLapisLazuliOreOnce sync.Once

func VanillaLapisLazuliOre() Behavior {
	vanillaLapisLazuliOreOnce.Do(func() {
		vanillaLapisLazuliOre = NewLapisOre(mustVanillaBlockIdentifier(LAPIS_LAZULI_ORE), "Lapis Lazuli Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierStone), nil, nil))
	})
	return vanillaLapisLazuliOre.Clone()
}

// VanillaRedstoneOre is a port of VanillaBlocks::REDSTONE_ORE() - see VanillaBlocksInputs.php's
// register("redstone_ore", ...): $stoneOreBreakInfo(ToolTier::IRON).
var vanillaRedstoneOreOnce sync.Once

func VanillaRedstoneOre() Behavior {
	vanillaRedstoneOreOnce.Do(func() {
		vanillaRedstoneOre = NewRedstoneOre(mustVanillaBlockIdentifier(REDSTONE_ORE), "Redstone Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierIron), nil, nil))
	})
	return vanillaRedstoneOre.Clone()
}

// VanillaEmeraldOre is a port of VanillaBlocks::EMERALD_ORE() - see VanillaBlocksInputs.php's
// register("emerald_ore", ...): $stoneOreBreakInfo(ToolTier::IRON).
var vanillaEmeraldOreOnce sync.Once

func VanillaEmeraldOre() Behavior {
	vanillaEmeraldOreOnce.Do(func() {
		vanillaEmeraldOre = NewEmeraldOre(mustVanillaBlockIdentifier(EMERALD_ORE), "Emerald Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierIron), nil, nil))
	})
	return vanillaEmeraldOre.Clone()
}

// VanillaSand is a port of VanillaBlocks::SAND() - see VanillaBlocksInputs.php's
// register("sand", ...): $sandTypeInfo = new Info(BreakInfo::shovel(0.5), [Tags::SAND]).
var vanillaSandOnce sync.Once

func VanillaSand() Behavior {
	vanillaSandOnce.Do(func() {
		vanillaSand = NewSand(mustVanillaBlockIdentifier(SAND), "Sand", NewBlockTypeInfo(BlockBreakInfoShovel(0.5, nil, nil), []string{BlockTypeTagsSand}, nil))
	})
	return vanillaSand.Clone()
}

// VanillaSandstone is a port of VanillaBlocks::SANDSTONE() - see VanillaBlocksInputs.php's
// register("sandstone", ...): $sandstoneBreakInfo = new Info(BreakInfo::pickaxe(0.8, ToolTier::WOOD)).
var vanillaSandstoneOnce sync.Once

func VanillaSandstone() Behavior {
	vanillaSandstoneOnce.Do(func() {
		sandstone := &Opaque{Block: NewBlock(mustVanillaBlockIdentifier(SANDSTONE), "Sandstone", NewBlockTypeInfo(BlockBreakInfoPickaxe(0.8, vanillaToolTierWood, nil), nil, nil))}
		sandstone.Init(sandstone)
		vanillaSandstone = sandstone
	})
	return vanillaSandstone.Clone()
}

// VanillaSnowLayer is a port of VanillaBlocks::SNOW_LAYER() - see VanillaBlocksInputs.php's
// register("snow_layer", ...): BreakInfo::shovel(0.1, ToolTier::WOOD).
var vanillaSnowLayerOnce sync.Once

func VanillaSnowLayer() Behavior {
	vanillaSnowLayerOnce.Do(func() {
		vanillaSnowLayer = NewSnowLayer(mustVanillaBlockIdentifier(SNOW_LAYER), "Snow Layer", NewBlockTypeInfo(BlockBreakInfoShovel(0.1, vanillaToolTierWood, nil), nil, nil))
	})
	return vanillaSnowLayer.Clone()
}

// vanillaLogBreakInfo mirrors VanillaBlocksInputs::registerWoodenBlocks's local
// $logBreakInfo = new Info(BreakInfo::axe(2.0)) - no tool tier requirement.
func vanillaLogBreakInfo() *BlockBreakInfo { return BlockBreakInfoAxe(2.0, nil, nil) }

// vanillaLeavesBreakInfo mirrors registerWoodenBlocks's local $leavesBreakInfo (BreakInfo::hoe(0.2)
// hardness with a shears-is-instant override in real PHP's anonymous BreakInfo subclass) - the
// shears-instant special case lives in Leaves' own GetBreakTime-equivalent behaviour, not in this
// registration, so this is just the base hardness/tool type.
func vanillaLeavesBreakInfo() *BlockBreakInfo {
	return NewBlockBreakInfo(0.2, ToolTypeHoe, 0, nil, nil)
}

// VanillaOakLog is a port of VanillaBlocks::OAK_LOG() - see VanillaBlocksInputs.php's
// registerWoodenBlocks loop (WoodType::OAK, standard log suffix).
var vanillaOakLogOnce sync.Once

func VanillaOakLog() Behavior {
	vanillaOakLogOnce.Do(func() {
		vanillaOakLog = NewWood(mustVanillaBlockIdentifier(OAK_LOG), "Oak Log", NewBlockTypeInfo(vanillaLogBreakInfo(), nil, nil), blockutils.WoodTypeOak)
	})
	return vanillaOakLog.Clone()
}

// VanillaOakLeaves is a port of VanillaBlocks::OAK_LEAVES() - see VanillaBlocksInputs.php's
// LeavesType::cases() loop.
var vanillaOakLeavesOnce sync.Once

func VanillaOakLeaves() Behavior {
	vanillaOakLeavesOnce.Do(func() {
		vanillaOakLeaves = NewLeaves(mustVanillaBlockIdentifier(OAK_LEAVES), "Oak Leaves", NewBlockTypeInfo(vanillaLeavesBreakInfo(), nil, nil), blockutils.LeavesTypeOak)
	})
	return vanillaOakLeaves.Clone()
}

// VanillaSpruceLog is a port of VanillaBlocks::SPRUCE_LOG().
var vanillaSpruceLogOnce sync.Once

func VanillaSpruceLog() Behavior {
	vanillaSpruceLogOnce.Do(func() {
		vanillaSpruceLog = NewWood(mustVanillaBlockIdentifier(SPRUCE_LOG), "Spruce Log", NewBlockTypeInfo(vanillaLogBreakInfo(), nil, nil), blockutils.WoodTypeSpruce)
	})
	return vanillaSpruceLog.Clone()
}

// VanillaSpruceLeaves is a port of VanillaBlocks::SPRUCE_LEAVES().
var vanillaSpruceLeavesOnce sync.Once

func VanillaSpruceLeaves() Behavior {
	vanillaSpruceLeavesOnce.Do(func() {
		vanillaSpruceLeaves = NewLeaves(mustVanillaBlockIdentifier(SPRUCE_LEAVES), "Spruce Leaves", NewBlockTypeInfo(vanillaLeavesBreakInfo(), nil, nil), blockutils.LeavesTypeSpruce)
	})
	return vanillaSpruceLeaves.Clone()
}

// VanillaBirchLog is a port of VanillaBlocks::BIRCH_LOG().
var vanillaBirchLogOnce sync.Once

func VanillaBirchLog() Behavior {
	vanillaBirchLogOnce.Do(func() {
		vanillaBirchLog = NewWood(mustVanillaBlockIdentifier(BIRCH_LOG), "Birch Log", NewBlockTypeInfo(vanillaLogBreakInfo(), nil, nil), blockutils.WoodTypeBirch)
	})
	return vanillaBirchLog.Clone()
}

// VanillaBirchLeaves is a port of VanillaBlocks::BIRCH_LEAVES().
var vanillaBirchLeavesOnce sync.Once

func VanillaBirchLeaves() Behavior {
	vanillaBirchLeavesOnce.Do(func() {
		vanillaBirchLeaves = NewLeaves(mustVanillaBlockIdentifier(BIRCH_LEAVES), "Birch Leaves", NewBlockTypeInfo(vanillaLeavesBreakInfo(), nil, nil), blockutils.LeavesTypeBirch)
	})
	return vanillaBirchLeaves.Clone()
}

// VanillaFire is a port of VanillaBlocks::FIRE() - see VanillaBlocksInputs.php's
// register("fire", ...): new Info(BreakInfo::instant(), [Tags::FIRE]).
var vanillaFireOnce sync.Once

func VanillaFire() Behavior {
	vanillaFireOnce.Do(func() {
		vanillaFire = NewFire(mustVanillaBlockIdentifier(FIRE), "Fire", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), []string{BlockTypeTagsFire}, nil))
	})
	return vanillaFire.Clone()
}

// VanillaTNT is a port of VanillaBlocks::TNT() - see VanillaBlocksInputs.php's
// register("tnt", ...): new Info(BreakInfo::instant()).
var vanillaTNTOnce sync.Once

func VanillaTNT() Behavior {
	vanillaTNTOnce.Do(func() {
		vanillaTNT = NewTNT(mustVanillaBlockIdentifier(TNT), "TNT", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	})
	return vanillaTNT.Clone()
}

// VanillaLava is a port of VanillaBlocks::LAVA() - see VanillaBlocksInputs.php's
// register("lava", ...): BreakInfo::indestructible(500.0).
var vanillaLavaOnce sync.Once

func VanillaLava() Behavior {
	vanillaLavaOnce.Do(func() {
		vanillaLava = NewLava(mustVanillaBlockIdentifier(LAVA), "Lava", NewBlockTypeInfo(BlockBreakInfoIndestructible(500.0), nil, nil))
	})
	return vanillaLava.Clone()
}

// VanillaNetherQuartzOre is a port of VanillaBlocks::NETHER_QUARTZ_ORE() - see
// VanillaBlocksInputs.php's registerOres's $netherrackOreBreakInfo = BreakInfo::pickaxe(3.0, ToolTier::WOOD).
var vanillaNetherQuartzOreOnce sync.Once

func VanillaNetherQuartzOre() Behavior {
	vanillaNetherQuartzOreOnce.Do(func() {
		vanillaNetherQuartzOre = NewNetherQuartzOre(mustVanillaBlockIdentifier(NETHER_QUARTZ_ORE), "Nether Quartz Ore", NewBlockTypeInfo(BlockBreakInfoPickaxe(3.0, vanillaToolTierWood, nil), nil, nil))
	})
	return vanillaNetherQuartzOre.Clone()
}

// VanillaIce is a port of VanillaBlocks::ICE() - see VanillaBlocksInputs.php's
// register("ice", ...): new Info(BreakInfo::pickaxe(0.5)).
var vanillaIceOnce sync.Once

func VanillaIce() Behavior {
	vanillaIceOnce.Do(func() {
		vanillaIce = NewIce(mustVanillaBlockIdentifier(ICE), "Ice", NewBlockTypeInfo(BlockBreakInfoPickaxe(0.5, nil, nil), nil, nil))
	})
	return vanillaIce.Clone()
}

// VanillaFrostedIce is a port of VanillaBlocks::FROSTED_ICE() - see VanillaBlocksInputs.php's
// register("frosted_ice", ...): new Info(BreakInfo::pickaxe(0.5)).
var vanillaFrostedIceOnce sync.Once

func VanillaFrostedIce() Behavior {
	vanillaFrostedIceOnce.Do(func() {
		vanillaFrostedIce = NewFrostedIce(mustVanillaBlockIdentifier(FROSTED_ICE), "Frosted Ice", NewBlockTypeInfo(BlockBreakInfoPickaxe(0.5, nil, nil), nil, nil))
	})
	return vanillaFrostedIce.Clone()
}

// VanillaOakPlanks is a port of VanillaBlocks::OAK_PLANKS() - see VanillaBlocksInputs.php's
// wood-type loop: new Planks($id, "Oak Planks", new Info(BreakInfo::axe(2.0, null, 15.0)), WoodType::OAK).
var vanillaOakPlanksOnce sync.Once

func VanillaOakPlanks() Behavior {
	vanillaOakPlanksOnce.Do(func() {
		blastResistance := 15.0
		vanillaOakPlanks = NewPlanks(mustVanillaBlockIdentifier(OAK_PLANKS), "Oak Planks", NewBlockTypeInfo(BlockBreakInfoAxe(2.0, nil, &blastResistance), nil, nil), blockutils.WoodTypeOak)
	})
	return vanillaOakPlanks.Clone()
}
