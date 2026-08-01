package item

// PoisonousPotato is a port of pocketmine\item\PoisonousPotato. GetAdditionalEffects (a ~60%
// chance of a Poison effect) isn't ported - see GoldenApple's doc comment for why.
type PoisonousPotato struct {
	Food
}

func NewPoisonousPotato(identifier ItemIdentifier, name string) *PoisonousPotato {
	p := &PoisonousPotato{}
	p.Init(p, identifier, name)
	return p
}

func (p *PoisonousPotato) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *PoisonousPotato) GetFoodRestore() int { return 2 }

func (p *PoisonousPotato) GetSaturationRestore() float64 { return 1.2 }
