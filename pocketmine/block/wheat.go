package block

// Wheat is a port of pocketmine\block\Wheat.
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

// GetDropsForCompatibleTool is a port of Wheat::getDropsForCompatibleTool.
func (w *Wheat) GetDropsForCompatibleTool(item Item) []Item {
	if w.Age >= CropsMaxAge {
		var drops []Item
		if wheat := vanillaItem("wheat"); wheat != nil {
			drops = append(drops, wheat)
		}
		if seeds := vanillaItemCount("wheat_seeds", FortuneBinomial(item, 0, 3, 4.0/7)); seeds != nil {
			drops = append(drops, seeds)
		}
		return drops
	}
	return itemDrops(vanillaItem("wheat_seeds"))
}
