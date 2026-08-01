package item

// Potato is a port of pocketmine\item\Potato. GetBlock (should return VanillaBlocks.POTATOES())
// isn't ported - see StringItem's doc comment for why.
type Potato struct {
	Food
}

func NewPotato(identifier ItemIdentifier, name string) *Potato {
	p := &Potato{}
	p.Init(p, identifier, name)
	return p
}

func (p *Potato) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *Potato) GetFoodRestore() int { return 1 }

func (p *Potato) GetSaturationRestore() float64 { return 0.6 }
