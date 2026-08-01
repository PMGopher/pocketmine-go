package block

import blockutils "pocketmine-go/pocketmine/block/utils"

// WoodenTrapdoor is a port of pocketmine\block\WoodenTrapdoor.
type WoodenTrapdoor struct {
	Trapdoor
	WoodTypeComponent
}

func NewWoodenTrapdoor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *WoodenTrapdoor {
	w := &WoodenTrapdoor{
		Trapdoor: Trapdoor{
			Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
			HorizontalFacingComponent: NewHorizontalFacingComponent(),
		},
		WoodTypeComponent: NewWoodTypeComponent(woodType),
	}
	w.Init(w)
	return w
}

func (w *WoodenTrapdoor) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WoodenTrapdoor) GetFuelTime() int { return 300 }
