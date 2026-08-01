package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// RedstoneTorch is a port of pocketmine\block\RedstoneTorch.
type RedstoneTorch struct {
	Torch
	LightableComponent
}

func NewRedstoneTorch(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *RedstoneTorch {
	r := &RedstoneTorch{
		Torch:              Torch{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, Facing: math.Up},
		LightableComponent: LightableComponent{Lit: true},
	}
	r.Init(r)
	return r
}

func (r *RedstoneTorch) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RedstoneTorch) DescribeBlockOnlyState(w runtime.DataDescriber) {
	r.Torch.DescribeBlockOnlyState(w)
	w.Bool(&r.Lit)
}

func (r *RedstoneTorch) GetLightLevel() int {
	if r.Lit {
		return 7
	}
	return 0
}
