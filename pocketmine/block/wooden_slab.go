package block

import blockutils "pocketmine-go/pocketmine/block/utils"

// WoodenSlab is a port of pocketmine\block\WoodenSlab.
type WoodenSlab struct {
	Slab
	WoodTypeComponent
}

func NewWoodenSlab(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *WoodenSlab {
	w := &WoodenSlab{
		Slab:              Slab{Transparent: Transparent{NewBlock(idInfo, name+" Slab", typeInfo)}, SlabTypeValue: blockutils.SlabTypeBottom},
		WoodTypeComponent: NewWoodTypeComponent(woodType),
	}
	w.Init(w)
	return w
}

func (w *WoodenSlab) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WoodenSlab) GetFuelTime() int {
	if w.WoodType.IsFlammable() {
		return 300
	}
	return 0
}

func (w *WoodenSlab) GetFlameEncouragement() int {
	if w.WoodType.IsFlammable() {
		return 5
	}
	return 0
}

func (w *WoodenSlab) GetFlammability() int {
	if w.WoodType.IsFlammable() {
		return 20
	}
	return 0
}
