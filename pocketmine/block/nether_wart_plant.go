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

// OnRandomTick is a port of NetherWartPlant::onRandomTick.
func (n *NetherWartPlant) OnRandomTick() {
	if n.Age < n.MaxAge && rand.Intn(11) == 0 { // Still growing
		grown := n.self.Clone().(*NetherWartPlant)
		grown.Age++
		Grow(n.self, grown, nil)
	}
}

// GetDropsForCompatibleTool is a port of NetherWartPlant::getDropsForCompatibleTool.
func (n *NetherWartPlant) GetDropsForCompatibleTool(item Item) []Item {
	drop := asItemOrNil(n.self)
	if drop == nil {
		return nil
	}
	count := 1
	if n.Age == NetherWartPlantMaxAge {
		count = FortuneDiscrete(item, 2, 4)
	}
	drop.SetCount(count)
	return []Item{drop}
}
