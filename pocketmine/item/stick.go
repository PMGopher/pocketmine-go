package item

// Stick is a port of pocketmine\item\Stick.
type Stick struct {
	ItemBase
}

func NewStick(identifier ItemIdentifier, name string) *Stick {
	s := &Stick{}
	s.Init(s, identifier, name)
	return s
}

func (s *Stick) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Stick) GetFuelTime() int { return 100 }
