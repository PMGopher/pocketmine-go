package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// StainedHardenedGlassPane is a port of pocketmine\block\StainedHardenedGlassPane.
type StainedHardenedGlassPane struct {
	HardenedGlassPane
	ColorComponent
}

func NewStainedHardenedGlassPane(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *StainedHardenedGlassPane {
	s := &StainedHardenedGlassPane{
		HardenedGlassPane: HardenedGlassPane{Thin{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Connections: map[math.Facing]bool{}}},
		ColorComponent:    NewColorComponent(),
	}
	s.Init(s)
	return s
}

// Clone can't use the usual one-line pattern: Connections is a map, a reference type - see
// Vine.Clone's doc comment for the same reasoning.
func (s *StainedHardenedGlassPane) Clone() Behavior {
	c := *s
	c.Connections = make(map[math.Facing]bool, len(s.Connections))
	for k, v := range s.Connections {
		c.Connections[k] = v
	}
	c.rebind(&c)
	return &c
}

func (s *StainedHardenedGlassPane) DescribeBlockItemState(w runtime.DataDescriber) {
	s.DescribeColor(w)
}
