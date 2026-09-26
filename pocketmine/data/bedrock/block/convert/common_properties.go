package blockconvert

import (
	"sync"

	blockutils "pocketmine-go/pocketmine/block/utils"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
	"pocketmine-go/pocketmine/math"
)

// CommonProperties is a port of pocketmine\data\bedrock\block\convert\property\CommonProperties:
// property descriptors shared by many blocks.
//
// ID components (the idComponents lists of FlattenedIdModel) are []any holding strings and
// StringProperty values, like PHP's list<string|StringProperty>.
type CommonProperties struct {
	BlockFace                    Property
	PillarAxis                   Property
	TorchFacing                  Property
	HorizontalFacingCardinal     Property
	HorizontalFacingSWNE         Property
	HorizontalFacingSWNEInverted Property
	HorizontalFacingClassic      Property
	AnyFacingClassic             Property
	MultiFacingFlags             Property
	FloorSignLikeRotation        Property
	AnalogRedstoneSignal         Property
	CropAgeMax7                  Property
	DoublePlantHalf              Property
	LiquidData                   Property
	Lit                          Property
	DummyCardinalDirection       Property
	DummyPillarAxis              Property
	DyeColorIdInfix              StringProperty
	LitIdInfix                   StringProperty
	SlabIdInfix                  StringProperty
	SlabPositionProperty         StringProperty

	CoralIdPrefixes   []any
	CopperIdPrefixes  []any
	FurnaceIdPrefixes []any
	LiquidIdPrefixes  []any
	WoodIdPrefixes    []any

	ButtonProperties              []Property
	CampfireProperties            []Property
	DoorProperties                []Property
	FenceGateProperties           []Property
	ItemFrameProperties           []Property
	SimplePressurePlateProperties []Property
	StairProperties               []Property
	StemProperties                []Property
	TrapdoorProperties            []Property
	WallProperties                []Property
}

var (
	commonProperties     *CommonProperties
	commonPropertiesOnce sync.Once
)

// GetCommonProperties is CommonProperties::getInstance().
func GetCommonProperties() *CommonProperties {
	commonPropertiesOnce.Do(func() { commonProperties = newCommonProperties() })
	return commonProperties
}

func hfGet(b facingHolder) math.Facing         { return b.GetFacing() }
func hfSet(b facingHolder, facing math.Facing) { b.SetFacing(facing) }

