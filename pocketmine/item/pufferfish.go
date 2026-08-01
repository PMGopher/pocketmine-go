package item

// Pufferfish is a port of pocketmine\item\Pufferfish. GetAdditionalEffects (Hunger + Poison +
// Nausea) isn't ported - see GoldenApple's doc comment for why.
type Pufferfish struct {
	Food
}

func NewPufferfish(identifier ItemIdentifier, name string) *Pufferfish {
	p := &Pufferfish{}
	p.Init(p, identifier, name)
	return p
}

func (p *Pufferfish) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *Pufferfish) GetFoodRestore() int { return 1 }

func (p *Pufferfish) GetSaturationRestore() float64 { return 0.2 }
