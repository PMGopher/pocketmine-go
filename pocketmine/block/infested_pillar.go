package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// InfestedPillar is a port of pocketmine\block\InfestedPillar.
type InfestedPillar struct {
	InfestedStone
	PillarRotationComponent
}

func NewInfestedPillar(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, imitated Behavior) *InfestedPillar {
	p := &InfestedPillar{
		InfestedStone: InfestedStone{
			Opaque:          Opaque{NewBlock(idInfo, name, typeInfo)},
			ImitatedStateID: imitated.GetStateId(),
		},
		PillarRotationComponent: NewPillarRotationComponent(),
	}
	p.Init(p)
	return p
}

func (p *InfestedPillar) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *InfestedPillar) DescribeBlockOnlyState(w runtime.DataDescriber) { p.DescribeAxis(w) }

func (p *InfestedPillar) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	p.SetAxisFromFace(face)
	return p.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
