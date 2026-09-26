package block

// Potato is a port of pocketmine\block\Potato.
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

// GetDropsForCompatibleTool is a port of Potato::getDropsForCompatibleTool.
func (p *Potato) GetDropsForCompatibleTool(item Item) []Item {
	count := 1
	if p.Age >= CropsMaxAge {
		//min/max would be 2-5 in Java
		count = FortuneBinomial(item, 1, 3, 4.0/7)
	}
	result := itemDrops(vanillaItemCount("potato", count))
	if p.Age >= CropsMaxAge && mtRand(0, 49) == 0 {
		if poisonous := vanillaItem("poisonous_potato"); poisonous != nil {
			result = append(result, poisonous)
		}
	}
	return result
}
