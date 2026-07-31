package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// HorizontalFacing is a port of pocketmine\block\utils\HorizontalFacing.
type HorizontalFacing interface {
	GetFacing() math.Facing
	SetFacing(facing math.Facing)
}

// HorizontalFacingComponent is a port of pocketmine\block\utils\HorizontalFacingTrait.
type HorizontalFacingComponent struct {
	Facing math.Facing
}

func NewHorizontalFacingComponent() HorizontalFacingComponent {
	return HorizontalFacingComponent{Facing: math.North}
}

func (h *HorizontalFacingComponent) DescribeHorizontalFacing(w runtime.DataDescriber) {
	w.HorizontalFacing(&h.Facing)
}

func (h *HorizontalFacingComponent) GetFacing() math.Facing { return h.Facing }

// SetFacing panics if facing isn't horizontal, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (h *HorizontalFacingComponent) SetFacing(facing math.Facing) {
	axis := math.FacingAxis(facing)
	if axis != math.AxisX && axis != math.AxisZ {
		panic("Facing must be horizontal")
	}
	h.Facing = facing
}
