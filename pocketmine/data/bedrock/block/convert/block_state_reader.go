package blockconvert

import (
	"fmt"
	"sort"
	"strings"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/data/bedrock"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
	"pocketmine-go/pocketmine/math"
)

// BlockStateReader is a port of pocketmine\data\bedrock\block\convert\BlockStateReader. Errors are
// raised (panics) as *BlockStateDeserializeError, like PHP's exceptions; the deserializer recovers
// them.
type BlockStateReader struct {
	data         bedrock.BlockStateData
	unusedStates map[string]bool
}

func NewBlockStateReader(data bedrock.BlockStateData) *BlockStateReader {
	unused := make(map[string]bool, len(data.States))
	for name := range data.States {
		unused[name] = true
	}
	return &BlockStateReader{data: data, unusedStates: unused}
}

// MissingOrWrongTypeError is a port of BlockStateReader::missingOrWrongTypeException.
func (r *BlockStateReader) MissingOrWrongTypeError(name string, tag any) *BlockStateDeserializeError {
	if tag != nil {
		return deserializeError("Property \"%s\" has unexpected type %T", name, tag)
	}
	return deserializeError("Property \"%s\" is missing", name)
}

// BadValueError is a port of BlockStateReader::badValueException.
func (r *BlockStateReader) BadValueError(name string, stringifiedValue string, reason string) *BlockStateDeserializeError {
	if reason != "" {
		return deserializeError("Property \"%s\" has unexpected value \"%s\" (%s)", name, stringifiedValue, reason)
	}
	return deserializeError("Property \"%s\" has unexpected value \"%s\"", name, stringifiedValue)
}

func (r *BlockStateReader) state(name string) any {
	delete(r.unusedStates, name)
	return r.data.States[name]
}

func (r *BlockStateReader) ReadBool(name string) bool {
	tag := r.state(name)
	switch v := tag.(type) {
	case uint8:
		switch v {
		case 0:
			return false
		case 1:
			return true
		}
		panic(r.BadValueError(name, fmt.Sprint(v), ""))
	case int8:
		switch v {
		case 0:
			return false
		case 1:
			return true
		}
		panic(r.BadValueError(name, fmt.Sprint(v), ""))
	}
	panic(r.MissingOrWrongTypeError(name, tag))
}

func (r *BlockStateReader) ReadInt(name string) int {
	tag := r.state(name)
	if v, ok := tag.(int32); ok {
		return int(v)
	}
	panic(r.MissingOrWrongTypeError(name, tag))
}

func (r *BlockStateReader) ReadBoundedInt(name string, min, max int) int {
	result := r.ReadInt(name)
	if result < min || result > max {
		panic(r.BadValueError(name, fmt.Sprint(result), fmt.Sprintf("Must be inside the range %d ... %d", min, max)))
	}
	return result
}

func (r *BlockStateReader) ReadString(name string) string {
	//TODO: only allow a specific set of values (strings are primarily used for enums)
	tag := r.state(name)
	if v, ok := tag.(string); ok {
		return v
	}
	panic(r.MissingOrWrongTypeError(name, tag))
}

func mapFromString[V comparable](r *BlockStateReader, name string, m *ValueMap[V, string]) V {
	raw := r.ReadString(name)
	value, ok := m.RawToValue(raw)
	if !ok {
		panic(r.BadValueError(name, raw, ""))
	}
	return value
}

func mapFromInt[V comparable](r *BlockStateReader, name string, m *ValueMap[V, int]) V {
	raw := r.ReadInt(name)
	value, ok := m.RawToValue(raw)
	if !ok {
		panic(r.BadValueError(name, fmt.Sprint(raw), ""))
	}
	return value
}

func (r *BlockStateReader) ReadFacingDirection() math.Facing {
	return mapFromInt(r, ids.FACING_DIRECTION, GetValueMappings().Facing)
}

func (r *BlockStateReader) ReadBlockFace() math.Facing {
	return mapFromString(r, ids.MC_BLOCK_FACE, GetValueMappings().BlockFace)
}

func (r *BlockStateReader) ReadFacingFlags() []math.Facing {
	var result []math.Facing
	flags := r.ReadBoundedInt(ids.MULTI_FACE_DIRECTION_BITS, 0, 63)
	for _, e := range []struct {
		flag   int
		facing math.Facing
	}{
		{ids.MULTI_FACE_DIRECTION_FLAG_DOWN, math.Down},
		{ids.MULTI_FACE_DIRECTION_FLAG_UP, math.Up},
		{ids.MULTI_FACE_DIRECTION_FLAG_NORTH, math.North},
		{ids.MULTI_FACE_DIRECTION_FLAG_SOUTH, math.South},
		{ids.MULTI_FACE_DIRECTION_FLAG_WEST, math.West},
		{ids.MULTI_FACE_DIRECTION_FLAG_EAST, math.East},
	} {
		if flags&e.flag != 0 {
			result = append(result, e.facing)
		}
	}
	return result
}

