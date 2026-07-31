package math

import stdmath "math"

// GetDirection2D returns a unit Vector2 pointing in the given azimuth (radians).
func GetDirection2D(azimuth float64) Vector2 {
	return Vector2{stdmath.Cos(azimuth), stdmath.Sin(azimuth)}
}
