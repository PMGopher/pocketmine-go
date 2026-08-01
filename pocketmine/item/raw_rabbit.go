package item

// RawRabbit is a port of pocketmine\item\RawRabbit.
type RawRabbit struct {
	Food
}

func NewRawRabbit(identifier ItemIdentifier, name string) *RawRabbit {
	r := &RawRabbit{}
	r.Init(r, identifier, name)
	return r
}

func (r *RawRabbit) Clone() Item {
	cl := *r
	cl.rebind(&cl)
	return &cl
}

func (r *RawRabbit) GetFoodRestore() int { return 3 }

func (r *RawRabbit) GetSaturationRestore() float64 { return 1.8 }
