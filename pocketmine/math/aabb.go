package math

import (
	"fmt"
	stdmath "math"
)

// AxisAlignedBB is a port of pocketmine\math\AxisAlignedBB.
//
// Where the PHP original mutates in place and returns $this for chaining (expand, offset,
// contract, extend, trim, stretch, squash), this keeps both a mutating pointer-receiver method
// and a "*Copy" value-receiver method — Go's value semantics make the copy variant trivial (the
// receiver is already an independent copy), so the mutating variant is implemented in terms of it.
type AxisAlignedBB struct {
	MinX, MinY, MinZ, MaxX, MaxY, MaxZ float64
}

func NewAxisAlignedBB(minX, minY, minZ, maxX, maxY, maxZ float64) (AxisAlignedBB, error) {
	if minX > maxX {
		return AxisAlignedBB{}, fmt.Errorf("minX %v is larger than maxX %v", minX, maxX)
	}
	if minY > maxY {
		return AxisAlignedBB{}, fmt.Errorf("minY %v is larger than maxY %v", minY, maxY)
	}
	if minZ > maxZ {
		return AxisAlignedBB{}, fmt.Errorf("minZ %v is larger than maxZ %v", minZ, maxZ)
	}
	return AxisAlignedBB{minX, minY, minZ, maxX, maxY, maxZ}, nil
}

// OneAABB returns a 1x1x1 bounding box starting at grid position 0,0,0.
func OneAABB() AxisAlignedBB {
	return AxisAlignedBB{0, 0, 0, 1, 1, 1}
}

// AddCoord returns a new AxisAlignedBB extended by the specified X, Y and Z. If each of X, Y and
// Z are positive, the relevant max bound is increased; if negative, the relevant min bound is
// decreased.
func (bb AxisAlignedBB) AddCoord(x, y, z float64) AxisAlignedBB {
	switch {
	case x < 0:
		bb.MinX += x
	case x > 0:
		bb.MaxX += x
	}
	switch {
	case y < 0:
		bb.MinY += y
	case y > 0:
		bb.MaxY += y
	}
	switch {
	case z < 0:
		bb.MinZ += z
	case z > 0:
		bb.MaxZ += z
	}
	return bb
}

func (bb AxisAlignedBB) ExpandedCopy(x, y, z float64) AxisAlignedBB {
	bb.MinX -= x
	bb.MinY -= y
	bb.MinZ -= z
	bb.MaxX += x
	bb.MaxY += y
	bb.MaxZ += z
	return bb
}

// Expand outsets the bounds of bb in place by the specified X, Y and Z.
func (bb *AxisAlignedBB) Expand(x, y, z float64) *AxisAlignedBB {
	*bb = bb.ExpandedCopy(x, y, z)
	return bb
}

func (bb AxisAlignedBB) OffsetCopy(x, y, z float64) AxisAlignedBB {
	bb.MinX += x
	bb.MinY += y
	bb.MinZ += z
	bb.MaxX += x
	bb.MaxY += y
	bb.MaxZ += z
	return bb
}

// Offset shifts bb in place by the given X, Y and Z.
func (bb *AxisAlignedBB) Offset(x, y, z float64) *AxisAlignedBB {
	*bb = bb.OffsetCopy(x, y, z)
	return bb
}

// OffsetTowardsCopy offsets bb in the given direction by the specified distance.
func (bb AxisAlignedBB) OffsetTowardsCopy(face Facing, distance float64) AxisAlignedBB {
	offset, ok := FacingOffset[face]
	if !ok {
		panic(fmt.Sprintf("Invalid Facing %d", face))
	}
	return bb.OffsetCopy(float64(offset[0])*distance, float64(offset[1])*distance, float64(offset[2])*distance)
}

func (bb *AxisAlignedBB) OffsetTowards(face Facing, distance float64) *AxisAlignedBB {
	*bb = bb.OffsetTowardsCopy(face, distance)
	return bb
}

func (bb AxisAlignedBB) ContractedCopy(x, y, z float64) AxisAlignedBB {
	bb.MinX += x
	bb.MinY += y
	bb.MinZ += z
	bb.MaxX -= x
	bb.MaxY -= y
	bb.MaxZ -= z
	return bb
}

func (bb *AxisAlignedBB) Contract(x, y, z float64) *AxisAlignedBB {
	*bb = bb.ContractedCopy(x, y, z)
	return bb
}

