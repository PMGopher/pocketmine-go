package block

import "fmt"

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
func VanillaAir() Behavior {
	if vanillaAir == nil {
		vanillaAir = NewAir(mustVanillaBlockIdentifier(AIR), "Air", NewBlockTypeInfo(BlockBreakInfoIndestructible(-1.0), nil, nil))
	}
	return vanillaAir.Clone()
}

// VanillaDirt is a port of VanillaBlocks::DIRT() - see VanillaBlocksInputs.php's
// register("dirt", ...): BreakInfo::shovel(0.5), [Tags::DIRT].
func VanillaDirt() Behavior {
	if vanillaDirt == nil {
		vanillaDirt = NewDirt(mustVanillaBlockIdentifier(DIRT), "Dirt", NewBlockTypeInfo(BlockBreakInfoShovel(0.5, nil, nil), []string{BlockTypeTagsDirt}, nil))
	}
	return vanillaDirt.Clone()
}

// VanillaWater is a port of VanillaBlocks::WATER() - see VanillaBlocksInputs.php's
// register("water", ...): BreakInfo::indestructible(500.0).
func VanillaWater() Behavior {
	if vanillaWater == nil {
		vanillaWater = NewWater(mustVanillaBlockIdentifier(WATER), "Water", NewBlockTypeInfo(BlockBreakInfoIndestructible(500.0), nil, nil))
	}
	return vanillaWater.Clone()
}

// VanillaNetherrack is a port of VanillaBlocks::NETHERRACK() - see VanillaBlocksInputs.php's
// register("netherrack", ...): BreakInfo::pickaxe(0.4, ToolTier::WOOD).
func VanillaNetherrack() Behavior {
	if vanillaNetherrack == nil {
		vanillaNetherrack = NewNetherrack(mustVanillaBlockIdentifier(NETHERRACK), "Netherrack", NewBlockTypeInfo(BlockBreakInfoPickaxe(0.4, vanillaToolTierWood, nil), nil, nil))
	}
	return vanillaNetherrack.Clone()
}

// VanillaObsidian is a port of VanillaBlocks::OBSIDIAN() - see VanillaBlocksInputs.php's
// register("obsidian", ...): BreakInfo::pickaxe(35.0, ToolTier::DIAMOND, 6000.0).
func VanillaObsidian() Behavior {
	if vanillaObsidian == nil {
		blastResistance := 6000.0
		obsidian := &Opaque{Block: NewBlock(mustVanillaBlockIdentifier(OBSIDIAN), "Obsidian", NewBlockTypeInfo(BlockBreakInfoPickaxe(35.0, vanillaToolTierDiamond, &blastResistance), nil, nil))}
		obsidian.Init(obsidian)
		vanillaObsidian = obsidian
	}
	return vanillaObsidian.Clone()
}

// VanillaSoulSoil is a port of VanillaBlocks::SOUL_SOIL() - see VanillaBlocksInputs.php's
// register("soul_soil", ...): BreakInfo::shovel(0.5).
func VanillaSoulSoil() Behavior {
	if vanillaSoulSoil == nil {
		soil := &Opaque{Block: NewBlock(mustVanillaBlockIdentifier(SOUL_SOIL), "Soul Soil", NewBlockTypeInfo(BlockBreakInfoShovel(0.5, nil, nil), nil, nil))}
		soil.Init(soil)
		vanillaSoulSoil = soil
	}
	return vanillaSoulSoil.Clone()
}

// VanillaCake is a port of VanillaBlocks::CAKE() - see VanillaBlocksInputs.php's
// register("cake", ...): new BreakInfo(0.5) (no tool type, no tier).
func VanillaCake() Behavior {
	if vanillaCake == nil {
		vanillaCake = NewCake(mustVanillaBlockIdentifier(CAKE), "Cake", NewBlockTypeInfo(NewBlockBreakInfo(0.5, ToolTypeNone, 0, nil, nil), nil, nil))
	}
	return vanillaCake.Clone()
}

