package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// Flowable is a port of pocketmine\block\Flowable.
//
// "Flowable" blocks are destroyed if water flows into the same space as the block. These blocks
// usually don't have any collision boxes, and can't provide support for other blocks.
type Flowable struct {
	Transparent
}

func (f *Flowable) CanBeFlowedInto() bool { return true }

func (f *Flowable) IsSolid() bool { return false }

func (f *Flowable) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	if f.self.CanBeFlowedInto() {
		if _, ok := blockReplace.(liquidBaser); ok {
			return false
		}
	}
	return f.Block.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (f *Flowable) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (f *Flowable) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
