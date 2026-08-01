package item

// Bread is a port of pocketmine\item\Bread.
type Bread struct {
	Food
}

func NewBread(identifier ItemIdentifier, name string) *Bread {
	b := &Bread{}
	b.Init(b, identifier, name)
	return b
}

func (b *Bread) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Bread) GetFoodRestore() int { return 5 }

func (b *Bread) GetSaturationRestore() float64 { return 6 }