// VanillaConcrete is a port of VanillaBlocks::CONCRETE() - see VanillaBlocksInputs.php's
// register("concrete", ...): BreakInfo::pickaxe(1.8, ToolTier::WOOD). Defaults to White, same as
// the real singleton - callers needing a specific color call SetColor on the result, same pattern
// as VanillaCake()/CakeWithCandle.GetResidue.
func VanillaConcrete() Behavior {
	if vanillaConcrete == nil {
		vanillaConcrete = NewConcrete(mustVanillaBlockIdentifier(CONCRETE), "Concrete", NewBlockTypeInfo(BlockBreakInfoPickaxe(1.8, vanillaToolTierWood, nil), nil, nil))
	}
	return vanillaConcrete.Clone()
}

// VanillaMelon is a port of VanillaBlocks::MELON() - see VanillaBlocksInputs.php's
// register("melon", ...): BreakInfo::axe(1.0).
func VanillaMelon() Behavior {
	if vanillaMelon == nil {
		vanillaMelon = NewMelon(mustVanillaBlockIdentifier(MELON), "Melon Block", NewBlockTypeInfo(BlockBreakInfoAxe(1.0, nil, nil), nil, nil))
	}
	return vanillaMelon.Clone()
}

// VanillaPumpkin is a port of VanillaBlocks::PUMPKIN() - see VanillaBlocksInputs.php's
// register("pumpkin", ...): BreakInfo::axe(1.0).
func VanillaPumpkin() Behavior {
	if vanillaPumpkin == nil {
		vanillaPumpkin = NewPumpkin(mustVanillaBlockIdentifier(PUMPKIN), "Pumpkin", NewBlockTypeInfo(BlockBreakInfoAxe(1.0, nil, nil), nil, nil))
	}
	return vanillaPumpkin.Clone()
}

// VanillaSugarcane is a port of VanillaBlocks::SUGARCANE() - see VanillaBlocksInputs.php's
// register("sugarcane", ...): BreakInfo::instant().
func VanillaSugarcane() Behavior {
	if vanillaSugarcane == nil {
		vanillaSugarcane = NewSugarcane(mustVanillaBlockIdentifier(SUGARCANE), "Sugarcane", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	}
	return vanillaSugarcane.Clone()
}

// VanillaCaveVines is a port of VanillaBlocks::CAVE_VINES() - see VanillaBlocksInputs.php's
// register("cave_vines", ...): BreakInfo::instant().
func VanillaCaveVines() Behavior {
	if vanillaCaveVines == nil {
		vanillaCaveVines = NewCaveVines(mustVanillaBlockIdentifier(CAVE_VINES), "Cave Vines", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	}
	return vanillaCaveVines.Clone()
}

// VanillaTorchflower is a port of VanillaBlocks::TORCHFLOWER() - see VanillaBlocksInputs.php's
// register("torchflower", ...): BreakInfo::instant(), [Tags::POTTABLE_PLANTS] ($flowerTypeInfo).
func VanillaTorchflower() Behavior {
	if vanillaTorchflower == nil {
		vanillaTorchflower = NewFlower(mustVanillaBlockIdentifier(TORCHFLOWER), "Torchflower", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), []string{BlockTypeTagsPottablePlants}, nil))
	}
	return vanillaTorchflower.Clone()
}

