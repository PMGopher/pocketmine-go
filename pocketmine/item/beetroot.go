package item

// Beetroot is a port of pocketmine\item\Beetroot.
type Beetroot struct {
	Food
}

func NewBeetroot(identifier ItemIdentifier, name string) *Beetroot {
	b := &Beetroot{}
	b.Init(b, identifier, name)
	return b
}

func (b *Beetroot) Clone() Item {
	cl := *b
	cl.rebind(&cl)
	return &cl
}

func (b *Beetroot) GetFoodRestore() int { return 1 }

func (b *Beetroot) GetSaturationRestore() float64 { return 1.2 }