func (r *BlockStateReader) ReadEndRodFacingDirection() math.Facing {
	result := r.ReadFacingDirection()
	if math.FacingAxis(result) != math.AxisY {
		return math.Opposite(result)
	}
	return result
}

func (r *BlockStateReader) ReadHorizontalFacing() math.Facing {
	return mapFromInt(r, ids.FACING_DIRECTION, GetValueMappings().HorizontalFacingClassic)
}

func (r *BlockStateReader) ReadWeirdoHorizontalFacing() math.Facing {
	return mapFromInt(r, ids.WEIRDO_DIRECTION, GetValueMappings().HorizontalFacing5Minus)
}

func (r *BlockStateReader) ReadLegacyHorizontalFacing() math.Facing {
	return mapFromInt(r, ids.DIRECTION, GetValueMappings().HorizontalFacingSWNE)
}

func (r *BlockStateReader) Read5MinusHorizontalFacing() math.Facing {
	return mapFromInt(r, ids.DIRECTION, GetValueMappings().HorizontalFacing5Minus)
}

func (r *BlockStateReader) ReadCardinalHorizontalFacing() math.Facing {
	return mapFromString(r, ids.MC_CARDINAL_DIRECTION, GetValueMappings().CardinalDirection)
}

func (r *BlockStateReader) ReadCoralFacing() math.Facing {
	return mapFromInt(r, ids.CORAL_DIRECTION, GetValueMappings().HorizontalFacingCoral)
}

func (r *BlockStateReader) ReadFacingWithoutDown() math.Facing {
	result := r.ReadFacingDirection()
	if result == math.Down { //shouldn't be legal, but 1.13 allows it
		result = math.Up
	}
	return result
}

func (r *BlockStateReader) ReadFacingWithoutUp() math.Facing {
	result := r.ReadFacingDirection()
	if result == math.Up {
		result = math.Down //shouldn't be legal, but 1.13 allows it
	}
	return result
}

func (r *BlockStateReader) ReadPillarAxis() math.Axis {
	return mapFromString(r, ids.PILLAR_AXIS, GetValueMappings().PillarAxis)
}

func (r *BlockStateReader) ReadSlabPosition() blockutils.SlabType {
	switch rawValue := r.ReadString(ids.MC_VERTICAL_HALF); rawValue {
	case ids.MC_VERTICAL_HALF_BOTTOM:
		return blockutils.SlabTypeBottom
	case ids.MC_VERTICAL_HALF_TOP:
		return blockutils.SlabTypeTop
	default:
		panic(r.BadValueError(ids.MC_VERTICAL_HALF, rawValue, "Invalid slab position"))
	}
}

func (r *BlockStateReader) ReadTorchFacing() math.Facing {
	return mapFromString(r, ids.TORCH_FACING_DIRECTION, GetValueMappings().TorchFacing)
}

func (r *BlockStateReader) ReadBellAttachmentType() blockutils.BellAttachmentType {
	return mapFromString(r, ids.ATTACHMENT, GetValueMappings().BellAttachmentType)
}

// ReadWallConnectionType is a port of BlockStateReader::readWallConnectionType (present false is
// PHP's null).
func (r *BlockStateReader) ReadWallConnectionType(name string) (blockutils.WallConnectionType, bool) {
	//TODO: this looks a bit confusing due to use of EAST, but the values are the same for all connections
	//we need to find a better way to auto-generate the constant names when they are reused
	//for now, using these constants is better than nothing since it still gives static analysability
	switch t := r.ReadString(name); t {
	case ids.WALL_CONNECTION_TYPE_EAST_NONE:
		return 0, false
	case ids.WALL_CONNECTION_TYPE_EAST_SHORT:
		return blockutils.WallConnectionTypeShort, true
	case ids.WALL_CONNECTION_TYPE_EAST_TALL:
		return blockutils.WallConnectionTypeTall, true
	default:
		panic(r.BadValueError(name, t, ""))
	}
}

// Ignored is a port of BlockStateReader::ignored.
func (r *BlockStateReader) Ignored(name string) {
	if _, ok := r.data.States[name]; ok {
		delete(r.unusedStates, name)
	} else {
		panic(r.MissingOrWrongTypeError(name, nil))
	}
}

// Todo is a port of BlockStateReader::todo.
func (r *BlockStateReader) Todo(name string) { r.Ignored(name) }

// CheckUnreadProperties is a port of BlockStateReader::checkUnreadProperties.
func (r *BlockStateReader) CheckUnreadProperties() {
	if len(r.unusedStates) > 0 {
		names := make([]string, 0, len(r.unusedStates))
		for name := range r.unusedStates {
			names = append(names, name)
		}
		sort.Strings(names)
		panic(deserializeError("Unread properties: %s", strings.Join(names, ", ")))
	}
}
