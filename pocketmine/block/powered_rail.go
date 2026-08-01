package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// PoweredRail is a port of pocketmine\block\PoweredRail.
type PoweredRail struct {
	StraightOnlyRail
	PoweredByRedstoneComponent
}

func NewPoweredRail(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *PoweredRail {
	p := &PoweredRail{StraightOnlyRail: StraightOnlyRail{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}}
	p.Init(p)
	return p
}

func (p *PoweredRail) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *PoweredRail) DescribeBlockOnlyState(w runtime.DataDescriber) {
	p.StraightOnlyRail.DescribeBlockOnlyState(w)
	w.Bool(&p.Powered)
}
