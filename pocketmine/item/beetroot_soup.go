package item

// BeetrootSoup is a port of pocketmine\item\BeetrootSoup (its use is in item_use.go).
type BeetrootSoup struct {
	Food
}

func NewBeetrootSoup(identifier ItemIdentifier, name string) *BeetrootSoup {
	b := &BeetrootSoup{}
	b.Init(b, identifier, name)
	return b
}

func (b *BeetrootSoup) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *BeetrootSoup) GetMaxStackSize() int { return 1 }

func (b *BeetrootSoup) GetFoodRestore() int { return 6 }

func (b *BeetrootSoup) GetSaturationRestore() float64 { return 7.2 }
