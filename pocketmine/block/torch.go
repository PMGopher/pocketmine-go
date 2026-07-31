package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Torch is a port of pocketmine\block\Torch.
//
// Doesn't reuse FacingComponent: unlike AnyFacingTrait, Torch's setFacing rejects Facing::DOWN
// (a torch can't face downwards), and its state encoding uses facingExcept rather than facing.
type Torch struct {
	Flowable

	Facing math.Facing
}

func NewTorch(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Torch {
	t := &Torch{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, Facing: math.Up}
	t.Init(t)
	return t
}

func (t *Torch) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *Torch) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.FacingExcept(&t.Facing, math.Down)
}

func (t *Torch) GetFacing() math.Facing { return t.Facing }

// SetFacing panics if facing is Down, mirroring the PHP original's \InvalidArgumentException (a
// programmer error at the call site — a torch can't face downwards).
func (t *Torch) SetFacing(facing math.Facing) {
	if facing == math.Down {
		panic("Torch may not face DOWN")
	}
	t.Facing = facing
}

func (t *Torch) GetLightLevel() int { return 14 }

func (t *Torch) OnNearbyBlockChange() {
	if !t.canBeSupportedAt(t.self, math.Opposite(t.Facing)) {
		if world, err := t.position.GetWorld(); err == nil {
			world.UseBreakOn(t.position.AsVector3())
		}
	}
}

func (t *Torch) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if face != math.Down && t.canBeSupportedAt(blockReplace, math.Opposite(face)) {
		t.Facing = face
		return t.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
	}
	for _, side := range []math.Facing{math.South, math.West, math.North, math.East, math.Down} {
		if t.canBeSupportedAt(blockReplace, side) {
			t.Facing = math.Opposite(side)
			return t.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
		}
	}
	return false
}

func (t *Torch) canBeSupportedAt(blk Behavior, face math.Facing) bool {
	support := blk.(blockGeometry).GetAdjacentSupportType(face)
	if face == math.Down {
		return support.HasCenterSupport()
	}
	return support == blockutils.SupportTypeFull
}
