package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// HayBale is a port of pocketmine\block\HayBale.
type HayBale struct {
	Opaque
	PillarRotationComponent
}

func NewHayBale(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *HayBale {
	h := &HayBale{
		Opaque:                  Opaque{NewBlock(idInfo, name, typeInfo)},
		PillarRotationComponent: NewPillarRotationComponent(),
	}
	h.Init(h)
	return h
}

func (h *HayBale) Clone() Behavior {
	c := *h
	c.rebind(&c)
	return &c
}

func (h *HayBale) DescribeBlockOnlyState(w runtime.DataDescriber) { h.DescribeAxis(w) }

func (h *HayBale) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	h.SetAxisFromFace(face)
	return h.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (h *HayBale) GetFlameEncouragement() int { return 60 }

func (h *HayBale) GetFlammability() int { return 20 }

func (h *HayBale) OnEntityLand(entity Entity) (float64, bool) {
	entity.SetFallDistance(entity.GetFallDistance() * 0.2)
	return 0, false
}
