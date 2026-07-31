package runtime

import (
	"fmt"

	"pocketmine-go/pocketmine/math"
)

// Writer is a port of pocketmine\data\runtime\RuntimeDataWriter.
type Writer struct {
	maxBits int
	value   int
	offset  int
}

func NewWriter(maxBits int) *Writer {
	return &Writer{maxBits: maxBits}
}

func (w *Writer) WriteInt(bits int, value int) {
	if w.offset+bits > w.maxBits {
		panic(fmt.Sprintf("bit buffer cannot be larger than %d bits (already have %d bits)", w.maxBits, w.offset))
	}
	if value&(^0<<bits) != 0 {
		panic(fmt.Sprintf("value %d does not fit into %d bits", value, bits))
	}
	w.value |= value << w.offset
	w.offset += bits
}

func (w *Writer) Int(bits int, value *int) { w.WriteInt(bits, *value) }

func (w *Writer) writeBoundedIntAuto(min, max, value int) {
	if value < min || value > max {
		panic(fmt.Sprintf("value %d is outside the range %d - %d", value, min, max))
	}
	bits := boundedIntAutoBits(min, max)
	w.WriteInt(bits, value-min)
}

func (w *Writer) BoundedIntAuto(min, max int, value *int) { w.writeBoundedIntAuto(min, max, *value) }

func (w *Writer) writeBool(value bool) {
	if value {
		w.WriteInt(1, 1)
	} else {
		w.WriteInt(1, 0)
	}
}

func (w *Writer) Bool(value *bool) { w.writeBool(*value) }

func (w *Writer) HorizontalFacing(facing *math.Facing) {
	var v int
	switch *facing {
	case math.North:
		v = 0
	case math.East:
		v = 1
	case math.South:
		v = 2
	case math.West:
		v = 3
	default:
		panic(fmt.Sprintf("invalid horizontal facing %d", *facing))
	}
	w.WriteInt(2, v)
}

func (w *Writer) FacingFlags(faces *[]math.Facing) {
	unique := map[math.Facing]bool{}
	for _, f := range *faces {
		unique[f] = true
	}
	for _, facing := range math.AllFacing {
		w.writeBool(unique[facing])
	}
}

func (w *Writer) HorizontalFacingFlags(faces *[]math.Facing) {
	unique := map[math.Facing]bool{}
	for _, f := range *faces {
		unique[f] = true
	}
	for _, facing := range math.HorizontalFacing {
		w.writeBool(unique[facing])
	}
}

func (w *Writer) Facing(facing *math.Facing) {
	var v int
	switch *facing {
	case math.Down:
		v = 0
	case math.Up:
		v = 1
	case math.North:
		v = 2
	case math.South:
		v = 3
	case math.West:
		v = 4
	case math.East:
		v = 5
	default:
		panic(fmt.Sprintf("invalid facing %d", *facing))
	}
	w.WriteInt(3, v)
}

func (w *Writer) FacingExcept(facing *math.Facing, except math.Facing) {
	w.Facing(facing)
}

func (w *Writer) Axis(axis *math.Axis) {
	var v int
	switch *axis {
	case math.AxisX:
		v = 0
	case math.AxisZ:
		v = 1
	case math.AxisY:
		v = 2
	default:
		panic(fmt.Sprintf("invalid axis %d", *axis))
	}
	w.WriteInt(2, v)
}

func (w *Writer) HorizontalAxis(axis *math.Axis) {
	var v int
	switch *axis {
	case math.AxisX:
		v = 0
	case math.AxisZ:
		v = 1
	default:
		panic(fmt.Sprintf("invalid horizontal axis %d", *axis))
	}
	w.WriteInt(1, v)
}

func (w *Writer) GetValue() int  { return w.value }
func (w *Writer) GetOffset() int { return w.offset }

var _ DataDescriber = (*Writer)(nil)
