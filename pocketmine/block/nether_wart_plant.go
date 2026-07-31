package block

import (
	"math/rand"

	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const NetherWartPlantMaxAge = 3

// NetherWartPlant is a port of pocketmine\block\NetherWartPlant.
type NetherWartPlant struct {
	Flowable
	AgeComponent
}

func NewNetherWartPlant(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *NetherWartPlant {
	n := &NetherWartPlant{
		Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(NetherWartPlantMaxAge),
	}
	n.Init(n)
	return n
}

func (n *NetherWartPlant) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *NetherWartPlant) DescribeBlockOnlyState(w runtime.DataDescriber) { n.DescribeAge(w) }

func (n *NetherWartPlant) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetSide(math.Down, 1).GetTypeId() == SOUL_SAND
}

func (n *NetherWartPlant) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return n.canBeSupportedAt(blockReplace) && n.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (n *NetherWartPlant) OnNearbyBlockChange() {
	if !n.canBeSupportedAt(n.self) {
		if world, err := n.position.GetWorld(); err == nil {
			world.UseBreakOn(n.position.AsVector3())
		}
	} else {
		n.Flowable.OnNearbyBlockChange()
	}
}

func (n *NetherWartPlant) TicksRandomly() bool { return n.Age < n.MaxAge }

// OnRandomTick should use BlockEventHelper.Grow (not yet ported — see block/utils) to fire the
// grow event before applying the change; for now it grows unconditionally when the random roll
// succeeds, matching everything except that event hook.
func (n *NetherWartPlant) OnRandomTick() {
	if n.Age < n.MaxAge && rand.Intn(11) == 0 {
		n.Age++
		if world, err := n.position.GetWorld(); err == nil {
			if err := world.SetBlock(n.position, n.self); err != nil {
				panic(err)
			}
		}
	}
}

// GetDropsForCompatibleTool should scale the drop count via FortuneDropHelper (not yet ported)
// when fully grown; needs real Item construction from the unported item package regardless (see
// Block.GetDropsForCompatibleTool's doc comment), so this returns nil for now.
func (n *NetherWartPlant) GetDropsForCompatibleTool(item Item) []Item { return nil }
