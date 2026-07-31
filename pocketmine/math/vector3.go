package math

import (
	"fmt"
	stdmath "math"
)

// Vector3 is a port of pocketmine\math\Vector3.
//
// PHP's X/Y/Z are typed `float|int` since Vector3 doubles as both a precise entity position and
// a whole-number block coordinate. Go has no such union for primitives, so this uses float64
// uniformly (matching Vector2, which was already float-only in the original) — float64 exactly
// represents every integer value PocketMine ever stores in a coordinate, so nothing is lost;
// callers needing a block coordinate use FloorX/FloorY/FloorZ, exactly as the PHP original does.
type Vector3 struct {
	X, Y, Z float64
}

func NewVector3(x, y, z float64) Vector3 { return Vector3{x, y, z} }

func Vector3Zero() Vector3 { return Vector3{0, 0, 0} }

func (v Vector3) FloorX() int { return int(stdmath.Floor(v.X)) }
func (v Vector3) FloorY() int { return int(stdmath.Floor(v.Y)) }
func (v Vector3) FloorZ() int { return int(stdmath.Floor(v.Z)) }

func (v Vector3) Add(x, y, z float64) Vector3      { return Vector3{v.X + x, v.Y + y, v.Z + z} }
func (v Vector3) AddVector(o Vector3) Vector3      { return v.Add(o.X, o.Y, o.Z) }
func (v Vector3) Subtract(x, y, z float64) Vector3 { return v.Add(-x, -y, -z) }
func (v Vector3) SubtractVector(o Vector3) Vector3 { return v.Add(-o.X, -o.Y, -o.Z) }

func (v Vector3) Multiply(n float64) Vector3 { return Vector3{v.X * n, v.Y * n, v.Z * n} }
func (v Vector3) Divide(n float64) Vector3   { return Vector3{v.X / n, v.Y / n, v.Z / n} }

func (v Vector3) Ceil() Vector3 {
	return Vector3{stdmath.Ceil(v.X), stdmath.Ceil(v.Y), stdmath.Ceil(v.Z)}
}
func (v Vector3) Floor() Vector3 {
	return Vector3{stdmath.Floor(v.X), stdmath.Floor(v.Y), stdmath.Floor(v.Z)}
}

// Round rounds each component to the given decimal precision (half-away-from-zero, matching
// PHP's default PHP_ROUND_HALF_UP mode — the only mode the PHP original's callers ever used).
func (v Vector3) Round(precision int) Vector3 {
	scale := stdmath.Pow(10, float64(precision))
	return Vector3{
		stdmath.Round(v.X*scale) / scale,
		stdmath.Round(v.Y*scale) / scale,
		stdmath.Round(v.Z*scale) / scale,
	}
}

func (v Vector3) Abs() Vector3 {
	return Vector3{stdmath.Abs(v.X), stdmath.Abs(v.Y), stdmath.Abs(v.Z)}
}

// GetSide returns the vector obtained by stepping `step` blocks in the given Facing direction.
func (v Vector3) GetSide(side Facing, step int) Vector3 {
	offset, ok := FacingOffset[side]
	if !ok {
		offset = [3]int{0, 0, 0}
	}
	return v.Add(float64(offset[0]*step), float64(offset[1]*step), float64(offset[2]*step))
}

func (v Vector3) Down(step int) Vector3  { return v.GetSide(Down, step) }
func (v Vector3) Up(step int) Vector3    { return v.GetSide(Up, step) }
func (v Vector3) North(step int) Vector3 { return v.GetSide(North, step) }
func (v Vector3) South(step int) Vector3 { return v.GetSide(South, step) }
func (v Vector3) West(step int) Vector3  { return v.GetSide(West, step) }
func (v Vector3) East(step int) Vector3  { return v.GetSide(East, step) }

// Sides returns the vectors stepped out from this one in all six directions, keyed by Facing.
func (v Vector3) Sides(step int) map[Facing]Vector3 {
	result := make(map[Facing]Vector3, len(AllFacing))
	for _, facing := range AllFacing {
		result[facing] = v.GetSide(facing, step)
	}
	return result
}

// SidesAroundAxis returns the vectors stepped out from this one in every direction except those
// on the given axis.
func (v Vector3) SidesAroundAxis(axis Axis, step int) map[Facing]Vector3 {
	result := map[Facing]Vector3{}
	for _, facing := range AllFacing {
		if FacingAxis(facing) != axis {
			result[facing] = v.GetSide(facing, step)
		}
	}
	return result
}

func (v Vector3) Distance(o Vector3) float64 { return stdmath.Sqrt(v.DistanceSquared(o)) }

func (v Vector3) DistanceSquared(o Vector3) float64 {
	dx, dy, dz := v.X-o.X, v.Y-o.Y, v.Z-o.Z
	return dx*dx + dy*dy + dz*dz
}

