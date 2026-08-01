package item

// RawPorkchop is a port of pocketmine\item\RawPorkchop.
type RawPorkchop struct {
	Food
}

func NewRawPorkchop(identifier ItemIdentifier, name string) *RawPorkchop {
	r := &RawPorkchop{}
	r.Init(r, identifier, name)
	return r
}

func (r *RawPorkchop) Clone() Item {
	cl := *r
	cl.rebind(&cl)
	return &cl
}

func (r *RawPorkchop) GetFoodRestore() int { return 3 }

func (r *RawPorkchop) GetSaturationRestore() float64 { return 0.6 }
