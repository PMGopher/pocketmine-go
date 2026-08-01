package block

import blockutils "pocketmine-go/pocketmine/block/utils"

// WoodenStairs is a port of pocketmine\block\WoodenStairs.
type WoodenStairs struct {
	Stair
	WoodTypeComponent
}

func NewWoodenStairs(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *WoodenStairs {
	w := &WoodenStairs{
		Stair: Stair{
			Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
			HorizontalFacingComponent: NewHorizontalFacingComponent(),
			Shape:                     blockutils.StairShapeStraight,
		},
		WoodTypeComponent: NewWoodTypeComponent(woodType),
	}
	w.Init(w)
	return w
}

func (w *WoodenStairs) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WoodenStairs) GetFuelTime() int {
	if w.WoodType.IsFlammable() {
		return 300
	}
	return 0
}

// GetFlameEncouragement/GetFlammability are unconditional in the PHP original (unlike other
// WoodType-dependent blocks) - not gated on w.WoodType.IsFlammable().
func (w *WoodenStairs) GetFlameEncouragement() int { return 5 }

func (w *WoodenStairs) GetFlammability() int { return 20 }
