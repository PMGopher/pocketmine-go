package block

import blockutils "pocketmine-go/pocketmine/block/utils"

// WoodenPressurePlate is a port of pocketmine\block\WoodenPressurePlate.
type WoodenPressurePlate struct {
	SimplePressurePlate
	WoodTypeComponent
}

func NewWoodenPressurePlate(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType, deactivationDelayTicks int) *WoodenPressurePlate {
	w := &WoodenPressurePlate{
		SimplePressurePlate: SimplePressurePlate{
			PressurePlate: PressurePlate{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, DeactivationDelayTicks: deactivationDelayTicks},
		},
		WoodTypeComponent: NewWoodTypeComponent(woodType),
	}
	w.Init(w)
	return w
}

func (w *WoodenPressurePlate) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WoodenPressurePlate) GetFuelTime() int { return 300 }
