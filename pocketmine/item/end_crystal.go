package item

// EndCrystal is a port of pocketmine\item\EndCrystal.
type EndCrystal struct {
	ItemBase
}

func NewEndCrystal(identifier ItemIdentifier, name string) *EndCrystal {
	e := &EndCrystal{}
	e.Init(e, identifier, name)
	return e
}

func (e *EndCrystal) Clone() Item {
	c := *e
	c.rebind(&c)
	return &c
}
