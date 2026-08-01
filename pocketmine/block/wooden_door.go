package block

import blockutils "pocketmine-go/pocketmine/block/utils"

// WoodenDoor is a port of pocketmine\block\WoodenDoor.
type WoodenDoor struct {
	Door
	WoodTypeComponent
}

func NewWoodenDoor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *WoodenDoor {
	w := &WoodenDoor{
		Door: Door{
			Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
			HorizontalFacingComponent: NewHorizontalFacingComponent(),
		},
		WoodTypeComponent: NewWoodTypeComponent(woodType),
	}
	w.Init(w)
	return w
}

func (w *WoodenDoor) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WoodenDoor) GetFuelTime() int {
	if w.WoodType.IsFlammable() {
		return 200
	}
	return 0
}