// VanillaTorchflowerCrop is a port of VanillaBlocks::TORCHFLOWER_CROP() - see
// VanillaBlocksInputs.php's register("torchflower_crop", ...): BreakInfo::instant().
func VanillaTorchflowerCrop() Behavior {
	if vanillaTorchflowerCrop == nil {
		vanillaTorchflowerCrop = NewTorchflowerCrop(mustVanillaBlockIdentifier(TORCHFLOWER_CROP), "Torchflower Crop", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	}
	return vanillaTorchflowerCrop.Clone()
}

// VanillaCrimsonNylium is a port of VanillaBlocks::CRIMSON_NYLIUM() - see VanillaBlocksInputs.php's
// registerDelayed("crimson_nylium", ...): BreakInfo::pickaxe(0.4, ToolTier::WOOD), [Tags::NYLIUM].
// Vegetation is nil (the real singleton lists CRIMSON_FUNGUS/CRIMSON_ROOTS, neither ported yet) -
// harmless for Netherrack.tryTransform, the only current caller, which never reads it.
func VanillaCrimsonNylium() Behavior {
	if vanillaCrimsonNylium == nil {
		vanillaCrimsonNylium = NewNylium(mustVanillaBlockIdentifier(CRIMSON_NYLIUM), "Crimson Nylium", NewBlockTypeInfo(BlockBreakInfoPickaxe(0.4, vanillaToolTierWood, nil), []string{BlockTypeTagsNylium}, nil), nil)
	}
	return vanillaCrimsonNylium.Clone()
}

// VanillaWarpedNylium is a port of VanillaBlocks::WARPED_NYLIUM() - see VanillaBlocksInputs.php's
// registerDelayed("warped_nylium", ...): BreakInfo::pickaxe(0.4, ToolTier::WOOD), [Tags::NYLIUM].
// Vegetation is nil, same reasoning as VanillaCrimsonNylium.
func VanillaWarpedNylium() Behavior {
	if vanillaWarpedNylium == nil {
		vanillaWarpedNylium = NewNylium(mustVanillaBlockIdentifier(WARPED_NYLIUM), "Warped Nylium", NewBlockTypeInfo(BlockBreakInfoPickaxe(0.4, vanillaToolTierWood, nil), []string{BlockTypeTagsNylium}, nil), nil)
	}
	return vanillaWarpedNylium.Clone()
}

// VanillaAmethystCluster is a port of VanillaBlocks::AMETHYST_CLUSTER() - see
// VanillaBlocksInputs.php's register("amethyst_cluster", ...): BreakInfo::pickaxe(1.5, ToolTier::WOOD).
func VanillaAmethystCluster() Behavior {
	if vanillaAmethystCluster == nil {
		vanillaAmethystCluster = NewAmethystCluster(mustVanillaBlockIdentifier(AMETHYST_CLUSTER), "Amethyst Cluster", NewBlockTypeInfo(BlockBreakInfoPickaxe(1.5, vanillaToolTierWood, nil), nil, nil))
	}
	return vanillaAmethystCluster.Clone()
}

// VanillaCobblestone is a port of VanillaBlocks::COBBLESTONE() - see VanillaBlocksInputs.php's
// register("cobblestone", ...): BreakInfo::pickaxe(2.0, ToolTier::WOOD, 30.0).
func VanillaCobblestone() Behavior {
	if vanillaCobblestone == nil {
		blastResistance := 30.0
		cobble := &Opaque{Block: NewBlock(mustVanillaBlockIdentifier(COBBLESTONE), "Cobblestone", NewBlockTypeInfo(BlockBreakInfoPickaxe(2.0, vanillaToolTierWood, &blastResistance), nil, nil))}
		cobble.Init(cobble)
		vanillaCobblestone = cobble
	}
	return vanillaCobblestone.Clone()
}

// VanillaBasalt is a port of VanillaBlocks::BASALT() - see VanillaBlocksInputs.php's
// register("basalt", ...): BreakInfo::pickaxe(1.25, ToolTier::WOOD, 21.0).
func VanillaBasalt() Behavior {
	if vanillaBasalt == nil {
		blastResistance := 21.0
		vanillaBasalt = NewSimplePillar(mustVanillaBlockIdentifier(BASALT), "Basalt", NewBlockTypeInfo(BlockBreakInfoPickaxe(1.25, vanillaToolTierWood, &blastResistance), nil, nil))
	}
	return vanillaBasalt.Clone()
}

// VanillaGrass is a port of VanillaBlocks::GRASS() - see VanillaBlocksInputs.php's
// register("grass", ...): BreakInfo::shovel(0.6), [Tags::DIRT].
func VanillaGrass() Behavior {
	if vanillaGrass == nil {
		vanillaGrass = NewGrass(mustVanillaBlockIdentifier(GRASS), "Grass", NewBlockTypeInfo(BlockBreakInfoShovel(0.6, nil, nil), []string{BlockTypeTagsDirt}, nil))
	}
	return vanillaGrass.Clone()
}

// VanillaMycelium is a port of VanillaBlocks::MYCELIUM() - see VanillaBlocksInputs.php's
// register("mycelium", ...): BreakInfo::shovel(0.6), [Tags::DIRT].
func VanillaMycelium() Behavior {
	if vanillaMycelium == nil {
		vanillaMycelium = NewMycelium(mustVanillaBlockIdentifier(MYCELIUM), "Mycelium", NewBlockTypeInfo(BlockBreakInfoShovel(0.6, nil, nil), []string{BlockTypeTagsDirt}, nil))
	}
	return vanillaMycelium.Clone()
}

// VanillaCactus is a port of VanillaBlocks::CACTUS() - see VanillaBlocksInputs.php's
// register("cactus", ...): new BreakInfo(0.4) (no tool type, no tier), [Tags::POTTABLE_PLANTS].
func VanillaCactus() Behavior {
	if vanillaCactus == nil {
		vanillaCactus = NewCactus(mustVanillaBlockIdentifier(CACTUS), "Cactus", NewBlockTypeInfo(NewBlockBreakInfo(0.4, ToolTypeNone, 0, nil, nil), []string{BlockTypeTagsPottablePlants}, nil))
	}
	return vanillaCactus.Clone()
}

// VanillaCactusFlower is a port of VanillaBlocks::CACTUS_FLOWER() - see VanillaBlocksInputs.php's
// register("cactus_flower", ...): BreakInfo::instant().
func VanillaCactusFlower() Behavior {
	if vanillaCactusFlower == nil {
		vanillaCactusFlower = NewCactusFlower(mustVanillaBlockIdentifier(CACTUS_FLOWER), "Cactus Flower", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	}
	return vanillaCactusFlower.Clone()
}

// VanillaOminousBanner is a port of VanillaBlocks::OMINOUS_BANNER() - see
// VanillaBlocksInputs.php's register("ominous_banner", ...): BreakInfo::axe(1.0) (no tool tier).
func VanillaOminousBanner() Behavior {
	if vanillaOminousBanner == nil {
		vanillaOminousBanner = NewOminousFloorBanner(mustVanillaBlockIdentifier(OMINOUS_BANNER), "Ominous Banner", NewBlockTypeInfo(BlockBreakInfoAxe(1.0, nil, nil), nil, nil))
	}
	return vanillaOminousBanner.Clone()
}

// VanillaOminousWallBanner is a port of VanillaBlocks::OMINOUS_WALL_BANNER() - see
// VanillaBlocksInputs.php's register("ominous_wall_banner", ...): BreakInfo::axe(1.0) (no tool
// tier, same $bannerBreakInfo as VanillaOminousBanner).
func VanillaOminousWallBanner() Behavior {
	if vanillaOminousWallBanner == nil {
		vanillaOminousWallBanner = NewOminousWallBanner(mustVanillaBlockIdentifier(OMINOUS_WALL_BANNER), "Ominous Wall Banner", NewBlockTypeInfo(BlockBreakInfoAxe(1.0, nil, nil), nil, nil))
	}
	return vanillaOminousWallBanner.Clone()
}

// VanillaBamboo is a port of VanillaBlocks::BAMBOO() - see VanillaBlocksInputs.php's
// register("bamboo", ...): new BreakInfo(1.0, ToolType::AXE), [Tags::POTTABLE_PLANTS]. The real
// BreakInfo is an anonymous subclass that also makes a sword break bamboo instantly
// (getBreakTime() returns 0.0 for ToolType::SWORD) - BlockBreakInfo isn't self-dispatched in this
// port (no concrete type overrides its methods anywhere else either), so that one behavioral
// nuance is dropped here; every other field is copied exactly.
func VanillaBamboo() Behavior {
	if vanillaBamboo == nil {
		vanillaBamboo = NewBamboo(mustVanillaBlockIdentifier(BAMBOO), "Bamboo", NewBlockTypeInfo(NewBlockBreakInfo(1.0, ToolTypeAxe, 0, nil, nil), []string{BlockTypeTagsPottablePlants}, nil))
	}
	return vanillaBamboo.Clone()
}

// VanillaChorusPlant is a port of VanillaBlocks::CHORUS_PLANT() - see VanillaBlocksInputs.php's
// register("chorus_plant", ...): BreakInfo::axe(0.4).
func VanillaChorusPlant() Behavior {
	if vanillaChorusPlant == nil {
		vanillaChorusPlant = NewChorusPlant(mustVanillaBlockIdentifier(CHORUS_PLANT), "Chorus Plant", NewBlockTypeInfo(BlockBreakInfoAxe(0.4, nil, nil), nil, nil))
	}
	return vanillaChorusPlant.Clone()
}

// VanillaDoublePitcherCrop is a port of VanillaBlocks::DOUBLE_PITCHER_CROP() - see
// VanillaBlocksInputs.php's register("double_pitcher_crop", ...): BreakInfo::instant().
func VanillaDoublePitcherCrop() Behavior {
	if vanillaDoublePitcherCrop == nil {
		vanillaDoublePitcherCrop = NewDoublePitcherCrop(mustVanillaBlockIdentifier(DOUBLE_PITCHER_CROP), "Double Pitcher Crop", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	}
	return vanillaDoublePitcherCrop.Clone()
}

// VanillaBigDripleafStem is a port of VanillaBlocks::BIG_DRIPLEAF_STEM() - see
// VanillaBlocksInputs.php's register("big_dripleaf_stem", ...): new BreakInfo(0.1) (no tool type,
// no tier).
func VanillaBigDripleafStem() Behavior {
	if vanillaBigDripleafStem == nil {
		vanillaBigDripleafStem = NewBigDripleafStem(mustVanillaBlockIdentifier(BIG_DRIPLEAF_STEM), "Big Dripleaf Stem", NewBlockTypeInfo(NewBlockBreakInfo(0.1, ToolTypeNone, 0, nil, nil), nil, nil))
	}
	return vanillaBigDripleafStem.Clone()
}

// VanillaSmallDripleaf is a port of VanillaBlocks::SMALL_DRIPLEAF() - see
// VanillaBlocksInputs.php's register("small_dripleaf", ...): BreakInfo::instant(ToolType::SHEARS,
// toolHarvestLevel: 1).
func VanillaSmallDripleaf() Behavior {
	if vanillaSmallDripleaf == nil {
		vanillaSmallDripleaf = NewSmallDripleaf(mustVanillaBlockIdentifier(SMALL_DRIPLEAF), "Small Dripleaf", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeShears, 1), nil, nil))
	}
	return vanillaSmallDripleaf.Clone()
}

// VanillaBigDripleafHead is a port of VanillaBlocks::BIG_DRIPLEAF_HEAD() - see
// VanillaBlocksInputs.php's register("big_dripleaf_head", ...): new BreakInfo(0.1) (no tool type,
// no tier).
func VanillaBigDripleafHead() Behavior {
	if vanillaBigDripleafHead == nil {
		vanillaBigDripleafHead = NewBigDripleafHead(mustVanillaBlockIdentifier(BIG_DRIPLEAF_HEAD), "Big Dripleaf", NewBlockTypeInfo(NewBlockBreakInfo(0.1, ToolTypeNone, 0, nil, nil), nil, nil))
	}
	return vanillaBigDripleafHead.Clone()
}

// VanillaStone is a port of VanillaBlocks::STONE() - see VanillaBlocksInputs.php's
// register("stone", ...): BreakInfo::pickaxe(1.5, ToolTier::WOOD, 30.0).
func VanillaStone() Behavior {
	if vanillaStone == nil {
		blastResistance := 30.0
		vanillaStone = NewStone(mustVanillaBlockIdentifier(STONE), "Stone", NewBlockTypeInfo(BlockBreakInfoPickaxe(1.5, vanillaToolTierWood, &blastResistance), nil, nil))
	}
	return vanillaStone.Clone()
}

// VanillaBedrock is a port of VanillaBlocks::BEDROCK() - see VanillaBlocksInputs.php's
// register("bedrock", ...): BreakInfo::indestructible(18000000.0).
func VanillaBedrock() Behavior {
	if vanillaBedrock == nil {
		vanillaBedrock = NewBedrock(mustVanillaBlockIdentifier(BEDROCK), "Bedrock", NewBlockTypeInfo(BlockBreakInfoIndestructible(18000000.0), nil, nil))
	}
	return vanillaBedrock.Clone()
}

// VanillaTallGrass is a port of VanillaBlocks::TALL_GRASS() - see VanillaBlocksInputs.php's
// register("tall_grass", ...): BreakInfo::instant(ToolType::SHEARS, 1). No DoublePlantVariant
// (nil) since VanillaDoubleTallGrass isn't ported yet.
func VanillaTallGrass() Behavior {
	if vanillaTallGrass == nil {
		vanillaTallGrass = NewTallGrass(mustVanillaBlockIdentifier(TALL_GRASS), "Tall Grass", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeShears, 1), nil, nil), nil)
	}
	return vanillaTallGrass.Clone()
}

