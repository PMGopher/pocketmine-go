package item

// BakedPotato is a port of pocketmine\item\BakedPotato.
type BakedPotato struct {
	Food
}

func NewBakedPotato(identifier ItemIdentifier, name string) *BakedPotato {
	i := &BakedPotato{}
	i.Init(i, identifier, name)
	return i
}

func (i *BakedPotato) Clone() Item {
	c := *i
	c.rebind(&c)
	return &c
}

func (i *BakedPotato) GetFoodRestore() int { return 5 }

func (i *BakedPotato) GetSaturationRestore() float64 { return 7.2 }
