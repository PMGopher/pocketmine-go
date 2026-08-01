package block

// Carrot is a port of pocketmine\block\Carrot.
//
// GetDropsForCompatibleTool/AsItem should return VanillaItems.CARROT() (scaled via
// FortuneDropHelper when mature) — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so both are left as Crops'/Block's defaults for
// now.
type Carrot struct {
	Crops
}

func NewCarrot(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Carrot {
	c := &Carrot{Crops{
		Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(CropsMaxAge),
	}}
	c.Init(c)
	return c
}

func (c *Carrot) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}
