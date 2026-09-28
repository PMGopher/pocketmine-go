package item

import "pocketmine-go/pocketmine/block"

// CocoaBeans is a port of pocketmine\item\CocoaBeans.
type CocoaBeans struct {
	ItemBase
}

func NewCocoaBeans(identifier ItemIdentifier, name string) *CocoaBeans {
	c := &CocoaBeans{}
	c.Init(c, identifier, name)
	return c
}

func (c *CocoaBeans) Clone() Item {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

// GetBlock is a port of CocoaBeans::getBlock.
func (x *CocoaBeans) GetBlock() block.Behavior { return block.VanillaBlock("cocoa_pod") }
