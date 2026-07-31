package math

// RayTraceResult is a port of pocketmine\math\RayTraceResult: a ray trace collision with an
// AxisAlignedBB.
type RayTraceResult struct {
	BB        AxisAlignedBB
	HitFace   Facing
	HitVector Vector3
}
