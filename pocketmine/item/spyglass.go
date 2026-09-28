package item

// Spyglass is a port of pocketmine\item\Spyglass (its use is in item_use.go).
type Spyglass struct {
	ItemBase
}

func NewSpyglass(identifier ItemIdentifier, name string) *Spyglass {
	s := &Spyglass{}
	s.Init(s, identifier, name)
	return s
}

func (s *Spyglass) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Spyglass) GetMaxStackSize() int { return 1 }
