package math

import (
	"fmt"
	stdmath "math"
)

func FloorFloat(n float64) int {
	i := int(n)
	if n >= float64(i) {
		return i
	}
	return i - 1
}

func CeilFloat(n float64) int {
	i := int(n)
	if n <= float64(i) {
		return i
	}
	return i + 1
}

// SolveQuadratic solves a quadratic equation with the given coefficients, returning up to two
// real solutions.
func SolveQuadratic(a, b, c float64) ([]float64, error) {
	if a == 0.0 {
		return nil, fmt.Errorf("coefficient a cannot be 0")
	}
	discriminant := b*b - 4*a*c
	switch {
	case discriminant > 0:
		sqrtDiscriminant := stdmath.Sqrt(discriminant)
		return []float64{
			(-b + sqrtDiscriminant) / (2 * a),
			(-b - sqrtDiscriminant) / (2 * a),
		}, nil
	case discriminant == 0.0:
		return []float64{-b / (2 * a)}, nil
	default:
		return nil, nil
	}
}