// VanillaGravel is a port of VanillaBlocks::GRAVEL() - see VanillaBlocksInputs.php's
// register("gravel", ...): BreakInfo::shovel(0.6).
func VanillaGravel() Behavior {
	if vanillaGravel == nil {
		vanillaGravel = NewGravel(mustVanillaBlockIdentifier(GRAVEL), "Gravel", NewBlockTypeInfo(BlockBreakInfoShovel(0.6, vanillaToolTierWood, nil), nil, nil))
	}
	return vanillaGravel.Clone()
}

// vanillaStoneOreBreakInfo mirrors VanillaBlocksInputs::registerOres's local
// $stoneOreBreakInfo = fn(ToolTier $toolTier) => new Info(BreakInfo::pickaxe(3.0, $toolTier)).
func vanillaStoneOreBreakInfo(toolTier ToolTier) *BlockBreakInfo {
	return BlockBreakInfoPickaxe(3.0, toolTier, nil)
}

// VanillaCoalOre is a port of VanillaBlocks::COAL_ORE() - see VanillaBlocksInputs.php's
// register("coal_ore", ...): $stoneOreBreakInfo(ToolTier::WOOD).
func VanillaCoalOre() Behavior {
	if vanillaCoalOre == nil {
		vanillaCoalOre = NewCoalOre(mustVanillaBlockIdentifier(COAL_ORE), "Coal Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierWood), nil, nil))
	}
	return vanillaCoalOre.Clone()
}

// VanillaDiamondOre is a port of VanillaBlocks::DIAMOND_ORE() - see VanillaBlocksInputs.php's
// register("diamond_ore", ...): $stoneOreBreakInfo(ToolTier::IRON).
func VanillaDiamondOre() Behavior {
	if vanillaDiamondOre == nil {
		vanillaDiamondOre = NewDiamondOre(mustVanillaBlockIdentifier(DIAMOND_ORE), "Diamond Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierIron), nil, nil))
	}
	return vanillaDiamondOre.Clone()
}

// VanillaGoldOre is a port of VanillaBlocks::GOLD_ORE() - see VanillaBlocksInputs.php's
// register("gold_ore", ...): $stoneOreBreakInfo(ToolTier::IRON).
func VanillaGoldOre() Behavior {
	if vanillaGoldOre == nil {
		vanillaGoldOre = NewGoldOre(mustVanillaBlockIdentifier(GOLD_ORE), "Gold Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierIron), nil, nil))
	}
	return vanillaGoldOre.Clone()
}

// VanillaIronOre is a port of VanillaBlocks::IRON_ORE() - see VanillaBlocksInputs.php's
// register("iron_ore", ...): $stoneOreBreakInfo(ToolTier::STONE).
func VanillaIronOre() Behavior {
	if vanillaIronOre == nil {
		vanillaIronOre = NewIronOre(mustVanillaBlockIdentifier(IRON_ORE), "Iron Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierStone), nil, nil))
	}
	return vanillaIronOre.Clone()
}

// VanillaLapisLazuliOre is a port of VanillaBlocks::LAPIS_LAZULI_ORE() - see
// VanillaBlocksInputs.php's register("lapis_lazuli_ore", ...): $stoneOreBreakInfo(ToolTier::STONE).
func VanillaLapisLazuliOre() Behavior {
	if vanillaLapisLazuliOre == nil {
		vanillaLapisLazuliOre = NewLapisOre(mustVanillaBlockIdentifier(LAPIS_LAZULI_ORE), "Lapis Lazuli Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierStone), nil, nil))
	}
	return vanillaLapisLazuliOre.Clone()
}

// VanillaRedstoneOre is a port of VanillaBlocks::REDSTONE_ORE() - see VanillaBlocksInputs.php's
// register("redstone_ore", ...): $stoneOreBreakInfo(ToolTier::IRON).
func VanillaRedstoneOre() Behavior {
	if vanillaRedstoneOre == nil {
		vanillaRedstoneOre = NewRedstoneOre(mustVanillaBlockIdentifier(REDSTONE_ORE), "Redstone Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierIron), nil, nil))
	}
	return vanillaRedstoneOre.Clone()
}

// VanillaEmeraldOre is a port of VanillaBlocks::EMERALD_ORE() - see VanillaBlocksInputs.php's
// register("emerald_ore", ...): $stoneOreBreakInfo(ToolTier::IRON).
func VanillaEmeraldOre() Behavior {
	if vanillaEmeraldOre == nil {
		vanillaEmeraldOre = NewEmeraldOre(mustVanillaBlockIdentifier(EMERALD_ORE), "Emerald Ore", NewBlockTypeInfo(vanillaStoneOreBreakInfo(vanillaToolTierIron), nil, nil))
	}
	return vanillaEmeraldOre.Clone()
}

// VanillaSand is a port of VanillaBlocks::SAND() - see VanillaBlocksInputs.php's
// register("sand", ...): $sandTypeInfo = new Info(BreakInfo::shovel(0.5), [Tags::SAND]).
func VanillaSand() Behavior {
	if vanillaSand == nil {
		vanillaSand = NewSand(mustVanillaBlockIdentifier(SAND), "Sand", NewBlockTypeInfo(BlockBreakInfoShovel(0.5, nil, nil), []string{BlockTypeTagsSand}, nil))
	}
	return vanillaSand.Clone()
}

// VanillaSandstone is a port of VanillaBlocks::SANDSTONE() - see VanillaBlocksInputs.php's
// register("sandstone", ...): $sandstoneBreakInfo = new Info(BreakInfo::pickaxe(0.8, ToolTier::WOOD)).
func VanillaSandstone() Behavior {
	if vanillaSandstone == nil {
		sandstone := &Opaque{Block: NewBlock(mustVanillaBlockIdentifier(SANDSTONE), "Sandstone", NewBlockTypeInfo(BlockBreakInfoPickaxe(0.8, vanillaToolTierWood, nil), nil, nil))}
		sandstone.Init(sandstone)
		vanillaSandstone = sandstone
	}
	return vanillaSandstone.Clone()
}

// VanillaSnowLayer is a port of VanillaBlocks::SNOW_LAYER() - see VanillaBlocksInputs.php's
// register("snow_layer", ...): BreakInfo::shovel(0.1, ToolTier::WOOD).
func VanillaSnowLayer() Behavior {
	if vanillaSnowLayer == nil {
		vanillaSnowLayer = NewSnowLayer(mustVanillaBlockIdentifier(SNOW_LAYER), "Snow Layer", NewBlockTypeInfo(BlockBreakInfoShovel(0.1, vanillaToolTierWood, nil), nil, nil))
	}
	return vanillaSnowLayer.Clone()
}
