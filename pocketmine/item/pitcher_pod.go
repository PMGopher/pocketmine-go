package item

import "pocketmine-go/pocketmine/block"

// PitcherPod is a port of pocketmine\item\PitcherPod.
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

// GetBlock is a port of PitcherPod::getBlock.
func (x *PitcherPod) GetBlock() block.Behavior { return block.VanillaBlock("pitcher_crop") }
