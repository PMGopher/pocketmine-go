package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// Wool is a port of pocketmine\block\Wool.
type Wool struct {
	Opaque
	ColorComponent
}

func NewWool(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Wool {
	w := &Wool{
		Opaque:         Opaque{NewBlock(idInfo, name, typeInfo)},
		ColorComponent: NewColorComponent(),
	}
	w.Init(w)
	return w
}

func (w *Wool) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *Wool) DescribeBlockItemState(d runtime.DataDescriber) { w.DescribeColor(d) }

func (w *Wool) GetFlameEncouragement() int { return 30 }

func (w *Wool) GetFlammability() int { return 60 }
