package item

// Steak is a port of pocketmine\item\Steak.
type Steak struct {
	Food
}

func NewSteak(identifier ItemIdentifier, name string) *Steak {
	s := &Steak{}
	s.Init(s, identifier, name)
	return s
}

func (s *Steak) Clone() Item {
	cl := *s
	cl.rebind(&cl)
	return &cl
}

func (s *Steak) GetFoodRestore() int { return 8 }

func (s *Steak) GetSaturationRestore() float64 { return 12.8 }
