package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
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
			if err := world.SetBlock(l.position, l.self); err != nil {
				panic(err)
			}
		}
	}
}

func (l *Leaves) TicksRandomly() bool { return !l.NoDecay && l.CheckDecay }

// OnRandomTick doesn't fire LeavesDecayEvent (deferred concrete event subclass - see the project
// todo list), so decay is never cancellable yet; the findLog nearby-log check still applies.
func (l *Leaves) OnRandomTick() {
	if l.NoDecay || !l.CheckDecay {
		return
	}
	world, err := l.position.GetWorld()
	if err != nil {
		return
	}
	if l.findLog(l.position.AsVector3(), map[[3]int]bool{}, 0) {
		l.CheckDecay = false
		if err := world.SetBlock(l.position, l.self); err != nil {
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

// GetDropsForCompatibleTool's shears branch is fully ported; the sapling/apple/stick fortune-based
// drops need FortuneDropHelper and the block/item registries (VanillaBlocks/VanillaItems), none
// ported yet, so that part returns nil for now (see Block.GetDropsForCompatibleTool's doc comment
// for the same category of gap).
func (l *Leaves) GetDropsForCompatibleTool(item Item) []Item {
	if item.GetBlockToolType()&ToolTypeShears != 0 {
		return l.Block.GetDropsForCompatibleTool(item)
	}
	return nil
}

func (l *Leaves) IsAffectedBySilkTouch() bool { return true }

func (l *Leaves) GetFlameEncouragement() int { return 30 }

func (l *Leaves) GetFlammability() int { return 60 }

func (l *Leaves) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
