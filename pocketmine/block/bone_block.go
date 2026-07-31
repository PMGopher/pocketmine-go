package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// BoneBlock is a port of pocketmine\block\BoneBlock.
type BoneBlock struct {
	Opaque
	PillarRotationComponent
}

func NewBoneBlock(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *BoneBlock {
	b := &BoneBlock{
		Opaque:                  Opaque{NewBlock(idInfo, name, typeInfo)},
		PillarRotationComponent: NewPillarRotationComponent(),
	}
	b.Init(b)
	return b
}

func (b *BoneBlock) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *BoneBlock) DescribeBlockOnlyState(w runtime.DataDescriber) { b.DescribeAxis(w) }

func (b *BoneBlock) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	b.SetAxisFromFace(face)
	return b.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}
