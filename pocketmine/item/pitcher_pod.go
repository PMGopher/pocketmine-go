package item

// PitcherPod is a port of pocketmine\item\PitcherPod. GetBlock (should return VanillaBlocks.PITCHER_CROP())
// isn't ported - see StringItem's doc comment for why.
type PitcherPod struct {
	ItemBase
}

func NewPitcherPod(identifier ItemIdentifier, name string) *PitcherPod {
	p := &PitcherPod{}
	p.Init(p, identifier, name)
	return p
}

func (p *PitcherPod) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}
