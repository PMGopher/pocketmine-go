package block

// Beetroot is a port of pocketmine\block\Beetroot.
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

// GetDropsForCompatibleTool is a port of Beetroot::getDropsForCompatibleTool.
func (b *Beetroot) GetDropsForCompatibleTool(item Item) []Item {
	if b.Age >= CropsMaxAge {
		var drops []Item
		if beetroot := vanillaItem("beetroot"); beetroot != nil {
			drops = append(drops, beetroot)
		}
		if seeds := vanillaItemCount("beetroot_seeds", FortuneBinomial(item, 0, 3, 4.0/7)); seeds != nil {
			drops = append(drops, seeds)
		}
		return drops
	}
	return itemDrops(vanillaItem("beetroot_seeds"))
}
