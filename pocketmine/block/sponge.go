package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// Sponge is a port of pocketmine\block\Sponge.
type Sponge struct {
	Opaque

	Wet bool
}

func NewSponge(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Sponge {
	s := &Sponge{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	s.Init(s)
	return s
}

func (s *Sponge) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Sponge) DescribeBlockItemState(w runtime.DataDescriber) { w.Bool(&s.Wet) }

func (s *Sponge) IsWet() bool { return s.Wet }

func (s *Sponge) SetWet(wet bool) { s.Wet = wet }