func newCommonProperties() *CommonProperties {
	vm := GetValueMappings()
	c := &CommonProperties{}

	c.HorizontalFacingCardinal = NewValueFromStringProperty(ids.MC_CARDINAL_DIRECTION, vm.CardinalDirection, hfGet, hfSet)
	c.BlockFace = NewValueFromStringProperty(ids.MC_BLOCK_FACE, vm.BlockFace, hfGet, hfSet)
	c.PillarAxis = NewValueFromStringProperty(ids.PILLAR_AXIS, vm.PillarAxis,
		func(b axisHolder) math.Axis { return b.GetAxis() },
		func(b axisHolder, v math.Axis) { b.SetAxis(v) },
	)
	c.TorchFacing = NewValueFromStringProperty(ids.TORCH_FACING_DIRECTION, vm.TorchFacing, hfGet, hfSet)
	c.HorizontalFacingSWNE = NewValueFromIntProperty(ids.DIRECTION, vm.HorizontalFacingSWNE, hfGet, hfSet)
	c.HorizontalFacingSWNEInverted = NewValueFromIntProperty(ids.DIRECTION, vm.HorizontalFacingSWNEInverted, hfGet, hfSet)
	c.HorizontalFacingClassic = NewValueFromIntProperty(ids.FACING_DIRECTION, vm.HorizontalFacingClassic, hfGet, hfSet)
	c.AnyFacingClassic = NewValueFromIntProperty(ids.FACING_DIRECTION, vm.Facing, hfGet, hfSet)
	c.MultiFacingFlags = NewValueSetFromIntProperty(
		ids.MULTI_FACE_DIRECTION_BITS,
		newValueMap(
			p(math.Down, ids.MULTI_FACE_DIRECTION_FLAG_DOWN),
			p(math.Up, ids.MULTI_FACE_DIRECTION_FLAG_UP),
			p(math.North, ids.MULTI_FACE_DIRECTION_FLAG_NORTH),
			p(math.South, ids.MULTI_FACE_DIRECTION_FLAG_SOUTH),
			p(math.West, ids.MULTI_FACE_DIRECTION_FLAG_WEST),
			p(math.East, ids.MULTI_FACE_DIRECTION_FLAG_EAST),
		),
		func(b multiFacingHolder) []math.Facing { return b.GetFaces() },
		func(b multiFacingHolder, v []math.Facing) { b.SetFaces(v) },
	)
	c.FloorSignLikeRotation = NewIntProperty(ids.GROUND_SIGN_DIRECTION, 0, 15,
		func(b rotationHolder) int { return b.GetRotation() },
		func(b rotationHolder, v int) { b.SetRotation(v) },
	)
	c.AnalogRedstoneSignal = NewIntProperty(ids.REDSTONE_SIGNAL, 0, 15,
		func(b analogRedstoneSignalEmitter) int { return b.GetOutputSignalStrength() },
		func(b analogRedstoneSignalEmitter, v int) { b.SetOutputSignalStrength(v) },
	)
	c.CropAgeMax7 = NewIntProperty(ids.GROWTH, 0, 7,
		func(b ageable) int { return b.GetAge() },
		func(b ageable, v int) { b.SetAge(v) },
	)
	c.DoublePlantHalf = NewBoolProperty(ids.UPPER_BLOCK_BIT,
		func(b topHalf) bool { return b.IsTop() },
		func(b topHalf, v bool) { b.SetTop(v) },
	)
	fallingFlag := ids.LIQUID_FALLING_FLAG
	c.LiquidData = NewIntProperty(ids.LIQUID_DEPTH, 0, 15,
		func(b liquidLike) int {
			v := b.GetDecay()
			if b.IsFalling() {
				v |= fallingFlag
			}
			return v
		},
		func(b liquidLike, v int) {
			b.SetDecay(v &^ fallingFlag)
			b.SetFalling(v&fallingFlag != 0)
		},
	)
	c.Lit = NewBoolProperty(ids.LIT,
		func(b lightable) bool { return b.IsLit() },
		func(b lightable, v bool) { b.SetLit(v) },
	)
	c.DummyCardinalDirection = NewDummyProperty(ids.MC_CARDINAL_DIRECTION, ids.MC_CARDINAL_DIRECTION_SOUTH)
	c.DummyPillarAxis = NewDummyProperty(ids.PILLAR_AXIS, ids.PILLAR_AXIS_Y)
	c.DyeColorIdInfix = NewValueFromStringProperty("color", vm.DyeColor,
		func(b colored) blockutils.DyeColor { return b.GetColor() },
		func(b colored, v blockutils.DyeColor) { b.SetColor(v) },
	)
	c.LitIdInfix = NewBoolFromStringProperty("lit", "", "lit_",
		func(b lightable) bool { return b.IsLit() },
		func(b lightable, v bool) { b.SetLit(v) },
	)
	c.SlabIdInfix = NewBoolFromStringProperty("double", "", "double_",
		func(b slabLike) bool { return b.GetSlabType() == blockutils.SlabTypeDouble },
		//we don't know this is actually a bottom slab yet but we don't have enough information to set the
		//correct type in this handler
		//BOTTOM serves as a signal value for the state deserializer to decide whether to ignore the
		//upper_block_bit property
		func(b slabLike, v bool) {
			if v {
				b.SetSlabType(blockutils.SlabTypeDouble)
			} else {
				b.SetSlabType(blockutils.SlabTypeBottom)
			}
		},
	)
	c.SlabPositionProperty = NewBoolFromStringProperty(ids.MC_VERTICAL_HALF, ids.MC_VERTICAL_HALF_BOTTOM, ids.MC_VERTICAL_HALF_TOP,
		func(b slabLike) bool { return b.GetSlabType() == blockutils.SlabTypeTop },
		//Ignore the value for double slabs (should be set by ID component before this is reached)
		func(b slabLike, v bool) {
			if b.GetSlabType() != blockutils.SlabTypeDouble {
				if v {
					b.SetSlabType(blockutils.SlabTypeTop)
				} else {
					b.SetSlabType(blockutils.SlabTypeBottom)
				}
			}
		},
	)
	c.CoralIdPrefixes = []any{
		"minecraft:",
		NewBoolFromStringProperty("dead", "", "dead_",
			func(b coralMaterial) bool { return b.IsDead() },
			func(b coralMaterial, v bool) { b.SetDead(v) },
		),
		NewValueFromStringProperty("type", newValueMap(
			p(blockutils.CoralTypeTube, "tube"),
			p(blockutils.CoralTypeBrain, "brain"),
			p(blockutils.CoralTypeBubble, "bubble"),
			p(blockutils.CoralTypeFire, "fire"),
			p(blockutils.CoralTypeHorn, "horn"),
		),
			func(b coralMaterial) blockutils.CoralType { return b.GetCoralType() },
			func(b coralMaterial, v blockutils.CoralType) { b.SetCoralType(v) },
		),
	}
	c.CopperIdPrefixes = []any{
		"minecraft:",
		NewBoolFromStringProperty("waxed", "", "waxed_",
			func(b copperMaterial) bool { return b.IsWaxed() },
			func(b copperMaterial, v bool) { b.SetWaxed(v) },
		),
		NewValueFromStringProperty("oxidation", newValueMap(
			p(blockutils.CopperOxidationNone, ""),
			p(blockutils.CopperOxidationExposed, "exposed_"),
			p(blockutils.CopperOxidationWeathered, "weathered_"),
			p(blockutils.CopperOxidationOxidized, "oxidized_"),
		),
			func(b copperMaterial) blockutils.CopperOxidation { return b.GetOxidation() },
			func(b copperMaterial, v blockutils.CopperOxidation) { b.SetOxidation(v) },
		),
	}
	c.FurnaceIdPrefixes = []any{"minecraft:", c.LitIdInfix}
	c.LiquidIdPrefixes = []any{
		"minecraft:",
		NewBoolFromStringProperty("still", "flowing_", "",
			func(b liquidLike) bool { return b.IsStill() },
			func(b liquidLike, v bool) { b.SetStill(v) },
		),
	}
	c.WoodIdPrefixes = []any{
		"minecraft:",
		NewBoolFromStringProperty("stripped", "", "stripped_",
			func(b woodLike) bool { return b.IsStripped() },
			func(b woodLike, v bool) { b.SetStripped(v) },
		),
	}
	c.ButtonProperties = []Property{
		c.AnyFacingClassic,
		NewBoolProperty(ids.BUTTON_PRESSED_BIT,
			func(b buttonLike) bool { return b.IsPressed() },
			func(b buttonLike, v bool) { b.SetPressed(v) },
		),
	}
	c.CampfireProperties = []Property{
		c.HorizontalFacingCardinal,
		NewInvertedBoolProperty(ids.EXTINGUISHED,
			func(b lightable) bool { return b.IsLit() },
			func(b lightable, v bool) { b.SetLit(v) },
		),
	}
	//TODO: check if these need any special treatment to get the appropriate data to both halves of the door
	c.DoorProperties = []Property{
		NewBoolProperty(ids.UPPER_BLOCK_BIT, func(b doorLike) bool { return b.IsTop() }, func(b doorLike, v bool) { b.SetTop(v) }),
		NewBoolProperty(ids.DOOR_HINGE_BIT, func(b doorLike) bool { return b.IsHingeRight() }, func(b doorLike, v bool) { b.SetHingeRight(v) }),
		NewBoolProperty(ids.OPEN_BIT, func(b doorLike) bool { return b.IsOpen() }, func(b doorLike, v bool) { b.SetOpen(v) }),
		NewValueFromStringProperty(
			ids.MC_CARDINAL_DIRECTION,
			newValueMap(
				//a door facing "east" is actually facing north - thanks mojang
				p(math.North, ids.MC_CARDINAL_DIRECTION_EAST),
				p(math.East, ids.MC_CARDINAL_DIRECTION_SOUTH),
				p(math.South, ids.MC_CARDINAL_DIRECTION_WEST),
				p(math.West, ids.MC_CARDINAL_DIRECTION_NORTH),
			),
			hfGet, hfSet,
		),
	}
	c.FenceGateProperties = []Property{
		NewBoolProperty(ids.IN_WALL_BIT, func(b fenceGateLike) bool { return b.IsInWall() }, func(b fenceGateLike, v bool) { b.SetInWall(v) }),
		NewBoolProperty(ids.OPEN_BIT, func(b fenceGateLike) bool { return b.IsOpen() }, func(b fenceGateLike, v bool) { b.SetOpen(v) }),
		c.HorizontalFacingCardinal,
	}
	c.ItemFrameProperties = []Property{
		NewDummyProperty(ids.ITEM_FRAME_PHOTO_BIT, false), //TODO: not sure what the point of this is
		NewBoolProperty(ids.ITEM_FRAME_MAP_BIT, func(b itemFrameLike) bool { return b.HasMap() }, func(b itemFrameLike, v bool) { b.SetHasMap(v) }),
		c.AnyFacingClassic,
	}
	c.SimplePressurePlateProperties = []Property{
		//TODO: not sure what the deal is here ... seems like a mojang bug / artifact of bad implementation?
		//best to keep this separate from weighted plates anyway...
		NewIntProperty(ids.REDSTONE_SIGNAL, 0, 15,
			func(b buttonLike) int {
				if b.IsPressed() {
					return 15
				}
				return 0
			},
			func(b buttonLike, v int) { b.SetPressed(v != 0) },
		),
	}
	c.StairProperties = []Property{
		NewBoolProperty(ids.UPSIDE_DOWN_BIT, func(b stairLike) bool { return b.IsUpsideDown() }, func(b stairLike, v bool) { b.SetUpsideDown(v) }),
		NewValueFromIntProperty(ids.WEIRDO_DIRECTION, vm.HorizontalFacing5Minus, hfGet, hfSet),
		StairCornerProperty,
	}
	c.StemProperties = []Property{
		NewValueFromIntProperty(ids.FACING_DIRECTION, vm.FacingStem, hfGet, hfSet),
		c.CropAgeMax7,
	}
	c.TrapdoorProperties = []Property{
		//this uses the same values as stairs, but the state is named differently
		NewValueFromIntProperty(ids.DIRECTION, vm.HorizontalFacing5Minus, hfGet, hfSet),
		NewBoolProperty(ids.UPSIDE_DOWN_BIT, func(b topHalf) bool { return b.IsTop() }, func(b topHalf, v bool) { b.SetTop(v) }),
		NewBoolProperty(ids.OPEN_BIT, func(b openable) bool { return b.IsOpen() }, func(b openable, v bool) { b.SetOpen(v) }),
	}
	wallProperties := []Property{
		NewBoolProperty(ids.WALL_POST_BIT, func(b wallLike) bool { return b.IsPost() }, func(b wallLike, v bool) { b.SetPost(v) }),
	}
	for _, e := range []struct {
		facing    math.Facing
		stateName string
	}{
		{math.North, ids.WALL_CONNECTION_TYPE_NORTH},
		{math.South, ids.WALL_CONNECTION_TYPE_SOUTH},
		{math.West, ids.WALL_CONNECTION_TYPE_WEST},
		{math.East, ids.WALL_CONNECTION_TYPE_EAST},
	} {
		facing := e.facing
		wallProperties = append(wallProperties, NewValueFromStringProperty(
			e.stateName,
			wallConnectionTypeShimMap,
			func(b wallLike) wallConnectionTypeShim {
				return serializeWallConnectionTypeShim(b.GetConnection(facing))
			},
			func(b wallLike, v wallConnectionTypeShim) {
				t, present := v.deserialize()
				b.SetConnection(facing, t, present)
			},
		))
	}
	c.WallProperties = wallProperties
	return c
}

