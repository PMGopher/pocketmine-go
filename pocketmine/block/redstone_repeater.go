package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	redstoneRepeaterMinDelay = 1
	redstoneRepeaterMaxDelay = 4
)

// RedstoneRepeater is a port of pocketmine\block\RedstoneRepeater.
//
// Redstone functionality is a TODO in the PHP original too - only state/placement/collision are
// implemented upstream either.
type RedstoneRepeater struct {
	Flowable
	HorizontalFacingComponent
	PoweredByRedstoneComponent

	Delay int
}

func NewRedstoneRepeater(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *RedstoneRepeater {
	r := &RedstoneRepeater{
		Flowable:                  Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
		Delay:                     redstoneRepeaterMinDelay,
	}
	r.Init(r)
	return r
}

func (r *RedstoneRepeater) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RedstoneRepeater) DescribeBlockOnlyState(w runtime.DataDescriber) {
	r.DescribeHorizontalFacing(w)
	delay := r.Delay
	w.BoundedIntAuto(redstoneRepeaterMinDelay, redstoneRepeaterMaxDelay, &delay)
	r.Delay = delay
	w.Bool(&r.Powered)
}

func (r *RedstoneRepeater) GetDelay() int { return r.Delay }

func (r *RedstoneRepeater) SetDelay(delay int) {
	if delay < redstoneRepeaterMinDelay || delay > redstoneRepeaterMaxDelay {
		panic("Delay must be in range 1 ... 4")
	}
	r.Delay = delay
}

func (r *RedstoneRepeater) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 7.0/8)}
}

func (r *RedstoneRepeater) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		r.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return r.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (r *RedstoneRepeater) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	r.Delay++
	if r.Delay > redstoneRepeaterMaxDelay {
		r.Delay = redstoneRepeaterMinDelay
	}
	if world, err := r.position.GetWorld(); err == nil {
		if err := world.SetBlock(r.position, r.self); err != nil {
			panic(err)
		}
	}
	return true
}

func (r *RedstoneRepeater) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Down) != blockutils.SupportTypeNone
}

func (r *RedstoneRepeater) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return r.canBeSupportedAt(blockReplace) && r.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (r *RedstoneRepeater) OnNearbyBlockChange() {
	if !r.canBeSupportedAt(r.self) {
		if world, err := r.position.GetWorld(); err == nil {
			world.UseBreakOn(r.position.AsVector3())
		}
	} else {
		r.Flowable.OnNearbyBlockChange()
	}
}
