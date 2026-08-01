package item

// RawBeef is a port of pocketmine\item\RawBeef.
type RawBeef struct {
	Food
}

func NewRawBeef(identifier ItemIdentifier, name string) *RawBeef {
	r := &RawBeef{}
	r.Init(r, identifier, name)
	return r
}

func (r *RawBeef) Clone() Item {
	cl := *r
	cl.rebind(&cl)
	return &cl
}

func (r *RawBeef) GetFoodRestore() int { return 3 }

func (r *RawBeef) GetSaturationRestore() float64 { return 1.8 }
