package blockconvert

import (
	"strconv"
	"sync"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
	"pocketmine-go/pocketmine/math"
)

// ValueMappings is a port of pocketmine\data\bedrock\block\convert\property\ValueMappings: the
// value <-> raw state value maps shared by many blocks.
type ValueMappings struct {
	DyeColor           *ValueMap[blockutils.DyeColor, string]
	DyeColorWithSilver *ValueMap[blockutils.DyeColor, string]
	MobHeadType        *ValueMap[blockutils.MobHeadType, string]
	FroglightType      *ValueMap[blockutils.FroglightType, string]
	DirtType           *ValueMap[blockutils.DirtType, string]
	DripleafState      *ValueMap[blockutils.DripleafState, string]
	BellAttachmentType *ValueMap[blockutils.BellAttachmentType, string]
	LeverFacing        *ValueMap[blockutils.LeverFacing, string]
	MushroomBlockType  *ValueMap[blockutils.MushroomBlockType, int]

	CardinalDirection            *ValueMap[math.Facing, string]
	BlockFace                    *ValueMap[math.Facing, string]
	PillarAxis                   *ValueMap[math.Axis, string]
	TorchFacing                  *ValueMap[math.Facing, string]
	PortalAxis                   *ValueMap[math.Axis, string]
	BambooLeafSize               *ValueMap[int, string]
	HorizontalFacing5Minus       *ValueMap[math.Facing, int]
	HorizontalFacingSWNE         *ValueMap[math.Facing, int]
	HorizontalFacingSWNEInverted *ValueMap[math.Facing, int]
	HorizontalFacingCoral        *ValueMap[math.Facing, int]
	HorizontalFacingClassic      *ValueMap[math.Facing, int]
	Facing                       *ValueMap[math.Facing, int]
	FacingEndRod                 *ValueMap[math.Facing, int]
	CoralAxis                    *ValueMap[math.Axis, int]
	FacingExceptDown             *ValueMap[math.Facing, int]
	FacingExceptUp               *ValueMap[math.Facing, int]
	FacingStem                   *ValueMap[math.Facing, int]
}

var (
	valueMappings     *ValueMappings
	valueMappingsOnce sync.Once
)

// GetValueMappings is ValueMappings::getInstance().
func GetValueMappings() *ValueMappings {
	valueMappingsOnce.Do(func() { valueMappings = newValueMappings() })
	return valueMappings
}

// dyeColorNames is the "flattened ID component" name of each dye colour, in DyeColor::cases() order.
var dyeColorNames = []pair[blockutils.DyeColor, string]{
	p(blockutils.DyeColorWhite, "white"),
	p(blockutils.DyeColorOrange, "orange"),
	p(blockutils.DyeColorMagenta, "magenta"),
	p(blockutils.DyeColorLightBlue, "light_blue"),
	p(blockutils.DyeColorYellow, "yellow"),
	p(blockutils.DyeColorLime, "lime"),
	p(blockutils.DyeColorPink, "pink"),
	p(blockutils.DyeColorGray, "gray"),
	p(blockutils.DyeColorLightGray, "light_gray"),
	p(blockutils.DyeColorCyan, "cyan"),
	p(blockutils.DyeColorPurple, "purple"),
	p(blockutils.DyeColorBlue, "blue"),
	p(blockutils.DyeColorBrown, "brown"),
	p(blockutils.DyeColorGreen, "green"),
	p(blockutils.DyeColorRed, "red"),
	p(blockutils.DyeColorBlack, "black"),
}

