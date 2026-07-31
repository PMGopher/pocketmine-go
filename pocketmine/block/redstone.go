package block

// Redstone is a port of pocketmine\block\Redstone.
type Redstone struct {
	Opaque
}

func NewRedstone(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Redstone {
	r := &Redstone{Opaque{NewBlock(idInfo, name, typeInfo)}}
	r.Init(r)
	return r
}

func (r *Redstone) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}
