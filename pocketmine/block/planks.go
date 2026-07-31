package block

import blockutils "pocketmine-go/pocketmine/block/utils"

// Planks is a port of pocketmine\block\Planks.
type Planks struct {
	Opaque
	WoodTypeComponent
}

func NewPlanks(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *Planks {
	p := &Planks{
		Opaque:            Opaque{NewBlock(idInfo, name, typeInfo)},
		WoodTypeComponent: NewWoodTypeComponent(woodType),
	}
	p.Init(p)
	return p
}

func (p *Planks) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *Planks) GetFuelTime() int {
	if p.WoodType.IsFlammable() {
		return 300
	}
	return 0
}

func (p *Planks) GetFlameEncouragement() int {
	if p.WoodType.IsFlammable() {
		return 5
	}
	return 0
}

func (p *Planks) GetFlammability() int {
	if p.WoodType.IsFlammable() {
		return 20
	}
	return 0
}
