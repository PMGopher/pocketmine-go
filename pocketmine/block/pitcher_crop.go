package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/event"
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

// grow is a port of PitcherCrop::grow.
func (p *PitcherCrop) grow(player Player) bool {
	if p.Age > pitcherCropMaxAge {
		return false
	}

	if p.Age == pitcherCropMaxAge {
		up := p.self.(blockGeometry).GetSide(math.Up, 1)
		if up.GetTypeId() != AIR {
			return false
		}

		world, err := p.position.GetWorld()
		if err != nil {
			return false
		}
		tx := NewBlockTransaction(world)
		bottom := VanillaDoublePitcherCrop().(*DoublePitcherCrop)
		bottom.SetTop(false)
		top := VanillaDoublePitcherCrop().(*DoublePitcherCrop)
		top.SetTop(true)
		tx.AddBlock(p.position, bottom)
		tx.AddBlock(p.position.GetSide(math.Up, 1), top)

		ev := &StructureGrowEvent{Block: p.self, Transaction: tx, Player: player}
		event.Call(ev)
		return !ev.IsCancelled() && tx.Apply()
	}

	clone := p.self.Clone().(*PitcherCrop)
	clone.SetAge(p.Age + 1)
	return Grow(p.self, clone, player)
}

// OnInteract is a port of PitcherCrop::onInteract. `$item instanceof Fertilizer` is checked via
// item type ID (bone meal is the only Fertilizer-marked item in the PHP original), same
// structural-marker convention as Crops.OnInteract.
func (p *PitcherCrop) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if item.GetTypeId() == itemTypeIDsBoneMeal && p.grow(player) {
		item.Pop()
		return true
	}
	return false
}

func (p *PitcherCrop) TicksRandomly() bool { return true }

// OnRandomTick is a port of PitcherCrop::onRandomTick.
func (p *PitcherCrop) OnRandomTick() {
	if CropGrowthCanGrow(p.self) {
		p.grow(nil)
	}
}

// AsItem should return VanillaItems.PITCHER_POD() — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.