// wallConnectionTypeShim is a port of property\WallConnectionTypeShim: WallConnectionType plus
// NONE (PHP's null).
type wallConnectionTypeShim int

const (
	wallConnectionTypeShimNone wallConnectionTypeShim = iota
	wallConnectionTypeShimShort
	wallConnectionTypeShimTall
)

var wallConnectionTypeShimMap = newValueMap(
	p(wallConnectionTypeShimNone, ids.WALL_CONNECTION_TYPE_EAST_NONE),
	p(wallConnectionTypeShimShort, ids.WALL_CONNECTION_TYPE_EAST_SHORT),
	p(wallConnectionTypeShimTall, ids.WALL_CONNECTION_TYPE_EAST_TALL),
)

func (s wallConnectionTypeShim) deserialize() (blockutils.WallConnectionType, bool) {
	switch s {
	case wallConnectionTypeShimShort:
		return blockutils.WallConnectionTypeShort, true
	case wallConnectionTypeShimTall:
		return blockutils.WallConnectionTypeTall, true
	}
	return 0, false
}

func serializeWallConnectionTypeShim(value blockutils.WallConnectionType, present bool) wallConnectionTypeShim {
	if !present {
		return wallConnectionTypeShimNone
	}
	if value == blockutils.WallConnectionTypeTall {
		return wallConnectionTypeShimTall
	}
	return wallConnectionTypeShimShort
}

// The 1.26.50 palette's extra properties (see bedrockblock.MC_CORNER): they aren't part of the
// internal block state (like PocketMine-MP, the stair shape and fence/pane connections are worked
// out from the neighbours when the block is read from the world), so the neutral value is written
// and the value read from a saved state is ignored.
var (
	StairCornerProperty = NewDummyProperty(ids.MC_CORNER, ids.MC_CORNER_NONE)

	ConnectionProperties = []Property{
		NewDummyProperty(ids.MC_CONNECTION_EAST, false),
		NewDummyProperty(ids.MC_CONNECTION_NORTH, false),
		NewDummyProperty(ids.MC_CONNECTION_SOUTH, false),
		NewDummyProperty(ids.MC_CONNECTION_WEST, false),
	}
)
