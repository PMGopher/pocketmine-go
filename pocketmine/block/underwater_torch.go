package block

import "pocketmine-go/pocketmine/math"

// UnderwaterTorch is a port of pocketmine\block\UnderwaterTorch.
type UnderwaterTorch struct {
	Torch
}

func NewUnderwaterTorch(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *UnderwaterTorch {
	u := &UnderwaterTorch{Torch{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, Facing: math.Up}}
	u.Init(u)
	return u
}

func (u *UnderwaterTorch) Clone() Behavior {
	c := *u
	c.rebind(&c)
	return &c
}

func (u *UnderwaterTorch) CanBeFlowedInto() bool { return false }
