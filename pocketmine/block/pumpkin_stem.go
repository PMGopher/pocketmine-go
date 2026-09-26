package block

import "pocketmine-go/pocketmine/math"

// PumpkinStem is a port of pocketmine\block\PumpkinStem.
type PumpkinStem struct {
	Stem
}

func NewPumpkinStem(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *PumpkinStem {
	p := &PumpkinStem{Stem{
		Crops: Crops{
			Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
			AgeComponent: NewAgeComponent(CropsMaxAge),
		},
		Facing: math.Up,
	}}
	p.Init(p)
	return p
}

func (p *PumpkinStem) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *PumpkinStem) GetPlantTypeID() int { return PUMPKIN }

func (p *PumpkinStem) GetPlant() Behavior { return VanillaPumpkin() }
