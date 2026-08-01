package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const doublePitcherCropMaxAge = 1

// DoublePitcherCrop is a port of pocketmine\block\DoublePitcherCrop.
type DoublePitcherCrop struct {
	DoublePlant
	AgeComponent
}

func NewDoublePitcherCrop(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DoublePitcherCrop {
	d := &DoublePitcherCrop{
		DoublePlant:  DoublePlant{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}},
		AgeComponent: NewAgeComponent(doublePitcherCropMaxAge),
	}
	d.Init(d)
	return d
}

func (d *DoublePitcherCrop) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *DoublePitcherCrop) DescribeBlockOnlyState(w runtime.DataDescriber) {
	d.DoublePlant.DescribeBlockOnlyState(w)
	d.DescribeAge(w)
}

func (d *DoublePitcherCrop) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	if d.Top {
		return nil
	}
	// the pod exists only in the bottom half of the plant
	return []math.AxisAlignedBB{
		math.OneAABB().
			TrimmedCopy(math.Up, 11.0/16).
			SquashedCopy(math.AxisX, 3.0/16).
			SquashedCopy(math.AxisZ, 3.0/16).
			ExtendedCopy(math.Down, 1.0/16),
	}
}

// grow (DoublePitcherCrop::grow) would advance Age and rebuild both halves via a
// BlockTransaction/StructureGrowEvent - needs the unported block registry and event, same gap as
// PitcherCrop.grow, so it's not ported.

// OnInteract's fertilizer-driven grow needs a Fertilizer item marker, not ported yet - same gap
// documented on PitcherCrop/TorchflowerCrop/Sapling's OnInteract. Block's default OnInteract
// (return false) already matches this gap, so there's nothing to override here.

// TicksRandomly is a port of DoublePitcherCrop::ticksRandomly - only the bottom half grows.
func (d *DoublePitcherCrop) TicksRandomly() bool { return d.Age < doublePitcherCropMaxAge && !d.Top }

// OnRandomTick should use CropGrowthHelper.CanGrow then grow() on the bottom half - same gap as
// PitcherCrop.OnRandomTick, so this is a no-op for now.
func (d *DoublePitcherCrop) OnRandomTick() {}

// GetDropsForCompatibleTool should return VanillaBlocks.PITCHER_PLANT().AsItem() once mature, or
// VanillaItems.PITCHER_POD() otherwise - needs the unported block registry and item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.

// AsItem should return VanillaItems.PITCHER_POD() — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.
