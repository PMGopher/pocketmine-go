package math

import "fmt"

// Facing is a port of pocketmine\math\Facing.
type Facing int

const flagAxisPositive = 1

// most significant 2 bits = axis, least significant bit = is positive direction
const (
	Down  Facing = Facing(AxisY << 1)
	Up    Facing = Facing(AxisY<<1) | flagAxisPositive
	North Facing = Facing(AxisZ << 1)
	South Facing = Facing(AxisZ<<1) | flagAxisPositive
	West  Facing = Facing(AxisX << 1)
	East  Facing = Facing(AxisX<<1) | flagAxisPositive
)

var AllFacing = []Facing{Down, Up, North, South, West, East}

var HorizontalFacing = []Facing{North, South, West, East}

// FacingOffset maps a Facing to its [x, y, z] unit offset.
var FacingOffset = map[Facing][3]int{
	Down:  {0, -1, 0},
	Up:    {0, 1, 0},
	North: {0, 0, -1},
	South: {0, 0, 1},
	West:  {-1, 0, 0},
	East:  {1, 0, 0},
}

var facingClockwise = map[Axis]map[Facing]Facing{
	AxisY: {
		North: East,
		East:  South,
		South: West,
		West:  North,
	},
	AxisZ: {
		Up:   East,
		East: Down,
		Down: West,
		West: Up,
	},
	AxisX: {
		Up:    North,
		North: Down,
		Down:  South,
		South: Up,
	},
}

// FacingAxis returns the axis of the given direction.
func FacingAxis(direction Facing) Axis {
	return Axis(direction >> 1) // shift off positive/negative bit
}

// IsPositive returns whether the direction is facing the positive of its axis.
func IsPositive(direction Facing) bool {
	return int(direction)&flagAxisPositive == flagAxisPositive
}

// Opposite returns the opposite Facing of the specified one.
func Opposite(direction Facing) Facing {
	return Facing(int(direction) ^ flagAxisPositive)
}

// Rotate rotates the given direction around the axis. Panics if the rotation isn't possible
// (mirroring the PHP original's \InvalidArgumentException, since this is a programmer error).
func Rotate(direction Facing, axis Axis, clockwise bool) Facing {
	byDirection, ok := facingClockwise[axis]
	if !ok {
		panic(fmt.Sprintf("Invalid axis %d", axis))
	}
	rotated, ok := byDirection[direction]
	if !ok {
		panic(fmt.Sprintf("Cannot rotate facing %q around axis %q", direction, axis))
	}
	if clockwise {
		return rotated
	}
	return Opposite(rotated)
}

func RotateY(direction Facing, clockwise bool) Facing { return Rotate(direction, AxisY, clockwise) }
func RotateZ(direction Facing, clockwise bool) Facing { return Rotate(direction, AxisZ, clockwise) }
func RotateX(direction Facing, clockwise bool) Facing { return Rotate(direction, AxisX, clockwise) }

// ValidateFacing panics if facing isn't a valid Facing constant (mirroring the PHP original,
// which is used purely for input validation on programmer-controlled values).
func ValidateFacing(facing Facing) {
	for _, f := range AllFacing {
		if f == facing {
			return
		}
	}
	panic(fmt.Sprintf("Invalid direction %d", facing))
}

// String returns a human-readable representation of the Facing direction.
func (f Facing) String() string {
	switch f {
	case Down:
		return "down"
	case Up:
		return "up"
	case North:
		return "north"
	case South:
		return "south"
	case West:
		return "west"
	case East:
		return "east"
	default:
		panic(fmt.Sprintf("Invalid facing %d", int(f)))
	}
}
