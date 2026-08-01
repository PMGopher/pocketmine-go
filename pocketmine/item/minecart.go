package item

// Minecart is a port of pocketmine\item\Minecart. Placing/spawning the minecart entity (marked
// //TODO even in the PHP original) needs the unported entity package, so only MaxStackSize is
// ported here.
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
