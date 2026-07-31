package math

import "fmt"

// Axis is a port of pocketmine\math\Axis.
type Axis int

const (
	AxisY Axis = 0
	AxisZ Axis = 1
	AxisX Axis = 2
)

// String returns a human-readable representation of the axis.
func (a Axis) String() string {
	switch a {
	case AxisY:
		return "y"
	case AxisZ:
		return "z"
	case AxisX:
		return "x"
	default:
		panic(fmt.Sprintf("Invalid axis %d", a))
	}
}
