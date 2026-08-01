package item

// CocoaBeans is a port of pocketmine\item\CocoaBeans. GetBlock (should return VanillaBlocks.COCOA_POD())
// isn't ported - see StringItem's doc comment for why.
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
