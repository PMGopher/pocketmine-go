package item

// Minecart is a port of pocketmine\item\Minecart. Placing the minecart entity is //TODO in PHP too, so
// only the stack size is ported.
type Minecart struct {
	ItemBase
}

func NewMinecart(identifier ItemIdentifier, name string) *Minecart {
	m := &Minecart{}
	m.Init(m, identifier, name)
	return m
}

func (m *Minecart) Clone() Item {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *Minecart) GetMaxStackSize() int { return 1 }
