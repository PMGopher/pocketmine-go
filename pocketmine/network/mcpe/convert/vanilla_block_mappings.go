package convert

import (
	"fmt"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/data/bedrock"
)

// init is a small slice of VanillaBlockMappings::setupBlockStateSerializer - see this package's
// doc comment for why it's only a handful of entries rather than the full ~700.
func init() {
	// mapSimple(Blocks::AIR(), Ids::AIR) - stateless.
	RegisterBlockStateSerializer(block.AIR, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		return bedrock.BlockStateData{Name: "minecraft:air", States: map[string]any{}}, nil
	})

	// mapSimple(Blocks::STONE(), Ids::STONE) - stateless.
	RegisterBlockStateSerializer(block.STONE, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		return bedrock.BlockStateData{Name: "minecraft:stone", States: map[string]any{}}, nil
	})

	// mapSimple(Blocks::GRASS(), Ids::GRASS_BLOCK) - stateless (Bedrock renamed grass ->
	// grass_block; this port keeps the PM-historical "Grass" Go type name regardless, since that's
	// an internal naming choice unrelated to the network property name).
	RegisterBlockStateSerializer(block.GRASS, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		return bedrock.BlockStateData{Name: "minecraft:grass_block", States: map[string]any{}}, nil
	})

	// mapFlattenedId(FlattenedIdModel::create(Blocks::DIRT())->...) - only the NORMAL branch
	// (Ids::DIRT) is ported; COARSE/ROOTED (separate Bedrock block names, coarse_dirt/
	// dirt_with_roots) aren't registered yet.
	RegisterBlockStateSerializer(block.DIRT, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		dirt := blk.(*block.Dirt)
		if dirt.DirtTypeValue != blockutils.DirtTypeNormal {
			return bedrock.BlockStateData{}, fmt.Errorf("convert: unsupported DirtType %v (only DirtTypeNormal is registered so far)", dirt.DirtTypeValue)
		}
		return bedrock.BlockStateData{Name: "minecraft:dirt", States: map[string]any{}}, nil
	})

	// mapModel(Model::create(Blocks::BEDROCK(), Ids::BEDROCK)->properties([infiniburn_bit])).
	RegisterBlockStateSerializer(block.BEDROCK, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		bd := blk.(*block.Bedrock)
		return bedrock.BlockStateData{
			Name:   "minecraft:bedrock",
			States: map[string]any{"infiniburn_bit": boolToByte(bd.BurnsForeverFlag)},
		}, nil
	})

	// mapSimple for every remaining block the world/generator package can place in real terrain
	// (Normal's stone/water fill, GroundCover, Ore decoration, TallGrass) - added together since a
	// screenshot showed water (and everything past the original 5) rendering as the
	// minecraft:info_update fallback "Update!" block, i.e. simply never having been registered here
	// at all.

	// mapSimple(Blocks::GRAVEL(), Ids::GRAVEL) - stateless.
	RegisterBlockStateSerializer(block.GRAVEL, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		return bedrock.BlockStateData{Name: "minecraft:gravel", States: map[string]any{}}, nil
	})

	// mapSimple(Blocks::SAND(), Ids::SAND) - only the default (non-red) variant; Sand has no
	// SandType field in this port yet (see block.Sand's doc comment), matching Bedrock's stateless
	// minecraft:sand in this version (red sand is a separate block name, not ported).
	RegisterBlockStateSerializer(block.SAND, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		return bedrock.BlockStateData{Name: "minecraft:sand", States: map[string]any{}}, nil
	})

	// mapSimple(Blocks::SANDSTONE(), Ids::SANDSTONE) - stateless (chiseled/cut/smooth sandstone are
	// separate block names/types, not ported).
	RegisterBlockStateSerializer(block.SANDSTONE, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		return bedrock.BlockStateData{Name: "minecraft:sandstone", States: map[string]any{}}, nil
	})

	// mapSimple(Blocks::TALL_GRASS(), Ids::SHORT_GRASS) - Bedrock renamed the single-block
	// "tallgrass"/"tall_grass" legacy name to "short_grass" (this port keeps the PM-historical
	// "TallGrass" Go type name regardless - same reasoning as Grass/grass_block above).
	// minecraft:tall_grass is a different, double-tall block (Bedrock's DoublePlant top/bottom
	// half, upper_block_bit) - not this one.
	RegisterBlockStateSerializer(block.TALL_GRASS, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		return bedrock.BlockStateData{Name: "minecraft:short_grass", States: map[string]any{}}, nil
	})

	// mapSimple for every stateless ore (Blocks::*_ORE(), Ids::*_ORE) - deepslate variants aren't
	// ported.
	for typeID, name := range map[int]string{
		block.COAL_ORE:         "minecraft:coal_ore",
		block.IRON_ORE:         "minecraft:iron_ore",
		block.LAPIS_LAZULI_ORE: "minecraft:lapis_ore",
		block.GOLD_ORE:         "minecraft:gold_ore",
		block.DIAMOND_ORE:      "minecraft:diamond_ore",
		block.EMERALD_ORE:      "minecraft:emerald_ore",
	} {
		name := name
		RegisterBlockStateSerializer(typeID, func(blk block.Behavior) (bedrock.BlockStateData, error) {
			return bedrock.BlockStateData{Name: name, States: map[string]any{}}, nil
		})
	}

	// mapModel(Model::create(Blocks::REDSTONE_ORE())) - lit_redstone_ore/redstone_ore are two
	// separate stateless Bedrock block names (matching LightableComponent's Lit field), not one
	// name with a property.
	RegisterBlockStateSerializer(block.REDSTONE_ORE, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		ore := blk.(*block.RedstoneOre)
		name := "minecraft:redstone_ore"
		if ore.Lit {
			name = "minecraft:lit_redstone_ore"
		}
		return bedrock.BlockStateData{Name: name, States: map[string]any{}}, nil
	})

	// mapModel(Model::create(Blocks::SNOW_LAYER())->properties([height])) - Layers is 1-8 (matching
	// the item count semantics SetLayers/GetLayers use), Bedrock's height property is 0-7. covered_bit
	// (whether the block below is also snow, purely cosmetic texture blending) isn't tracked by this
	// port's SnowLayer, so it's always written false.
	RegisterBlockStateSerializer(block.SNOW_LAYER, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		snow := blk.(*block.SnowLayer)
		return bedrock.BlockStateData{
			Name:   "minecraft:snow_layer",
			States: map[string]any{"height": int32(snow.Layers - 1), "covered_bit": uint8(0)},
		}, nil
	})

	// mapModel(Model::create(Blocks::WATER())->properties([liquid_depth])) - water/flowing_water
	// are two separate Bedrock block names (matching Liquid's Still field, same shape as
	// RedstoneOre's Lit above); liquid_depth packs Decay (0-7) and Falling into a single 0-15 value,
	// the well-known standard Bedrock liquid-level encoding (falling adds 8).
	RegisterBlockStateSerializer(block.WATER, func(blk block.Behavior) (bedrock.BlockStateData, error) {
		water := blk.(*block.Water)
		name := "minecraft:flowing_water"
		if water.Still {
			name = "minecraft:water"
		}
		depth := water.Decay
		if water.Falling {
			depth += 8
		}
		return bedrock.BlockStateData{Name: name, States: map[string]any{"liquid_depth": int32(depth)}}, nil
	})
}

// boolToByte matches BoolProperty's real Bedrock NBT encoding (a TAG_Byte, 0 or 1) - see
// BlockStateWriter::writeBool in the PHP original.
func boolToByte(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}
