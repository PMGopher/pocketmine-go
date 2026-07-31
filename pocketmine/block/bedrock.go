package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// Bedrock is a port of pocketmine\block\Bedrock.
type Bedrock struct {
	Opaque

	BurnsForeverFlag bool
}

func NewBedrock(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Bedrock {
	b := &Bedrock{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	b.Init(b)
	return b
}

func (b *Bedrock) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Bedrock) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Bool(&b.BurnsForeverFlag) }

func (b *Bedrock) BurnsForever() bool { return b.BurnsForeverFlag }

func (b *Bedrock) SetBurnsForever(burnsForever bool) { b.BurnsForeverFlag = burnsForever }
