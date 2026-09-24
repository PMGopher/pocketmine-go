package object

import (
	"github.com/go-gl/mathgl/mgl32"

	"pocketmine-go/pocketmine/math"
)

func vec32(v math.Vector3) mgl32.Vec3 {
	return mgl32.Vec3{float32(v.X), float32(v.Y), float32(v.Z)}
}
