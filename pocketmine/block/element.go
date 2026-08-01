package block

// Element is a port of pocketmine\block\Element.
type Element struct {
	Opaque

	Symbol       string
	AtomicWeight int
	Group        int
}

func NewElement(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, symbol string, atomicWeight int, group int) *Element {
	e := &Element{
		Opaque:       Opaque{NewBlock(idInfo, name, typeInfo)},
		Symbol:       symbol,
		AtomicWeight: atomicWeight,
		Group:        group,
	}
	e.Init(e)
	return e
}

func (e *Element) Clone() Behavior {
	c := *e
	c.rebind(&c)
	return &c
}

func (e *Element) GetAtomicWeight() int { return e.AtomicWeight }

func (e *Element) GetGroup() int { return e.Group }

func (e *Element) GetSymbol() string { return e.Symbol }
