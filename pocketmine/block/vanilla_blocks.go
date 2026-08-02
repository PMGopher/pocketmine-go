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
	vanillaAir        Behavior
	vanillaDirt       Behavior
	vanillaWater      Behavior
	vanillaNetherrack Behavior
	vanillaObsidian   Behavior
	vanillaSoulSoil   Behavior
	vanillaCake       Behavior
	vanillaConcrete   Behavior
	vanillaMelon      Behavior
	vanillaPumpkin    Behavior
	vanillaSugarcane  Behavior
)

// vanillaHarvestLevel is a minimal block.ToolTier implementation (GetHarvestLevel only) for use
// within this file - the real per-tier metadata (item.ToolTier) lives in the item package, which
// this package can't import (item already imports block). The harvest-level numbers themselves
// are copied from item.ToolTier's real table (WOOD=1, DIAMOND=5, etc.), not guessed.
type vanillaHarvestLevel int

func (h vanillaHarvestLevel) GetHarvestLevel() int { return int(h) }

const (
	vanillaToolTierWood    vanillaHarvestLevel = 1
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
