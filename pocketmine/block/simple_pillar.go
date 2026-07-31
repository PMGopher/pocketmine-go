package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// SimplePillar is a port of pocketmine\block\SimplePillar.
//
// The PHP doc comment warns this is @internal and shouldn't be used for API contract binding,
// since not all pillar-like blocks extend it — kept here purely as a base for the blocks that do.
type SimplePillar struct {
	Opaque
	PillarRotationComponent
}

func NewSimplePillar(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SimplePillar {
	s := &SimplePillar{
		Opaque:                  Opaque{NewBlock(idInfo, name, typeInfo)},
		PillarRotationComponent: NewPillarRotationComponent(),
	}
	s.Init(s)
	return s
}

func (s *SimplePillar) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SimplePillar) DescribeBlockOnlyState(w runtime.DataDescriber) { s.DescribeAxis(w) }

func (s *SimplePillar) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	s.SetAxisFromFace(face)
	return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
