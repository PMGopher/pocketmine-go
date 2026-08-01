package item

// PumpkinPie is a port of pocketmine\item\PumpkinPie.
type PumpkinPie struct {
	Food
}

func NewPumpkinPie(identifier ItemIdentifier, name string) *PumpkinPie {
	p := &PumpkinPie{}
	p.Init(p, identifier, name)
	return p
}

func (p *PumpkinPie) Clone() Item {
	cl := *p
	cl.rebind(&cl)
	return &cl
}

func (p *PumpkinPie) GetFoodRestore() int { return 8 }

func (p *PumpkinPie) GetSaturationRestore() float64 { return 4.8 }
