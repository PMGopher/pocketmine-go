package item

// Apple is a port of pocketmine\item\Apple.
type Apple struct {
	Food
}

func NewApple(identifier ItemIdentifier, name string) *Apple {
	a := &Apple{}
	a.Init(a, identifier, name)
	return a
}

func (a *Apple) Clone() Item {
	c := *a
	c.rebind(&c)
	return &c
}

func (a *Apple) GetFoodRestore() int { return 4 }

func (a *Apple) GetSaturationRestore() float64 { return 2.4 }
