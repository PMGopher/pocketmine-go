package block

// Wheat is a port of pocketmine\block\Wheat.
//
// GetDropsForCompatibleTool/AsItem should return VanillaItems.WHEAT_SEEDS()/WHEAT() — needs the
// unported item package (see Block.GetDropsForCompatibleTool's doc comment), so both are left as
// Crops'/Block's defaults for now.
type Wheat struct {
	Crops
}

func NewWheat(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Wheat {
	w := &Wheat{Crops{
		Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		AgeComponent: NewAgeComponent(CropsMaxAge),
	}}
	w.Init(w)
	return w
}

func (w *Wheat) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}
