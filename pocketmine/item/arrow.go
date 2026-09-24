package item

// Arrow is a port of pocketmine\item\Arrow (no behaviour of its own in PHP either - bows consume
// it, and Arrow entities are picked up as it).
type Arrow struct {
	ItemBase
}

func NewArrow(identifier ItemIdentifier, name string) *Arrow {
	a := &Arrow{}
	a.Init(a, identifier, name)
	return a
}

func (a *Arrow) Clone() Item {
	c := *a
	c.rebind(&c)
	return &c
}
