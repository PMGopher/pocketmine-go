package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// StainedGlass is a port of pocketmine\block\StainedGlass.
type StainedGlass struct {
	Glass
	ColorComponent
}

func NewStainedGlass(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *StainedGlass {
	s := &StainedGlass{
		Glass:          Glass{Transparent{NewBlock(idInfo, name, typeInfo)}},
		ColorComponent: NewColorComponent(),
	}
	s.Init(s)
	return s
}

func (s *StainedGlass) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *StainedGlass) DescribeBlockItemState(w runtime.DataDescriber) { s.DescribeColor(w) }
