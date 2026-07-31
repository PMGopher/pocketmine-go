package math

import (
	"fmt"
	stdmath "math"
)

// Vector2 is a port of pocketmine\math\Vector2. It's a plain value type in Go, so every method
// naturally returns a new value rather than needing PHP's discipline of "public fields, but
// only ever construct new instances" to stay effectively immutable.
type Vector2 struct {
	X, Y float64
}

func NewVector2(x, y float64) Vector2 { return Vector2{x, y} }

func (v Vector2) FloorX() int { return int(stdmath.Floor(v.X)) }
func (v Vector2) FloorY() int { return int(stdmath.Floor(v.Y)) }

func (v Vector2) Add(x, y float64) Vector2         { return Vector2{v.X + x, v.Y + y} }
func (v Vector2) AddVector(o Vector2) Vector2      { return v.Add(o.X, o.Y) }
func (v Vector2) Subtract(x, y float64) Vector2    { return v.Add(-x, -y) }
func (v Vector2) SubtractVector(o Vector2) Vector2 { return v.Add(-o.X, -o.Y) }

func (v Vector2) Ceil() Vector2  { return Vector2{stdmath.Ceil(v.X), stdmath.Ceil(v.Y)} }
func (v Vector2) Floor() Vector2 { return Vector2{stdmath.Floor(v.X), stdmath.Floor(v.Y)} }
func (v Vector2) Round() Vector2 { return Vector2{stdmath.Round(v.X), stdmath.Round(v.Y)} }
func (v Vector2) Abs() Vector2   { return Vector2{stdmath.Abs(v.X), stdmath.Abs(v.Y)} }

func (v Vector2) Multiply(n float64) Vector2 { return Vector2{v.X * n, v.Y * n} }
func (v Vector2) Divide(n float64) Vector2   { return Vector2{v.X / n, v.Y / n} }

func (v Vector2) Distance(o Vector2) float64 { return stdmath.Sqrt(v.DistanceSquared(o)) }
func (v Vector2) DistanceSquared(o Vector2) float64 {
	dx, dy := v.X-o.X, v.Y-o.Y
	return dx*dx + dy*dy
}

func (v Vector2) Length() float64        { return stdmath.Sqrt(v.LengthSquared()) }
func (v Vector2) LengthSquared() float64 { return v.X*v.X + v.Y*v.Y }

func (v Vector2) Normalize() Vector2 {
	len := v.LengthSquared()
	if len > 0 {
		return v.Divide(stdmath.Sqrt(len))
	}
	return Vector2{0, 0}
}

func (v Vector2) Dot(o Vector2) float64 { return v.X*o.X + v.Y*o.Y }

func (v Vector2) String() string { return fmt.Sprintf("Vector2(x=%v,y=%v)", v.X, v.Y) }
