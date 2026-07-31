package block

// Reserved6 is a port of pocketmine\block\Reserved6.
type Reserved6 struct {
	Opaque
}

func NewReserved6(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Reserved6 {
	r := &Reserved6{Opaque{NewBlock(idInfo, name, typeInfo)}}
	r.Init(r)
	return r
}

func (r *Reserved6) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}