func newValueMappings() *ValueMappings {
	m := &ValueMappings{}
	//flattened ID components - we can't generate constants for these
	m.DyeColor = newValueMap(dyeColorNames...)
	withSilver := make([]pair[blockutils.DyeColor, string], len(dyeColorNames))
	for i, e := range dyeColorNames {
		if e.value == blockutils.DyeColorLightGray {
			e.raw = "silver"
		}
		withSilver[i] = e
	}
	m.DyeColorWithSilver = newValueMap(withSilver...)
	m.MobHeadType = newValueMap(
		p(blockutils.MobHeadTypeSkeleton, ids.SKELETON_SKULL),
		p(blockutils.MobHeadTypeWitherSkeleton, ids.WITHER_SKELETON_SKULL),
		p(blockutils.MobHeadTypeZombie, ids.ZOMBIE_HEAD),
		p(blockutils.MobHeadTypePlayer, ids.PLAYER_HEAD),
		p(blockutils.MobHeadTypeCreeper, ids.CREEPER_HEAD),
		p(blockutils.MobHeadTypeDragon, ids.DRAGON_HEAD),
		p(blockutils.MobHeadTypePiglin, ids.PIGLIN_HEAD),
	)
	m.FroglightType = newValueMap(
		p(blockutils.FroglightTypeOchre, ids.OCHRE_FROGLIGHT),
		p(blockutils.FroglightTypePearlescent, ids.PEARLESCENT_FROGLIGHT),
		p(blockutils.FroglightTypeVerdant, ids.VERDANT_FROGLIGHT),
	)
	m.DirtType = newValueMap(
		p(blockutils.DirtTypeNormal, ids.DIRT),
		p(blockutils.DirtTypeCoarse, ids.COARSE_DIRT),
		p(blockutils.DirtTypeRooted, ids.DIRT_WITH_ROOTS),
	)
	//state value mappings
	m.DripleafState = newValueMap(
		p(blockutils.DripleafStateStable, ids.BIG_DRIPLEAF_TILT_NONE),
		p(blockutils.DripleafStateUnstable, ids.BIG_DRIPLEAF_TILT_UNSTABLE),
		p(blockutils.DripleafStatePartialTilt, ids.BIG_DRIPLEAF_TILT_PARTIAL_TILT),
		p(blockutils.DripleafStateFullTilt, ids.BIG_DRIPLEAF_TILT_FULL_TILT),
	)
	m.BellAttachmentType = newValueMap(
		p(blockutils.BellAttachmentTypeCeiling, ids.ATTACHMENT_HANGING),
		p(blockutils.BellAttachmentTypeFloor, ids.ATTACHMENT_STANDING),
		p(blockutils.BellAttachmentTypeOneWall, ids.ATTACHMENT_SIDE),
		p(blockutils.BellAttachmentTypeTwoWalls, ids.ATTACHMENT_MULTIPLE),
	)
	m.LeverFacing = newValueMap(
		p(blockutils.LeverFacingUpAxisX, ids.LEVER_DIRECTION_UP_EAST_WEST),
		p(blockutils.LeverFacingUpAxisZ, ids.LEVER_DIRECTION_UP_NORTH_SOUTH),
		p(blockutils.LeverFacingDownAxisX, ids.LEVER_DIRECTION_DOWN_EAST_WEST),
		p(blockutils.LeverFacingDownAxisZ, ids.LEVER_DIRECTION_DOWN_NORTH_SOUTH),
		p(blockutils.LeverFacingNorth, ids.LEVER_DIRECTION_NORTH),
		p(blockutils.LeverFacingEast, ids.LEVER_DIRECTION_EAST),
		p(blockutils.LeverFacingSouth, ids.LEVER_DIRECTION_SOUTH),
		p(blockutils.LeverFacingWest, ids.LEVER_DIRECTION_WEST),
	)
	m.MushroomBlockType = newValueMap(
		p(blockutils.MushroomBlockTypePores, ids.MUSHROOM_BLOCK_ALL_PORES),
		p(blockutils.MushroomBlockTypeCapNorthwest, ids.MUSHROOM_BLOCK_CAP_NORTHWEST_CORNER),
		p(blockutils.MushroomBlockTypeCapNorth, ids.MUSHROOM_BLOCK_CAP_NORTH_SIDE),
		p(blockutils.MushroomBlockTypeCapNortheast, ids.MUSHROOM_BLOCK_CAP_NORTHEAST_CORNER),
		p(blockutils.MushroomBlockTypeCapWest, ids.MUSHROOM_BLOCK_CAP_WEST_SIDE),
		p(blockutils.MushroomBlockTypeCapMiddle, ids.MUSHROOM_BLOCK_CAP_TOP_ONLY),
		p(blockutils.MushroomBlockTypeCapEast, ids.MUSHROOM_BLOCK_CAP_EAST_SIDE),
		p(blockutils.MushroomBlockTypeCapSouthwest, ids.MUSHROOM_BLOCK_CAP_SOUTHWEST_CORNER),
		p(blockutils.MushroomBlockTypeCapSouth, ids.MUSHROOM_BLOCK_CAP_SOUTH_SIDE),
		p(blockutils.MushroomBlockTypeCapSoutheast, ids.MUSHROOM_BLOCK_CAP_SOUTHEAST_CORNER),
		p(blockutils.MushroomBlockTypeAllCap, ids.MUSHROOM_BLOCK_ALL_CAP),
	).withAliases(
		p(blockutils.MushroomBlockTypeAllCap, 11),
		p(blockutils.MushroomBlockTypeAllCap, 12),
		p(blockutils.MushroomBlockTypeAllCap, 13),
	)
	m.CardinalDirection = newValueMap(
		p(math.North, ids.MC_CARDINAL_DIRECTION_NORTH),
		p(math.South, ids.MC_CARDINAL_DIRECTION_SOUTH),
		p(math.West, ids.MC_CARDINAL_DIRECTION_WEST),
		p(math.East, ids.MC_CARDINAL_DIRECTION_EAST),
	)
	m.BlockFace = newValueMap(
		p(math.Down, ids.MC_BLOCK_FACE_DOWN),
		p(math.Up, ids.MC_BLOCK_FACE_UP),
		p(math.North, ids.MC_BLOCK_FACE_NORTH),
		p(math.South, ids.MC_BLOCK_FACE_SOUTH),
		p(math.West, ids.MC_BLOCK_FACE_WEST),
		p(math.East, ids.MC_BLOCK_FACE_EAST),
	)
	m.PillarAxis = newValueMap(
		p(math.AxisX, ids.PILLAR_AXIS_X),
		p(math.AxisY, ids.PILLAR_AXIS_Y),
		p(math.AxisZ, ids.PILLAR_AXIS_Z),
	)
	m.TorchFacing = newValueMap(
		//TODO: horizontal directions are flipped (MCPE bug: https://bugs.mojang.com/browse/MCPE-152036)
		p(math.West, ids.TORCH_FACING_DIRECTION_EAST),
		p(math.South, ids.TORCH_FACING_DIRECTION_NORTH),
		p(math.North, ids.TORCH_FACING_DIRECTION_SOUTH),
		p(math.Up, ids.TORCH_FACING_DIRECTION_TOP),
		p(math.East, ids.TORCH_FACING_DIRECTION_WEST),
	).withAliases(
		p(math.Up, ids.TORCH_FACING_DIRECTION_UNKNOWN), //should be illegal, but still supported
	)
	m.PortalAxis = newValueMap(
		p(math.AxisX, ids.PORTAL_AXIS_X),
		p(math.AxisZ, ids.PORTAL_AXIS_Z),
	).withAliases(
		p(math.AxisX, ids.PORTAL_AXIS_UNKNOWN),
	)
	m.BambooLeafSize = newValueMap(
		p(block.BambooNoLeaves, ids.BAMBOO_LEAF_SIZE_NO_LEAVES),
		p(block.BambooSmallLeaves, ids.BAMBOO_LEAF_SIZE_SMALL_LEAVES),
		p(block.BambooLargeLeaves, ids.BAMBOO_LEAF_SIZE_LARGE_LEAVES),
	)
	m.HorizontalFacing5Minus = newValueMap(
		p(math.East, 0),
		p(math.West, 1),
		p(math.South, 2),
		p(math.North, 3),
	)
	m.HorizontalFacingSWNE = newValueMap(
		p(math.South, 0),
		p(math.West, 1),
		p(math.North, 2),
		p(math.East, 3),
	)
	m.HorizontalFacingSWNEInverted = newValueMap(
		p(math.North, 0),
		p(math.East, 1),
		p(math.South, 2),
		p(math.West, 3),
	)
	m.HorizontalFacingCoral = newValueMap(
		p(math.West, 0),
		p(math.East, 1),
		p(math.North, 2),
		p(math.South, 3),
	)
	horizontalFacingClassicTable := []pair[math.Facing, int]{
		p(math.North, 2),
		p(math.South, 3),
		p(math.West, 4),
		p(math.East, 5),
	}
	with := func(first []pair[math.Facing, int], rest []pair[math.Facing, int]) []pair[math.Facing, int] {
		return append(append([]pair[math.Facing, int](nil), first...), rest...)
	}
	m.HorizontalFacingClassic = newValueMap(horizontalFacingClassicTable...).withAliases(
		p(math.North, 0), p(math.North, 1), //should be illegal but still technically possible
	)
	m.Facing = newValueMap(with([]pair[math.Facing, int]{p(math.Down, 0), p(math.Up, 1)}, horizontalFacingClassicTable)...)
	//end rods have all the horizontal facing values opposite to classic facing
	m.FacingEndRod = newValueMap(
		p(math.Down, 0),
		p(math.Up, 1),
		p(math.South, 2),
		p(math.North, 3),
		p(math.East, 4),
		p(math.West, 5),
	)
	m.CoralAxis = newValueMap(
		p(math.AxisX, 0),
		p(math.AxisZ, 1),
	)
	//TODO: shitty copy pasta job, we can do this better but this is good enough for now
	m.FacingExceptDown = newValueMap(with([]pair[math.Facing, int]{p(math.Up, 1)}, horizontalFacingClassicTable)...).withAliases(p(math.Up, 0))
	m.FacingExceptUp = newValueMap(with([]pair[math.Facing, int]{p(math.Down, 0)}, horizontalFacingClassicTable)...).withAliases(p(math.Down, 1))
	//In PM, we use Facing::UP to indicate that the stem is not attached to a pumpkin/melon, since this makes the
	//most intuitive sense (the stem is pointing at the sky). However, Bedrock uses the DOWN state for this, which
	//is absurd, and I refuse to make our API similarly absurd.
	m.FacingStem = newValueMap(with([]pair[math.Facing, int]{p(math.Up, 0)}, horizontalFacingClassicTable)...).withAliases(p(math.Up, 1))
	return m
}

// intStringMap is IntFromRawStateMap::string(array_map(strval(...), range($min, $max))).
func intStringMap(min, max int) *ValueMap[int, string] {
	var entries []pair[int, string]
	for i := min; i <= max; i++ {
		entries = append(entries, p(i, strconv.Itoa(i)))
	}
	return newValueMap(entries...)
}
