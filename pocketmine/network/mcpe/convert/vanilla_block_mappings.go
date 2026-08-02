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
}

// boolToByte matches BoolProperty's real Bedrock NBT encoding (a TAG_Byte, 0 or 1) - see
// BlockStateWriter::writeBool in the PHP original.
func boolToByte(v bool) uint8 {
	if v {
		return 1
	}
	return 0
}