// ExtendedCopy extends the AABB in the given direction. Negative distance pulls the face in,
// positive pushes it out. Panics on an invalid face, mirroring the PHP original's
// InvalidArgumentException for what is a programmer error (a bad Facing constant).
func (bb AxisAlignedBB) ExtendedCopy(face Facing, distance float64) AxisAlignedBB {
	switch face {
	case Down:
		bb.MinY -= distance
	case Up:
		bb.MaxY += distance
	case North:
		bb.MinZ -= distance
	case South:
		bb.MaxZ += distance
	case West:
		bb.MinX -= distance
	case East:
		bb.MaxX += distance
	default:
		panic(fmt.Sprintf("Invalid face %d", face))
	}
	return bb
}

func (bb *AxisAlignedBB) Extend(face Facing, distance float64) *AxisAlignedBB {
	*bb = bb.ExtendedCopy(face, distance)
	return bb
}

// TrimmedCopy is the inverse of ExtendedCopy: positive distance pulls the face in.
func (bb AxisAlignedBB) TrimmedCopy(face Facing, distance float64) AxisAlignedBB {
	return bb.ExtendedCopy(face, -distance)
}

func (bb *AxisAlignedBB) Trim(face Facing, distance float64) *AxisAlignedBB {
	*bb = bb.TrimmedCopy(face, distance)
	return bb
}

// StretchedCopy increases the dimension of the AABB along the given axis. Negative distance
// reduces width, positive increases it.
func (bb AxisAlignedBB) StretchedCopy(axis Axis, distance float64) AxisAlignedBB {
	switch axis {
	case AxisY:
		bb.MinY -= distance
		bb.MaxY += distance
	case AxisZ:
		bb.MinZ -= distance
		bb.MaxZ += distance
	case AxisX:
		bb.MinX -= distance
		bb.MaxX += distance
	default:
		panic(fmt.Sprintf("Invalid axis %d", axis))
	}
	return bb
}

func (bb *AxisAlignedBB) Stretch(axis Axis, distance float64) *AxisAlignedBB {
	*bb = bb.StretchedCopy(axis, distance)
	return bb
}

// SquashedCopy is the inverse of StretchedCopy.
func (bb AxisAlignedBB) SquashedCopy(axis Axis, distance float64) AxisAlignedBB {
	return bb.StretchedCopy(axis, -distance)
}

func (bb *AxisAlignedBB) Squash(axis Axis, distance float64) *AxisAlignedBB {
	*bb = bb.SquashedCopy(axis, distance)
	return bb
}

func (bb AxisAlignedBB) CalculateXOffset(other AxisAlignedBB, x float64) float64 {
	if other.MaxY <= bb.MinY || other.MinY >= bb.MaxY {
		return x
	}
	if other.MaxZ <= bb.MinZ || other.MinZ >= bb.MaxZ {
		return x
	}
	if x > 0 && other.MaxX <= bb.MinX {
		if x1 := bb.MinX - other.MaxX; x1 < x {
			x = x1
		}
	} else if x < 0 && other.MinX >= bb.MaxX {
		if x2 := bb.MaxX - other.MinX; x2 > x {
			x = x2
		}
	}
	return x
}

func (bb AxisAlignedBB) CalculateYOffset(other AxisAlignedBB, y float64) float64 {
	if other.MaxX <= bb.MinX || other.MinX >= bb.MaxX {
		return y
	}
	if other.MaxZ <= bb.MinZ || other.MinZ >= bb.MaxZ {
		return y
	}
	if y > 0 && other.MaxY <= bb.MinY {
		if y1 := bb.MinY - other.MaxY; y1 < y {
			y = y1
		}
	} else if y < 0 && other.MinY >= bb.MaxY {
		if y2 := bb.MaxY - other.MinY; y2 > y {
			y = y2
		}
	}
	return y
}

func (bb AxisAlignedBB) CalculateZOffset(other AxisAlignedBB, z float64) float64 {
	if other.MaxX <= bb.MinX || other.MinX >= bb.MaxX {
		return z
	}
	if other.MaxY <= bb.MinY || other.MinY >= bb.MaxY {
		return z
	}
	if z > 0 && other.MaxZ <= bb.MinZ {
		if z1 := bb.MinZ - other.MaxZ; z1 < z {
			z = z1
		}
	} else if z < 0 && other.MinZ >= bb.MaxZ {
		if z2 := bb.MaxZ - other.MinZ; z2 > z {
			z = z2
		}
	}
	return z
}

// IntersectsWith returns whether any part of the specified AABB intersects with bb.
func (bb AxisAlignedBB) IntersectsWith(other AxisAlignedBB, epsilon float64) bool {
	if other.MaxX-bb.MinX > epsilon && bb.MaxX-other.MinX > epsilon {
		if other.MaxY-bb.MinY > epsilon && bb.MaxY-other.MinY > epsilon {
			return other.MaxZ-bb.MinZ > epsilon && bb.MaxZ-other.MinZ > epsilon
		}
	}
	return false
}

