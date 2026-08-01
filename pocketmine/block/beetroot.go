package block

// Beetroot is a port of pocketmine\block\Beetroot.
//
// GetDropsForCompatibleTool/AsItem should return VanillaItems.BEETROOT_SEEDS()/BEETROOT() — needs
// the unported item package (see Block.GetDropsForCompatibleTool's doc comment), so both are left
// as Crops'/Block's defaults for now.
type Beetroot struct {
	Crops
}

func NewBeetroot(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Beetroot {
	b := &Beetroot{Crops{
		Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(CropsMaxAge),
	}}
	b.Init(b)
	return b
}

func (b *Beetroot) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}
