package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
	"pocketmine-go/pocketmine/math"
)

const DoublePitcherCropMaxAge = 1

// DoublePitcherCrop is a port of pocketmine\block\DoublePitcherCrop.
type DoublePitcherCrop struct {
	DoublePlant
	AgeComponent
}

func NewDoublePitcherCrop(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DoublePitcherCrop {
	d := &DoublePitcherCrop{
		DoublePlant:  DoublePlant{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}},
		AgeComponent: NewAgeComponent(DoublePitcherCropMaxAge),
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
	if d.Age >= DoublePitcherCropMaxAge {
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

	ev := blockevent.NewStructureGrowEvent(bottomBlock, tx, player)
	event.Call(ev)
	return !ev.IsCancelled() && tx.Apply()
}

// OnInteract is a port of DoublePitcherCrop::onInteract (bone meal is the only Fertilizer).
func (d *DoublePitcherCrop) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if item.GetTypeId() == itemTypeIDsBoneMeal && d.grow(player) {
		item.Pop()
		return true
	}
	return false
}

// TicksRandomly is a port of DoublePitcherCrop::ticksRandomly - only the bottom half grows.
func (d *DoublePitcherCrop) TicksRandomly() bool { return d.Age < DoublePitcherCropMaxAge && !d.Top }

// OnRandomTick is a port of DoublePitcherCrop::onRandomTick - only the bottom half of the plant
// can grow randomly.
func (d *DoublePitcherCrop) OnRandomTick() {
	if CropGrowthCanGrow(d.self) && !d.Top {
		d.grow(nil)
	}
}

// GetDropsForCompatibleTool is a port of DoublePitcherCrop::getDropsForCompatibleTool.
func (d *DoublePitcherCrop) GetDropsForCompatibleTool(item Item) []Item {
	if d.Age >= DoublePitcherCropMaxAge {
		return itemDrops(asItemOrNil(VanillaBlock("pitcher_plant")))
	}
	return itemDrops(vanillaItem("pitcher_pod"))
}
