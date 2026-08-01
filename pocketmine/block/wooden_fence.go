package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// WoodenFence is a port of pocketmine\block\WoodenFence.
type WoodenFence struct {
	Fence
	WoodTypeComponent
}

func NewWoodenFence(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *WoodenFence {
	w := &WoodenFence{
		Fence:             Fence{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Connections: map[math.Facing]bool{}},
		WoodTypeComponent: NewWoodTypeComponent(woodType),
	}
	w.Init(w)
	return w
}

// Clone can't use the usual one-line pattern: Connections is a map, a reference type - see
// Vine.Clone's doc comment for the same reasoning.
func (w *WoodenFence) Clone() Behavior {
	c := *w
	c.Connections = make(map[math.Facing]bool, len(w.Connections))
	for k, v := range w.Connections {
		c.Connections[k] = v
	}
	c.rebind(&c)
	return &c
}

func (w *WoodenFence) GetFuelTime() int {
	if w.WoodType.IsFlammable() {
		return 300
	}
	return 0
}

func (w *WoodenFence) GetFlameEncouragement() int {
	if w.WoodType.IsFlammable() {
		return 5
	}
	return 0
}

func (w *WoodenFence) GetFlammability() int {
	if w.WoodType.IsFlammable() {
		return 20
	}
	return 0
}