// MaxPlainDistanceXZ, MaxPlainDistanceVector2 and MaxPlainDistanceVector3 replace the PHP
// original's single maxPlainDistance(), which overloads on a Vector3|Vector2|float parameter —
// Go has no overloading, so each accepted shape gets its own named function.
func (v Vector3) MaxPlainDistanceXZ(x, z float64) float64 {
	return stdmath.Max(stdmath.Abs(v.X-x), stdmath.Abs(v.Z-z))
}
func (v Vector3) MaxPlainDistanceVector2(o Vector2) float64 { return v.MaxPlainDistanceXZ(o.X, o.Y) }
func (v Vector3) MaxPlainDistanceVector3(o Vector3) float64 { return v.MaxPlainDistanceXZ(o.X, o.Z) }

func (v Vector3) Length() float64        { return stdmath.Sqrt(v.LengthSquared()) }
func (v Vector3) LengthSquared() float64 { return v.X*v.X + v.Y*v.Y + v.Z*v.Z }

func (v Vector3) Normalize() Vector3 {
	len := v.LengthSquared()
	if len > 0 {
		return v.Divide(stdmath.Sqrt(len))
	}
	return Vector3{0, 0, 0}
}

func (v Vector3) Dot(o Vector3) float64 { return v.X*o.X + v.Y*o.Y + v.Z*o.Z }

func (v Vector3) Cross(o Vector3) Vector3 {
	return Vector3{
		v.Y*o.Z - v.Z*o.Y,
		v.Z*o.X - v.X*o.Z,
		v.X*o.Y - v.Y*o.X,
	}
}

func (v Vector3) Equals(o Vector3) bool {
	return v.X == o.X && v.Y == o.Y && v.Z == o.Z
}

// GetIntermediateWithXValue returns the point with the given x value along the line between v
// and o, and whether such a point exists.
func (v Vector3) GetIntermediateWithXValue(o Vector3, x float64) (Vector3, bool) {
	xDiff := o.X - v.X
	if xDiff*xDiff < 0.0000001 {
		return Vector3{}, false
	}
	f := (x - v.X) / xDiff
	if f < 0 || f > 1 {
		return Vector3{}, false
	}
	return Vector3{x, v.Y + (o.Y-v.Y)*f, v.Z + (o.Z-v.Z)*f}, true
}

// GetIntermediateWithYValue returns the point with the given y value along the line between v
// and o, and whether such a point exists.
func (v Vector3) GetIntermediateWithYValue(o Vector3, y float64) (Vector3, bool) {
	yDiff := o.Y - v.Y
	if yDiff*yDiff < 0.0000001 {
		return Vector3{}, false
	}
	f := (y - v.Y) / yDiff
	if f < 0 || f > 1 {
		return Vector3{}, false
	}
	return Vector3{v.X + (o.X-v.X)*f, y, v.Z + (o.Z-v.Z)*f}, true
}

// GetIntermediateWithZValue returns the point with the given z value along the line between v
// and o, and whether such a point exists.
func (v Vector3) GetIntermediateWithZValue(o Vector3, z float64) (Vector3, bool) {
	zDiff := o.Z - v.Z
	if zDiff*zDiff < 0.0000001 {
		return Vector3{}, false
	}
	f := (z - v.Z) / zDiff
	if f < 0 || f > 1 {
		return Vector3{}, false
	}
	return Vector3{v.X + (o.X-v.X)*f, v.Y + (o.Y-v.Y)*f, z}, true
}

func (v Vector3) String() string {
	return fmt.Sprintf("Vector3(x=%v,y=%v,z=%v)", v.X, v.Y, v.Z)
}

// WithComponents returns a Vector3 with the given components, falling back to v's own component
// for any nil pointer. Mirrors PHP's `?float $x, ?float $y, ?float $z` "null means unchanged".
func (v Vector3) WithComponents(x, y, z *float64) Vector3 {
	result := v
	if x != nil {
		result.X = *x
	}
	if y != nil {
		result.Y = *y
	}
	if z != nil {
		result.Z = *z
	}
	return result
}

// MaxComponents returns a new Vector3 taking the maximum of each component across all inputs.
func MaxComponents(first Vector3, rest ...Vector3) Vector3 {
	result := first
	for _, v := range rest {
		result.X = stdmath.Max(result.X, v.X)
		result.Y = stdmath.Max(result.Y, v.Y)
		result.Z = stdmath.Max(result.Z, v.Z)
	}
	return result
}

// MinComponents returns a new Vector3 taking the minimum of each component across all inputs.
func MinComponents(first Vector3, rest ...Vector3) Vector3 {
	result := first
	for _, v := range rest {
		result.X = stdmath.Min(result.X, v.X)
		result.Y = stdmath.Min(result.Y, v.Y)
		result.Z = stdmath.Min(result.Z, v.Z)
	}
	return result
}

func SumVector3(vectors ...Vector3) Vector3 {
	var result Vector3
	for _, v := range vectors {
		result.X += v.X
		result.Y += v.Y
		result.Z += v.Z
	}
	return result
}