// IsVectorInside returns whether the given vector is strictly within the bounds of bb on all axes.
func (bb AxisAlignedBB) IsVectorInside(v Vector3) bool {
	if v.X <= bb.MinX || v.X >= bb.MaxX {
		return false
	}
	if v.Y <= bb.MinY || v.Y >= bb.MaxY {
		return false
	}
	return v.Z > bb.MinZ && v.Z < bb.MaxZ
}

func (bb AxisAlignedBB) GetAverageEdgeLength() float64 {
	return (bb.MaxX - bb.MinX + bb.MaxY - bb.MinY + bb.MaxZ - bb.MinZ) / 3
}

func (bb AxisAlignedBB) GetXLength() float64 { return bb.MaxX - bb.MinX }
func (bb AxisAlignedBB) GetYLength() float64 { return bb.MaxY - bb.MinY }
func (bb AxisAlignedBB) GetZLength() float64 { return bb.MaxZ - bb.MinZ }

func (bb AxisAlignedBB) IsCube(epsilon float64) bool {
	xLen, yLen, zLen := bb.GetXLength(), bb.GetYLength(), bb.GetZLength()
	return stdmath.Abs(xLen-yLen) < epsilon && stdmath.Abs(yLen-zLen) < epsilon
}

func (bb AxisAlignedBB) GetVolume() float64 {
	return (bb.MaxX - bb.MinX) * (bb.MaxY - bb.MinY) * (bb.MaxZ - bb.MinZ)
}

func (bb AxisAlignedBB) IsVectorInYZ(v Vector3) bool {
	return v.Y >= bb.MinY && v.Y <= bb.MaxY && v.Z >= bb.MinZ && v.Z <= bb.MaxZ
}
func (bb AxisAlignedBB) IsVectorInXZ(v Vector3) bool {
	return v.X >= bb.MinX && v.X <= bb.MaxX && v.Z >= bb.MinZ && v.Z <= bb.MaxZ
}
func (bb AxisAlignedBB) IsVectorInXY(v Vector3) bool {
	return v.X >= bb.MinX && v.X <= bb.MaxX && v.Y >= bb.MinY && v.Y <= bb.MaxY
}

// CalculateIntercept performs a ray-trace and finds the point on bb's edge nearest pos1 that the
// ray-trace collided with. ok is false if no colliding point was found.
func (bb AxisAlignedBB) CalculateIntercept(pos1, pos2 Vector3) (result RayTraceResult, ok bool) {
	type candidate struct {
		face Facing
		v    Vector3
		ok   bool
	}
	v1, ok1 := pos1.GetIntermediateWithXValue(pos2, bb.MinX)
	v2, ok2 := pos1.GetIntermediateWithXValue(pos2, bb.MaxX)
	v3, ok3 := pos1.GetIntermediateWithYValue(pos2, bb.MinY)
	v4, ok4 := pos1.GetIntermediateWithYValue(pos2, bb.MaxY)
	v5, ok5 := pos1.GetIntermediateWithZValue(pos2, bb.MinZ)
	v6, ok6 := pos1.GetIntermediateWithZValue(pos2, bb.MaxZ)

	candidates := []candidate{
		{West, v1, ok1 && bb.IsVectorInYZ(v1)},
		{East, v2, ok2 && bb.IsVectorInYZ(v2)},
		{Down, v3, ok3 && bb.IsVectorInXZ(v3)},
		{Up, v4, ok4 && bb.IsVectorInXZ(v4)},
		{North, v5, ok5 && bb.IsVectorInXY(v5)},
		{South, v6, ok6 && bb.IsVectorInXY(v6)},
	}

	found := false
	var bestDistance float64
	var bestVector Vector3
	var bestFace Facing
	for _, c := range candidates {
		if !c.ok {
			continue
		}
		d := pos1.DistanceSquared(c.v)
		if !found || d < bestDistance {
			found = true
			bestDistance = d
			bestVector = c.v
			bestFace = c.face
		}
	}

	if !found {
		return RayTraceResult{}, false
	}
	return RayTraceResult{BB: bb, HitFace: bestFace, HitVector: bestVector}, true
}

func (bb AxisAlignedBB) String() string {
	return fmt.Sprintf("AxisAlignedBB(%v, %v, %v, %v, %v, %v)", bb.MinX, bb.MinY, bb.MinZ, bb.MaxX, bb.MaxY, bb.MaxZ)
}
