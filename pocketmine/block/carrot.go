package block

// Carrot is a port of pocketmine\block\Carrot.
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

// GetDropsForCompatibleTool is a port of Carrot::getDropsForCompatibleTool.
func (c *Carrot) GetDropsForCompatibleTool(item Item) []Item {
	count := 1
	if c.Age >= CropsMaxAge {
		count = FortuneBinomial(item, 1, 3, 4.0/7)
	}
	return itemDrops(vanillaItemCount("carrot", count))
}
