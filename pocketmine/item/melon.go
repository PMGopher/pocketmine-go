package item

// Melon is a port of pocketmine\item\Melon.
type Melon struct {
	Food
}

func NewMelon(identifier ItemIdentifier, name string) *Melon {
	m := &Melon{}
	m.Init(m, identifier, name)
	return m
}

func (m *Melon) Clone() Item {
	cl := *m
	cl.rebind(&cl)
	return &cl
}

func (m *Melon) GetFoodRestore() int { return 2 }

func (m *Melon) GetSaturationRestore() float64 { return 1.2 }
