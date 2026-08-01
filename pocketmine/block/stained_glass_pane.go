package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// StainedGlassPane is a port of pocketmine\block\StainedGlassPane.
type StainedGlassPane struct {
	GlassPane
	ColorComponent
}

func NewStainedGlassPane(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *StainedGlassPane {
	s := &StainedGlassPane{
		GlassPane:      GlassPane{Thin{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Connections: map[math.Facing]bool{}}},
		ColorComponent: NewColorComponent(),
	}
	s.Init(s)
	return s
}

// Clone can't use the usual one-line pattern: Connections is a map, a reference type - see
// Vine.Clone's doc comment for the same reasoning.
func (s *StainedGlassPane) Clone() Behavior {
	c := *s
	c.Connections = make(map[math.Facing]bool, len(s.Connections))
	for k, v := range s.Connections {
		c.Connections[k] = v
	}
	c.rebind(&c)
	return &c
}

func (s *StainedGlassPane) DescribeBlockItemState(w runtime.DataDescriber) { s.DescribeColor(w) }
