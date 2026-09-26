package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
	"pocketmine-go/pocketmine/math"
)

const leavesMaxLogDistance = 4

// Leaves is a port of pocketmine\block\Leaves.
type Leaves struct {
	Transparent

	LeavesTypeValue blockutils.LeavesType // immutable for now
	NoDecay         bool
	CheckDecay      bool
}

func NewLeaves(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, leavesType blockutils.LeavesType) *Leaves {
	l := &Leaves{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, LeavesTypeValue: leavesType}
	l.Init(l)
	return l
}

func (l *Leaves) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *Leaves) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.Bool(&l.NoDecay)
	w.Bool(&l.CheckDecay)
}

func (l *Leaves) GetLeavesType() blockutils.LeavesType { return l.LeavesTypeValue }

func (l *Leaves) IsNoDecay() bool { return l.NoDecay }

func (l *Leaves) SetNoDecay(noDecay bool) { l.NoDecay = noDecay }

func (l *Leaves) IsCheckDecay() bool { return l.CheckDecay }

func (l *Leaves) SetCheckDecay(checkDecay bool) { l.CheckDecay = checkDecay }

func (l *Leaves) BlocksDirectSkyLight() bool { return true }

func (l *Leaves) findLog(pos math.Vector3, visited map[[3]int]bool, distance int) bool {
	key := [3]int{pos.FloorX(), pos.FloorY(), pos.FloorZ()}
	if visited[key] {
		return false
	}
	visited[key] = true

	world, err := l.position.GetWorld()
	if err != nil {
		return false
	}
	blk := world.GetBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ())

	if _, ok := blk.(*Wood); ok { // type doesn't matter
		return true
	}

	if _, ok := blk.(*Leaves); ok && distance <= leavesMaxLogDistance {
		for _, side := range math.AllFacing {
			if l.findLog(pos.GetSide(side, 1), visited, distance+1) {
				return true
			}
		}
	}

	return false
}

// OnNearbyBlockChange doesn't model World::setBlock's $update=false parameter (block-update
// suppression isn't represented in the ported World interface yet) - functionally this still
// flags the leaf block for a decay check, just without that optimization.
func (l *Leaves) OnNearbyBlockChange() {
	if !l.NoDecay && !l.CheckDecay {
		l.CheckDecay = true
		if world, err := l.position.GetWorld(); err == nil {
			if err := setBlockWithoutUpdate(world, l.position, l.self); err != nil {
				panic(err)
			}
		}
	}
}

func (l *Leaves) TicksRandomly() bool { return !l.NoDecay && l.CheckDecay }

// OnRandomTick is a port of Leaves::onRandomTick.
func (l *Leaves) OnRandomTick() {
	if l.NoDecay || !l.CheckDecay {
		return
	}
	cancelled := false
	if event.HasHandlers[blockevent.LeavesDecayEvent]() {
		ev := blockevent.NewLeavesDecayEvent(l.self)
		event.Call(ev)
		cancelled = ev.IsCancelled()
	}
	world, err := l.position.GetWorld()
	if err != nil {
		return
	}
	if cancelled || l.findLog(l.position.AsVector3(), map[[3]int]bool{}, 0) {
		l.CheckDecay = false
		if err := setBlockWithoutUpdate(world, l.position, l.self); err != nil {
			panic(err)
		}
	} else {
		world.UseBreakOn(l.position.AsVector3())
	}
}

func (l *Leaves) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	l.NoDecay = true // artificial leaves don't decay
	return l.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// GetDropsForCompatibleTool is a port of Leaves::getDropsForCompatibleTool.
func (l *Leaves) GetDropsForCompatibleTool(item Item) []Item {
	if item.GetBlockToolType()&ToolTypeShears != 0 {
		return l.Block.GetDropsForCompatibleTool(item)
	}
	var drops []Item
	if FortuneBonusChanceDivisor(item, 20, 4) { // Saplings
		// TODO: according to the wiki, the jungle saplings have a different drop rate
		var sapling string
		switch l.LeavesTypeValue {
		case blockutils.LeavesTypeAcacia:
			sapling = "acacia_sapling"
		case blockutils.LeavesTypeBirch:
			sapling = "birch_sapling"
		case blockutils.LeavesTypeDarkOak:
			sapling = "dark_oak_sapling"
		case blockutils.LeavesTypeJungle:
			sapling = "jungle_sapling"
		case blockutils.LeavesTypeOak:
			sapling = "oak_sapling"
		case blockutils.LeavesTypeSpruce:
			sapling = "spruce_sapling"
		case blockutils.LeavesTypeMangrove, //TODO: mangrove propagule
			blockutils.LeavesTypeAzalea:
			sapling = "azalea"
		case blockutils.LeavesTypeFloweringAzalea:
			sapling = "flowering_azalea"
		}
		//TODO: cherry, pale oak
		if sapling != "" {
			if it := asItemOrNil(VanillaBlock(sapling)); it != nil {
				drops = append(drops, it)
			}
		}
	}
	if (l.LeavesTypeValue == blockutils.LeavesTypeOak || l.LeavesTypeValue == blockutils.LeavesTypeDarkOak) &&
		FortuneBonusChanceDivisor(item, 200, 20) { // Apples
		if apple := vanillaItem("apple"); apple != nil {
			drops = append(drops, apple)
		}
	}
	if FortuneBonusChanceDivisor(item, 50, 5) {
		if sticks := vanillaItemCount("stick", mtRand(1, 2)); sticks != nil {
			drops = append(drops, sticks)
		}
	}
	return drops
}

func (l *Leaves) IsAffectedBySilkTouch() bool { return true }

func (l *Leaves) GetFlameEncouragement() int { return 30 }

func (l *Leaves) GetFlammability() int { return 60 }

func (l *Leaves) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
