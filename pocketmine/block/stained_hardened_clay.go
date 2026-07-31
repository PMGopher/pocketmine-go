package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// StainedHardenedClay is a port of pocketmine\block\StainedHardenedClay.
type StainedHardenedClay struct {
	HardenedClay
	ColorComponent
}

func NewStainedHardenedClay(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *StainedHardenedClay {
	s := &StainedHardenedClay{
		HardenedClay:   HardenedClay{Opaque{NewBlock(idInfo, name, typeInfo)}},
		ColorComponent: NewColorComponent(),
	}
	s.Init(s)
	return s
}

func (s *StainedHardenedClay) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *StainedHardenedClay) DescribeBlockItemState(w runtime.DataDescriber) { s.DescribeColor(w) }
