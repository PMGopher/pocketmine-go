package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// AnyFacing is a port of pocketmine\block\utils\AnyFacing.
type AnyFacing interface {
	GetFacing() math.Facing
	SetFacing(facing math.Facing)
}

// FacingComponent is a port of pocketmine\block\utils\AnyFacingTrait's state (the facing field
// and its accessors). PHP traits are copy-pasted per using class rather than inherited, so unlike
// Behavior's methods there's no virtual-dispatch pitfall in embedding this directly — but
// concrete types with extra state (like Button's pressed flag) still write their own
// DescribeBlockOnlyState, calling DescribeFacing(w) plus whatever else they need, exactly as the
// PHP subclass fully replaces (rather than extends) the trait's describeBlockOnlyState.
type FacingComponent struct {
	Facing math.Facing
}

func NewFacingComponent() FacingComponent { return FacingComponent{Facing: math.Down} }

func (f *FacingComponent) DescribeFacing(w runtime.DataDescriber) { w.Facing(&f.Facing) }

func (f *FacingComponent) GetFacing() math.Facing { return f.Facing }

// SetFacing validates and sets the new facing.
//
// The original PHP (AnyFacingTrait::setFacing) calls Facing::validate($this->facing) — validating
// the block's CURRENT facing rather than the incoming parameter, which validates nothing useful
// (the current value was already validated when it was set) and doesn't catch a bad argument.
// HorizontalFacingTrait's equivalent check validates the incoming value; that's clearly the
// intent here too, so this validates the parameter instead of replicating the apparent copy-paste
// bug.
func (f *FacingComponent) SetFacing(facing math.Facing) {
	math.ValidateFacing(facing)
	f.Facing = facing
}
