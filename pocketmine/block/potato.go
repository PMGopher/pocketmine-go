package block

// Potato is a port of pocketmine\block\Potato.
//
// GetDropsForCompatibleTool/AsItem should return VanillaItems.POTATO() (plus a chance of a
// poisonous potato when mature) — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so both are left as Crops'/Block's defaults for
// now.
type Potato struct {
	Crops
}

func NewPotato(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Potato {
	p := &Potato{Crops{
		Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(CropsMaxAge),
	}}
	p.Init(p)
	return p
}

func (p *Potato) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}
