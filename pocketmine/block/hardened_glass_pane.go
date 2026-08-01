package block

import "pocketmine-go/pocketmine/math"

// HardenedGlassPane is a port of pocketmine\block\HardenedGlassPane.
type HardenedGlassPane struct {
	Thin
}

func NewHardenedGlassPane(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *HardenedGlassPane {
	h := &HardenedGlassPane{Thin{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Connections: map[math.Facing]bool{}}}
	h.Init(h)
	return h
}

// Clone can't use the usual one-line pattern: Connections is a map, a reference type - see
// Vine.Clone's doc comment for the same reasoning.
func (h *HardenedGlassPane) Clone() Behavior {
	c := *h
	c.Connections = make(map[math.Facing]bool, len(h.Connections))
	for k, v := range h.Connections {
		c.Connections[k] = v
	}
	c.rebind(&c)
	return &c
}
