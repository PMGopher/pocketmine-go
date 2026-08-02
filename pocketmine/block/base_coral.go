package block

import (
	"math/rand"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// BaseCoral is a port of pocketmine\block\BaseCoral. Like Crops/Stem, this isn't meant to be
// instantiated directly - a concrete leaf type (Coral) must embed it and implement Clone.
type BaseCoral struct {
	Transparent
	CoralComponent
}

func (b *BaseCoral) DescribeBlockItemState(w runtime.DataDescriber) { b.DescribeCoral(w) }

func (b *BaseCoral) OnNearbyBlockChange() {
	if !b.Dead {
		if world, err := b.position.GetWorld(); err == nil {
			world.ScheduleDelayedBlockUpdate(b.position.AsVector3(), 40+rand.Intn(161))
		}
	}
}

func (b *BaseCoral) isCoveredWithWater() bool {
	geo := b.self.(blockGeometry)
	for _, face := range math.AllFacing {
		if geo.GetSide(face, 1).GetTypeId() == WATER {
			return true
		}
	}
	// TODO: check water inside the block itself (not supported on the API yet)
	return false
}

// OnScheduledUpdate is a port of BaseCoral::onScheduledUpdate.
func (b *BaseCoral) OnScheduledUpdate() {
	if !b.Dead && !b.isCoveredWithWater() {
		deadClone := b.self.Clone()
		deadClone.(CoralMaterial).SetDead(true)
		Die(b.self, deadClone)
	}
}

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap).
func (b *BaseCoral) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (b *BaseCoral) IsAffectedBySilkTouch() bool { return true }

func (b *BaseCoral) IsSolid() bool { return false }

func (b *BaseCoral) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (b *BaseCoral) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
