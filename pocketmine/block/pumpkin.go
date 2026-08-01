package block

// Pumpkin is a port of pocketmine\block\Pumpkin.
//
// OnInteract's shears-carving (turning into a facing CarvedPumpkin and dropping pumpkin seeds)
// needs a Shears item marker, the block registry (VanillaBlocks), and World.DropItem, none ported
// yet. Block's default OnInteract (return false) already matches this gap, so there's nothing to
// override here.
type Pumpkin struct {
	Opaque
}

func NewPumpkin(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Pumpkin {
	p := &Pumpkin{Opaque{NewBlock(idInfo, name, typeInfo)}}
	p.Init(p)
	return p
}

func (p *Pumpkin) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}
