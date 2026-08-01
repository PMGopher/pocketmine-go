package item

// RawFish is a port of pocketmine\item\RawFish.
type RawFish struct {
	Food
}

func NewRawFish(identifier ItemIdentifier, name string) *RawFish {
	r := &RawFish{}
	r.Init(r, identifier, name)
	return r
}

func (r *RawFish) Clone() Item {
	cl := *r
	cl.rebind(&cl)
	return &cl
}

func (r *RawFish) GetFoodRestore() int { return 2 }

func (r *RawFish) GetSaturationRestore() float64 { return 0.4 }
