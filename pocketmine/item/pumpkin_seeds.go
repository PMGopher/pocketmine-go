package item

// PumpkinSeeds is a port of pocketmine\item\PumpkinSeeds. GetBlock (should return VanillaBlocks.PUMPKIN_STEM())
// isn't ported - see StringItem's doc comment for why.
type PumpkinSeeds struct {
	ItemBase
}

func NewPumpkinSeeds(identifier ItemIdentifier, name string) *PumpkinSeeds {
	p := &PumpkinSeeds{}
	p.Init(p, identifier, name)
	return p
}

func (p *PumpkinSeeds) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}
