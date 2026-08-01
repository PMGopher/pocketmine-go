package item

// RawMutton is a port of pocketmine\item\RawMutton.
type RawMutton struct {
	Food
}

func NewRawMutton(identifier ItemIdentifier, name string) *RawMutton {
	r := &RawMutton{}
	r.Init(r, identifier, name)
	return r
}

func (r *RawMutton) Clone() Item {
	cl := *r
	cl.rebind(&cl)
	return &cl
}

func (r *RawMutton) GetFoodRestore() int { return 2 }

func (r *RawMutton) GetSaturationRestore() float64 { return 1.2 }
