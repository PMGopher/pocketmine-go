package blockconvert

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/data/bedrock"
	ids "pocketmine-go/pocketmine/data/bedrock/block"
	"pocketmine-go/pocketmine/math"
)

// CurrentBlockStateVersion is BlockStateData::CURRENT_VERSION (WorldDataVersions::BLOCK_STATES)
// for the vendored 1.26.50 palette: the version every state in canonical_block_states.nbt has.
const CurrentBlockStateVersion = bedrock.CurrentBlockStateVersion

// BlockStateWriter is a port of pocketmine\data\bedrock\block\convert\BlockStateWriter. Bool
// properties are written as uint8 (ByteTag), int properties as int32 (IntTag).
type BlockStateWriter struct {
	id     string
	states map[string]any
}

func NewBlockStateWriter(id string) *BlockStateWriter {
	return &BlockStateWriter{id: id, states: map[string]any{}}
}

func (w *BlockStateWriter) WriteBool(name string, value bool) *BlockStateWriter {
	var v uint8
	if value {
		v = 1
	}
	w.states[name] = v
	return w
}

func (w *BlockStateWriter) WriteInt(name string, value int) *BlockStateWriter {
	w.states[name] = int32(value)
	return w
}

func (w *BlockStateWriter) WriteString(name string, value string) *BlockStateWriter {
	w.states[name] = value
	return w
}

func mapToString[V comparable](w *BlockStateWriter, name string, m *ValueMap[V, string], value V) *BlockStateWriter {
	return w.WriteString(name, m.ValueToRaw(value))
}

func mapToInt[V comparable](w *BlockStateWriter, name string, m *ValueMap[V, int], value V) *BlockStateWriter {
	return w.WriteInt(name, m.ValueToRaw(value))
}

func (w *BlockStateWriter) WriteFacingDirection(value math.Facing) *BlockStateWriter {
	return mapToInt(w, ids.FACING_DIRECTION, GetValueMappings().Facing, value)
}

func (w *BlockStateWriter) WriteBlockFace(value math.Facing) *BlockStateWriter {
	return mapToString(w, ids.MC_BLOCK_FACE, GetValueMappings().BlockFace, value)
}

func (w *BlockStateWriter) WriteFacingFlags(faces []math.Facing) *BlockStateWriter {
	result := 0
	for _, face := range faces {
		switch face {
		case math.Down:
			result |= ids.MULTI_FACE_DIRECTION_FLAG_DOWN
		case math.Up:
			result |= ids.MULTI_FACE_DIRECTION_FLAG_UP
		case math.North:
			result |= ids.MULTI_FACE_DIRECTION_FLAG_NORTH
		case math.South:
			result |= ids.MULTI_FACE_DIRECTION_FLAG_SOUTH
		case math.West:
			result |= ids.MULTI_FACE_DIRECTION_FLAG_WEST
		case math.East:
			result |= ids.MULTI_FACE_DIRECTION_FLAG_EAST
		default:
			panic("Unhandled face")
		}
	}
	return w.WriteInt(ids.MULTI_FACE_DIRECTION_BITS, result)
}

func (w *BlockStateWriter) WriteEndRodFacingDirection(value math.Facing) *BlockStateWriter {
	//end rods are stupid in bedrock and have everything except up/down the wrong way round
	if math.FacingAxis(value) != math.AxisY {
		value = math.Opposite(value)
	}
	return w.WriteFacingDirection(value)
}

func (w *BlockStateWriter) WriteHorizontalFacing(value math.Facing) *BlockStateWriter {
	return mapToInt(w, ids.FACING_DIRECTION, GetValueMappings().HorizontalFacingClassic, value)
}

func (w *BlockStateWriter) WriteWeirdoHorizontalFacing(value math.Facing) *BlockStateWriter {
	return mapToInt(w, ids.WEIRDO_DIRECTION, GetValueMappings().HorizontalFacing5Minus, value)
}

func (w *BlockStateWriter) WriteLegacyHorizontalFacing(value math.Facing) *BlockStateWriter {
	return mapToInt(w, ids.DIRECTION, GetValueMappings().HorizontalFacingSWNE, value)
}

func (w *BlockStateWriter) Write5MinusHorizontalFacing(value math.Facing) *BlockStateWriter {
	return mapToInt(w, ids.DIRECTION, GetValueMappings().HorizontalFacing5Minus, value)
}

func (w *BlockStateWriter) WriteCardinalHorizontalFacing(value math.Facing) *BlockStateWriter {
	return mapToString(w, ids.MC_CARDINAL_DIRECTION, GetValueMappings().CardinalDirection, value)
}

func (w *BlockStateWriter) WriteCoralFacing(value math.Facing) *BlockStateWriter {
	return mapToInt(w, ids.CORAL_DIRECTION, GetValueMappings().HorizontalFacingCoral, value)
}

func (w *BlockStateWriter) WriteFacingWithoutDown(value math.Facing) *BlockStateWriter {
	if value == math.Down {
		serializeError("Invalid facing DOWN")
	}
	return w.WriteFacingDirection(value)
}

func (w *BlockStateWriter) WriteFacingWithoutUp(value math.Facing) *BlockStateWriter {
	if value == math.Up {
		serializeError("Invalid facing UP")
	}
	return w.WriteFacingDirection(value)
}

func (w *BlockStateWriter) WritePillarAxis(axis math.Axis) *BlockStateWriter {
	return mapToString(w, ids.PILLAR_AXIS, GetValueMappings().PillarAxis, axis)
}

func (w *BlockStateWriter) WriteSlabPosition(slabType blockutils.SlabType) *BlockStateWriter {
	switch slabType {
	case blockutils.SlabTypeTop:
		return w.WriteString(ids.MC_VERTICAL_HALF, ids.MC_VERTICAL_HALF_TOP)
	case blockutils.SlabTypeBottom:
		return w.WriteString(ids.MC_VERTICAL_HALF, ids.MC_VERTICAL_HALF_BOTTOM)
	}
	serializeError("Invalid slab type %v", slabType)
	return w
}

func (w *BlockStateWriter) WriteTorchFacing(facing math.Facing) *BlockStateWriter {
	return mapToString(w, ids.TORCH_FACING_DIRECTION, GetValueMappings().TorchFacing, facing)
}

func (w *BlockStateWriter) WriteBellAttachmentType(attachmentType blockutils.BellAttachmentType) *BlockStateWriter {
	return mapToString(w, ids.ATTACHMENT, GetValueMappings().BellAttachmentType, attachmentType)
}

// WriteWallConnectionType is a port of BlockStateWriter::writeWallConnectionType (present false
// is PHP's null: no connection).
func (w *BlockStateWriter) WriteWallConnectionType(name string, connectionType blockutils.WallConnectionType, present bool) *BlockStateWriter {
	if !present {
		return w.WriteString(name, ids.WALL_CONNECTION_TYPE_EAST_NONE)
	}
	switch connectionType {
	case blockutils.WallConnectionTypeShort:
		return w.WriteString(name, ids.WALL_CONNECTION_TYPE_EAST_SHORT)
	case blockutils.WallConnectionTypeTall:
		return w.WriteString(name, ids.WALL_CONNECTION_TYPE_EAST_TALL)
	}
	panic("unreachable")
}

// GetBlockStateData is a port of BlockStateWriter::getBlockStateData.
func (w *BlockStateWriter) GetBlockStateData() bedrock.BlockStateData {
	states := make(map[string]any, len(w.states))
	for k, v := range w.states {
		states[k] = v
	}
	return bedrock.BlockStateData{Name: w.id, States: states, Version: CurrentBlockStateVersion}
}
