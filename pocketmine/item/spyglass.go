package item

// Spyglass is a port of pocketmine\item\Spyglass. CanStartUsingItem (always true - needs a real
// Player) isn't ported - see the Item interface's doc comment on Player/Entity-interaction
// methods.
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
