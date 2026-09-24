package item

// EndCrystal is a port of pocketmine\item\EndCrystal. Not ported: onInteractBlock (placing an
// EndCrystal entity on obsidian/bedrock) - it needs the Player item-use flow (see the Item
// interface's doc comment).
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
