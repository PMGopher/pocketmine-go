package block

import (
	"math"

	runtime "pocketmine-go/pocketmine/data/runtime"
)

const (
	signLikeRotationMin = 0
	signLikeRotationMax = 15
)

// SignLikeRotation is a port of pocketmine\block\utils\SignLikeRotation.
type SignLikeRotation interface {
	GetRotation() int
	SetRotation(rotation int)
}

// SignLikeRotationComponent is a port of pocketmine\block\utils\SignLikeRotationTrait.
type SignLikeRotationComponent struct {
	Rotation int
}

func (s *SignLikeRotationComponent) DescribeRotation(w runtime.DataDescriber) {
	rotation := s.Rotation
	w.BoundedIntAuto(signLikeRotationMin, signLikeRotationMax, &rotation)
	s.Rotation = rotation
}

func (s *SignLikeRotationComponent) GetRotation() int { return s.Rotation }

// SetRotation panics if out of range, mirroring the PHP original's InvalidArgumentException (a
// programmer error at the call site).
func (s *SignLikeRotationComponent) SetRotation(rotation int) {
	if rotation < signLikeRotationMin || rotation > signLikeRotationMax {
		panic("Rotation must be in range 0-15")
	}
	s.Rotation = rotation
}

// SignLikeRotationFromYaw is a port of SignLikeRotationTrait::getRotationFromYaw.
func SignLikeRotationFromYaw(yaw float64) int {
	return int(math.Floor((yaw+180)*16/360+0.5)) & 0xf
}
