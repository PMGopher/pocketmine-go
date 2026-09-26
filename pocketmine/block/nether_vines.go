package block

import (
	"math/rand"
	blockevent "pocketmine-go/pocketmine/event/block"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
)

const NetherVinesMaxAge = 25

// NetherVines is a port of pocketmine\block\NetherVines. Used for both Weeping and Twisting
// vines, which share behaviour and differ only in growth direction and registration.
type NetherVines struct {
	Flowable
	AgeComponent

	// GrowthFace is the direction the vine grows towards.
	GrowthFace math.Facing
}

func NewNetherVines(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, growthFace math.Facing) *NetherVines {
	n := &NetherVines{
		Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(NetherVinesMaxAge),
		GrowthFace:   growthFace,
	}
	n.Init(n)
	return n
}

func (n *NetherVines) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *NetherVines) IsAffectedBySilkTouch() bool { return true }

func (n *NetherVines) CanClimb() bool { return true }

func (n *NetherVines) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Opposite(n.GrowthFace), 1)
	return support.GetSupportType(n.GrowthFace).HasCenterSupport() || support.(blockGeometry).HasSameTypeId(n.self)
}

func (n *NetherVines) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return n.canBeSupportedAt(blockReplace) && n.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (n *NetherVines) OnNearbyBlockChange() {
	if !n.canBeSupportedAt(n.self) {
		if world, err := n.position.GetWorld(); err == nil {
			world.UseBreakOn(n.position.AsVector3())
		}
	} else {
		n.Flowable.OnNearbyBlockChange()
	}
}

// seekToTip is a port of NetherVines::seekToTip.
func (n *NetherVines) seekToTip() *NetherVines {
	top := n
	for {
		next, ok := top.self.(blockGeometry).GetSide(top.GrowthFace, 1).(*NetherVines)
		if !ok || !next.HasSameTypeId(top.self) {
			break
		}
		top = next
	}
	return top
}

func (n *NetherVines) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	n.Age = rand.Intn(NetherVinesMaxAge)
	return n.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract is a port of NetherVines::onInteract.
func (n *NetherVines) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if isFertilizer(item) {
		if n.grow(player, mtRand(1, 5)) {
			item.Pop()
		}
		return true
	}
	return false
}

func (n *NetherVines) TicksRandomly() bool { return n.Age < NetherVinesMaxAge }

// grow is a port of NetherVines::grow.
func (n *NetherVines) grow(player Player, growthAmount int) bool {
	top := n.seekToTip()
	age := top.Age
	world, err := top.position.GetWorld()
	if err != nil {
		return false
	}
	changedBlocks := 0

	tx := NewBlockTransaction(world)
	for i := 1; i <= growthAmount; i++ {
		growthPos := top.position.GetSide(top.GrowthFace, i)
		if !world.IsInWorld(growthPos.FloorX(), growthPos.FloorY(), growthPos.FloorZ()) ||
			!world.GetBlockAt(growthPos.FloorX(), growthPos.FloorY(), growthPos.FloorZ()).CanBeReplaced() {
			break
		}
		age++
		if age > NetherVinesMaxAge {
			age = NetherVinesMaxAge
		}
		grown := top.self.Clone().(*NetherVines)
		grown.Age = age
		tx.AddBlock(growthPos, grown)
		changedBlocks++
	}

	if changedBlocks > 0 {
		ev := blockevent.NewStructureGrowEvent(top.self, tx, eventPlayer(player))
		event.Call(ev)
		if ev.IsCancelled() {
			return false
		}
		return tx.Apply()
	}

	return false
}

func (n *NetherVines) OnRandomTick() {
	if n.Age < NetherVinesMaxAge && rand.Intn(10) == 0 {
		if n.self.(blockGeometry).GetSide(n.GrowthFace, 1).CanBeReplaced() {
			n.grow(nil, 1)
		}
	}
}

func (n *NetherVines) HasEntityCollision() bool { return true }

func (n *NetherVines) OnEntityInside(entity Entity) bool {
	entity.ResetFallDistance()
	return false
}

func (n *NetherVines) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

// GetDropsForCompatibleTool is a port of NetherVines::getDropsForCompatibleTool.
func (n *NetherVines) GetDropsForCompatibleTool(item Item) []Item {
	if item.GetBlockToolType()&ToolTypeShears != 0 || FortuneBonusChanceFixed(item, 1.0/3, 2.0/9) {
		return itemDrops(asItemOrNil(n.self))
	}
	return nil
}

func (n *NetherVines) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
