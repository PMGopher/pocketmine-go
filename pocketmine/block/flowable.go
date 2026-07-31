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

// Liquid is a forward-compatible marker for pocketmine\block\Liquid — declared here since it's
// only needed for CanBePlacedAt's `instanceof Liquid` check below, matching the local-interface
// pattern used elsewhere for not-yet-ported types (block.Tile, block.Item, etc.). The future
// Liquid base block type just needs to satisfy this trivially for the check to keep working.
type Liquid interface {
	IsLiquid() bool
}

func (f *Flowable) CanBeFlowedInto() bool { return true }

func (f *Flowable) IsSolid() bool { return false }

func (f *Flowable) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	if f.self.CanBeFlowedInto() {
		if _, ok := blockReplace.(Liquid); ok {
			return false
		}
	}
	return f.Block.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (f *Flowable) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (f *Flowable) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
