package item

// RawChicken is a port of pocketmine\item\RawChicken. GetAdditionalEffects (a 30% chance of a
// Hunger effect) isn't ported - see GoldenApple's doc comment for why.
type RawChicken struct {
	Food
}

func NewRawChicken(identifier ItemIdentifier, name string) *RawChicken {
	r := &RawChicken{}
	r.Init(r, identifier, name)
	return r
}

func (r *RawChicken) Clone() Item {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RawChicken) GetFoodRestore() int { return 2 }

func (r *RawChicken) GetSaturationRestore() float64 { return 1.2 }
