package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// PillarRotation is a port of pocketmine\block\utils\PillarRotation.
type PillarRotation interface {
	GetAxis() math.Axis
	SetAxis(axis math.Axis)
}

// PillarRotationComponent is a port of pocketmine\block\utils\PillarRotationTrait's state.
// Place() isn't provided here: PHP traits are copy-pasted per using class rather than inherited,
// so (like FacingComponent/AnyFacingTrait) there's no code to share across the `parent::place()`
// call — each concrete pillar type calls SetAxisFromFace in its own Place() before delegating to
// its embedded Block.
type PillarRotationComponent struct {
	Axis math.Axis
}

func NewPillarRotationComponent() PillarRotationComponent {
	return PillarRotationComponent{Axis: math.AxisY}
}

func (p *PillarRotationComponent) DescribeAxis(w runtime.DataDescriber) { w.Axis(&p.Axis) }

func (p *PillarRotationComponent) GetAxis() math.Axis { return p.Axis }

func (p *PillarRotationComponent) SetAxis(axis math.Axis) { p.Axis = axis }

// SetAxisFromFace mirrors PillarRotationTrait::place()'s `$this->axis = Facing::axis($face);` step.
func (p *PillarRotationComponent) SetAxisFromFace(face math.Facing) { p.Axis = math.FacingAxis(face) }
