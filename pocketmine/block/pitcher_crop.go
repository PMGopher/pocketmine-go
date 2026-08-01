package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const pitcherCropMaxAge = 2

// PitcherCrop is a port of pocketmine\block\PitcherCrop.
type PitcherCrop struct {
	Flowable
	AgeComponent
}

func NewPitcherCrop(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *PitcherCrop {
	p := &PitcherCrop{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, AgeComponent: NewAgeComponent(pitcherCropMaxAge)}
	p.Init(p)
	return p
}

func (p *PitcherCrop) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *PitcherCrop) DescribeBlockOnlyState(w runtime.DataDescriber) { p.DescribeAge(w) }

func (p *PitcherCrop) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetSide(math.Down, 1).GetTypeId() == FARMLAND
}

func (p *PitcherCrop) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return p.canBeSupportedAt(blockReplace) && p.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (p *PitcherCrop) OnNearbyBlockChange() {
	if !p.canBeSupportedAt(p.self) {
		if world, err := p.position.GetWorld(); err == nil {
			world.UseBreakOn(p.position.AsVector3())
		}
	} else {
		p.Flowable.OnNearbyBlockChange()
	}
}

func (p *PitcherCrop) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	widthTrim, heightTrim := 5.0, 13.0
	if p.Age != 0 {
		widthTrim, heightTrim = 3.0, 11.0
	}
	return []math.AxisAlignedBB{
		math.OneAABB().
			TrimmedCopy(math.Up, heightTrim/16).
			SquashedCopy(math.AxisX, widthTrim/16).
			SquashedCopy(math.AxisZ, widthTrim/16).
			ExtendedCopy(math.Down, 1.0/16),
	}
}

// grow (PitcherCrop::grow) would either advance Age via BlockEventHelper.Grow, or - at MAX_AGE -
// build a two-block BlockTransaction turning this into a DoublePitcherCrop pair via
// StructureGrowEvent. Both branches need the unported block registry (VanillaBlocks) and
// BlockEventHelper/StructureGrowEvent, so this isn't ported - same category of gap as
// Sapling.grow/TorchflowerCrop's getNextState.

// OnInteract's fertilizer-driven grow needs a Fertilizer item marker, not ported yet - same gap
// documented on Crops/SweetBerryBush/CocoaBlock/Sapling/TorchflowerCrop's OnInteract. Block's
// default OnInteract (return false) already matches this gap, so there's nothing to override
// here.

func (p *PitcherCrop) TicksRandomly() bool { return true }

// OnRandomTick should use CropGrowthHelper.CanGrow then grow() - same gap as
// TorchflowerCrop.OnRandomTick, so this is a no-op for now.
func (p *PitcherCrop) OnRandomTick() {}

// AsItem should return VanillaItems.PITCHER_POD() — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.
