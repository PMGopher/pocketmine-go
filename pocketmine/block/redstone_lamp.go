package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// RedstoneLamp is a port of pocketmine\block\RedstoneLamp.
type RedstoneLamp struct {
	Opaque
	PoweredByRedstoneComponent
}

func NewRedstoneLamp(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *RedstoneLamp {
	r := &RedstoneLamp{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	r.Init(r)
	return r
}

func (r *RedstoneLamp) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RedstoneLamp) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Bool(&r.Powered) }

func (r *RedstoneLamp) GetLightLevel() int {
	if r.Powered {
		return 15
	}
	return 0
}

func (r *RedstoneLamp) IsLit() bool { return r.Powered }

func (r *RedstoneLamp) SetLit(lit bool) { r.Powered = lit }
