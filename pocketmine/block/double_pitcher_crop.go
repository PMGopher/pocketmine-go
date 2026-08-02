package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/event"
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

// grow is a port of DoublePitcherCrop::grow.
func (d *DoublePitcherCrop) grow(player Player) bool {
	if d.Age >= doublePitcherCropMaxAge {
		return false
	}

	var bottomBlock, topBlock Behavior
	if d.Top {
		bottomBlock = d.self.(blockGeometry).GetSide(math.Down, 1)
		topBlock = d.self
	} else {
		bottomBlock = d.self
		topBlock = d.self.(blockGeometry).GetSide(math.Up, 1)
	}
	if topBlock.GetTypeId() != AIR && !topBlock.(blockGeometry).HasSameTypeId(d.self) {
		return false
	}

	world, err := d.position.GetWorld()
	if err != nil {
		return false
	}

	newAge := d.Age + 1
	tx := NewBlockTransaction(world)
	bottom := d.self.Clone().(*DoublePitcherCrop)
	bottom.Age = newAge
	bottom.SetTop(false)
	top := d.self.Clone().(*DoublePitcherCrop)
	top.Age = newAge
	top.SetTop(true)
	tx.AddBlock(bottomBlock.GetPosition(), bottom)
	tx.AddBlock(topBlock.GetPosition(), top)

	ev := &StructureGrowEvent{Block: bottomBlock, Transaction: tx, Player: player}
	event.Call(ev)
	return !ev.IsCancelled() && tx.Apply()
}

// OnInteract's fertilizer-driven grow needs a Fertilizer item marker, not ported yet - same gap
// documented on PitcherCrop/TorchflowerCrop/Sapling's OnInteract. Block's default OnInteract
// (return false) already matches this gap, so there's nothing to override here.

// TicksRandomly is a port of DoublePitcherCrop::ticksRandomly - only the bottom half grows.
func (d *DoublePitcherCrop) TicksRandomly() bool { return d.Age < doublePitcherCropMaxAge && !d.Top }

// OnRandomTick is a port of DoublePitcherCrop::onRandomTick - only the bottom half of the plant
// can grow randomly.
func (d *DoublePitcherCrop) OnRandomTick() {
	if CropGrowthCanGrow(d.self) && !d.Top {
		d.grow(nil)
	}
}

// GetDropsForCompatibleTool should return VanillaBlocks.PITCHER_PLANT().AsItem() once mature, or
// VanillaItems.PITCHER_POD() otherwise - needs the unported block registry and item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.

// AsItem should return VanillaItems.PITCHER_POD() — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.
