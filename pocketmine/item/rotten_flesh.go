package item

// RottenFlesh is a port of pocketmine\item\RottenFlesh. GetAdditionalEffects (an 80% chance of a
// Hunger effect) isn't ported - see GoldenApple's doc comment for why.
type RottenFlesh struct {
	Food
}

func NewRottenFlesh(identifier ItemIdentifier, name string) *RottenFlesh {
	r := &RottenFlesh{}
	r.Init(r, identifier, name)
	return r
}

func (r *RottenFlesh) Clone() Item {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RottenFlesh) GetFoodRestore() int { return 4 }

func (r *RottenFlesh) GetSaturationRestore() float64 { return 0.8 }
