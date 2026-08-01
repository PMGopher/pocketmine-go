package item

// Bowl is a port of pocketmine\item\Bowl.
type Bowl struct {
	ItemBase
}

func NewBowl(identifier ItemIdentifier, name string) *Bowl {
	b := &Bowl{}
	b.Init(b, identifier, name)
	return b
}

func (b *Bowl) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Bowl) GetFuelTime() int { return 200 }
