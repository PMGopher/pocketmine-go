package math

import (
	"fmt"
	"iter"
	stdmath "math"
)

// InDirection performs a ray trace from start in the given direction for maxDistance, yielding
// the voxel coordinates it passes through. See BetweenPoints for precise semantics.
func InDirection(start, directionVector Vector3, maxDistance float64) (iter.Seq[Vector3], error) {
	return BetweenPoints(start, start.AddVector(directionVector.Multiply(maxDistance)))
}

// BetweenPoints performs a ray trace between start and end, yielding the voxel coordinates it
// passes through.
//
// The first Vector3 is start.Floor(). Every subsequent Vector3 has a taxicab distance of exactly
// 1 from the previous one; if the ray crosses the intersection of multiple axis boundaries
// directly, the algorithm prefers crossing them in the order Z -> Y -> X.
//
// This is an implementation of the algorithm described at http://www.cse.yorku.ca/~amana/research/grid.pdf
func BetweenPoints(start, end Vector3) (iter.Seq[Vector3], error) {
	directionVector := end.SubtractVector(start).Normalize()
	if directionVector.LengthSquared() <= 0 {
		return nil, fmt.Errorf("start and end points are the same, giving a zero direction vector")
	}

	return func(yield func(Vector3) bool) {
		currentBlock := start.Floor()
		radius := start.Distance(end)

		stepX := signOf(directionVector.X)
		stepY := signOf(directionVector.Y)
		stepZ := signOf(directionVector.Z)

		tMaxX := distanceFactorToBoundary(start.X, directionVector.X)
		tMaxY := distanceFactorToBoundary(start.Y, directionVector.Y)
		tMaxZ := distanceFactorToBoundary(start.Z, directionVector.Z)

		tDeltaX := 0.0
		if directionVector.X != 0.0 {
			tDeltaX = float64(stepX) / directionVector.X
		}
		tDeltaY := 0.0
		if directionVector.Y != 0.0 {
			tDeltaY = float64(stepY) / directionVector.Y
		}
		tDeltaZ := 0.0
		if directionVector.Z != 0.0 {
			tDeltaZ = float64(stepZ) / directionVector.Z
		}

		for {
			if !yield(currentBlock) {
				return
			}

			// tMaxX/Y/Z store the t-value at which we cross a cube boundary on that axis, so the
			// least tMax chooses the closest cube boundary.
			if tMaxX < tMaxY && tMaxX < tMaxZ {
				if tMaxX > radius {
					return
				}
				currentBlock = currentBlock.Add(float64(stepX), 0, 0)
				tMaxX += tDeltaX
			} else if tMaxY < tMaxZ {
				if tMaxY > radius {
					return
				}
				currentBlock = currentBlock.Add(0, float64(stepY), 0)
				tMaxY += tDeltaY
			} else {
				if tMaxZ > radius {
					return
				}
				currentBlock = currentBlock.Add(0, 0, float64(stepZ))
				tMaxZ += tDeltaZ
			}
		}
	}, nil
}

func signOf(v float64) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	default:
		return 0
	}
}

// distanceFactorToBoundary returns the number of times ds must be added to s to change its
// whole-number component — used to decide which direction to move in first when beginning a ray trace.
func distanceFactorToBoundary(s, ds float64) float64 {
	if ds == 0.0 {
		return stdmath.Inf(1)
	}
	if ds < 0 {
		return (s - stdmath.Floor(s)) / -ds
	}
	return (1 - (s - stdmath.Floor(s))) / ds
}
