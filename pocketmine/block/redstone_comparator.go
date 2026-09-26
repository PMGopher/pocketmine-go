package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// RedstoneComparator is a port of pocketmine\block\RedstoneComparator.
//
// Redstone functionality is a TODO in the PHP original too - only state/placement/collision and
// the signal strength stored in the Comparator tile are implemented upstream.
type RedstoneComparator struct {
	Flowable
	HorizontalFacingComponent
	AnalogRedstoneSignalEmitterComponent
	PoweredByRedstoneComponent

	SubtractMode bool
}

func NewRedstoneComparator(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *RedstoneComparator {
	r := &RedstoneComparator{
		Flowable:                  Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	r.Init(r)
	return r
}

func (r *RedstoneComparator) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RedstoneComparator) DescribeBlockOnlyState(w runtime.DataDescriber) {
	r.DescribeHorizontalFacing(w)
	w.Bool(&r.SubtractMode)
	w.Bool(&r.Powered)
}

func (r *RedstoneComparator) IsSubtractMode() bool { return r.SubtractMode }

func (r *RedstoneComparator) SetSubtractMode(isSubtractMode bool) { r.SubtractMode = isSubtractMode }

func (r *RedstoneComparator) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 7.0/8)}
}

func (r *RedstoneComparator) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		r.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return r.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (r *RedstoneComparator) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	r.SubtractMode = !r.SubtractMode
	if world, err := r.position.GetWorld(); err == nil {
		if err := world.SetBlock(r.position, r.self); err != nil {
			panic(err)
		}
	}
	return true
}

func (r *RedstoneComparator) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Down) != blockutils.SupportTypeNone
}

func (r *RedstoneComparator) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return r.canBeSupportedAt(blockReplace) && r.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (r *RedstoneComparator) OnNearbyBlockChange() {
	if !r.canBeSupportedAt(r.self) {
		if world, err := r.position.GetWorld(); err == nil {
			world.UseBreakOn(r.position.AsVector3())
		}
	} else {
		r.Flowable.OnNearbyBlockChange()
	}
}

// ReadStateFromWorld is a port of RedstoneComparator::readStateFromWorld.
func (r *RedstoneComparator) ReadStateFromWorld() Behavior {
	r.Block.ReadStateFromWorld()
	if t, ok := r.tileAt(); ok {
		if comparator, ok := t.(*tile.Comparator); ok {
			r.SignalStrength = comparator.GetSignalStrength()
		}
	}
	return r.self
}

// WriteStateToWorld is a port of RedstoneComparator::writeStateToWorld.
func (r *RedstoneComparator) WriteStateToWorld() {
	r.Block.WriteStateToWorld()
	if t, ok := r.tileAt(); ok {
		if comparator, ok := t.(*tile.Comparator); ok {
			comparator.SetSignalStrength(r.SignalStrength)
		}
	}
}
