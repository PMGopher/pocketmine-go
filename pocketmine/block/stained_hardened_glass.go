package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// StainedHardenedGlass is a port of pocketmine\block\StainedHardenedGlass.
type StainedHardenedGlass struct {
	HardenedGlass
	ColorComponent
}

func NewStainedHardenedGlass(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *StainedHardenedGlass {
	s := &StainedHardenedGlass{
		HardenedGlass:  HardenedGlass{Transparent{NewBlock(idInfo, name, typeInfo)}},
		ColorComponent: NewColorComponent(),
	}
	s.Init(s)
	return s
}

func (s *StainedHardenedGlass) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *StainedHardenedGlass) DescribeBlockItemState(w runtime.DataDescriber) { s.DescribeColor(w) }
