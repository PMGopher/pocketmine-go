package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// ActivatorRail is a port of pocketmine\block\ActivatorRail.
type ActivatorRail struct {
	StraightOnlyRail
	PoweredByRedstoneComponent
}

func NewActivatorRail(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ActivatorRail {
	a := &ActivatorRail{StraightOnlyRail: StraightOnlyRail{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}}
	a.Init(a)
	return a
}

func (a *ActivatorRail) Clone() Behavior {
	c := *a
	c.rebind(&c)
	return &c
}

func (a *ActivatorRail) DescribeBlockOnlyState(w runtime.DataDescriber) {
	a.StraightOnlyRail.DescribeBlockOnlyState(w)
	w.Bool(&a.Powered)
}
