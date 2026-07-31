package runtime

import (
	"fmt"

	"pocketmine-go/pocketmine/math"
)

// Reader is a port of pocketmine\data\runtime\RuntimeDataReader.
type Reader struct {
	maxBits int
	value   int
	offset  int
}

func NewReader(maxBits int, value int) *Reader {
	return &Reader{maxBits: maxBits, value: value}
}

func (r *Reader) ReadInt(bits int) (int, error) {
	bitsLeft := r.maxBits - r.offset
	if bits > bitsLeft {
		return 0, fmt.Errorf("no bits left in buffer (need %d, have %d)", bits, bitsLeft)
	}
	value := (r.value >> r.offset) & ^(^0 << bits)
	r.offset += bits
	return value, nil
}

func (r *Reader) Int(bits int, value *int) {
	v, err := r.ReadInt(bits)
	if err != nil {
		panic(err)
	}
	*value = v
}

func (r *Reader) readBoundedIntAuto(min, max int) (int, error) {
	bits := boundedIntAutoBits(min, max)
	raw, err := r.ReadInt(bits)
	if err != nil {
		return 0, err
	}
	result := raw + min
	if result < min || result > max {
		return 0, &InvalidSerializedRuntimeDataError{Message: fmt.Sprintf("value is outside the range %d - %d", min, max)}
	}
	return result, nil
}

func (r *Reader) BoundedIntAuto(min, max int, value *int) {
	v, err := r.readBoundedIntAuto(min, max)
	if err != nil {
		panic(err)
	}
	*value = v
}

func (r *Reader) readBool() bool {
	v, err := r.ReadInt(1)
	if err != nil {
		panic(err)
	}
	return v == 1
}

func (r *Reader) Bool(value *bool) { *value = r.readBool() }

func (r *Reader) HorizontalFacing(facing *math.Facing) {
	v, err := r.ReadInt(2)
	if err != nil {
		panic(err)
	}
	switch v {
	case 0:
		*facing = math.North
	case 1:
		*facing = math.East
	case 2:
		*facing = math.South
	case 3:
		*facing = math.West
	default:
		panic("unreachable")
	}
}

func (r *Reader) FacingFlags(faces *[]math.Facing) {
	var result []math.Facing
	for _, facing := range math.AllFacing {
		if r.readBool() {
			result = append(result, facing)
		}
	}
	*faces = result
}

func (r *Reader) HorizontalFacingFlags(faces *[]math.Facing) {
	var result []math.Facing
	for _, facing := range math.HorizontalFacing {
		if r.readBool() {
			result = append(result, facing)
		}
	}
	*faces = result
}

func (r *Reader) Facing(facing *math.Facing) {
	v, err := r.ReadInt(3)
	if err != nil {
		panic(err)
	}
	switch v {
	case 0:
		*facing = math.Down
	case 1:
		*facing = math.Up
	case 2:
		*facing = math.North
	case 3:
		*facing = math.South
	case 4:
		*facing = math.West
	case 5:
		*facing = math.East
	default:
		panic(&InvalidSerializedRuntimeDataError{Message: "invalid facing value"})
	}
}

func (r *Reader) FacingExcept(facing *math.Facing, except math.Facing) {
	var result math.Facing
	r.Facing(&result)
	if result == except {
		panic(&InvalidSerializedRuntimeDataError{Message: "illegal facing value"})
	}
	*facing = result
}

func (r *Reader) Axis(axis *math.Axis) {
	v, err := r.ReadInt(2)
	if err != nil {
		panic(err)
	}
	switch v {
	case 0:
		*axis = math.AxisX
	case 1:
		*axis = math.AxisZ
	case 2:
		*axis = math.AxisY
	default:
		panic(&InvalidSerializedRuntimeDataError{Message: "invalid axis value"})
	}
}

func (r *Reader) HorizontalAxis(axis *math.Axis) {
	v, err := r.ReadInt(1)
	if err != nil {
		panic(err)
	}
	switch v {
	case 0:
		*axis = math.AxisX
	case 1:
		*axis = math.AxisZ
	default:
		panic("unreachable")
	}
}

func (r *Reader) GetOffset() int { return r.offset }

var _ DataDescriber = (*Reader)(nil)
