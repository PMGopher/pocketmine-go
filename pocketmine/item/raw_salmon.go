package item

// RawSalmon is a port of pocketmine\item\RawSalmon.
type RawSalmon struct {
	Food
}

func NewRawSalmon(identifier ItemIdentifier, name string) *RawSalmon {
	r := &RawSalmon{}
	r.Init(r, identifier, name)
	return r
}

func (r *RawSalmon) Clone() Item {
	cl := *r
	cl.rebind(&cl)
	return &cl
}

func (r *RawSalmon) GetFoodRestore() int { return 2 }

func (r *RawSalmon) GetSaturationRestore() float64 { return 0.2 }
