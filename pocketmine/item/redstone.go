package item

// Redstone is a port of pocketmine\item\Redstone. GetBlock (should return VanillaBlocks.REDSTONE_WIRE())
// isn't ported - see StringItem's doc comment for why.
type Redstone struct {
	ItemBase
}

func NewRedstone(identifier ItemIdentifier, name string) *Redstone {
	r := &Redstone{}
	r.Init(r, identifier, name)
	return r
}

func (r *Redstone) Clone() Item {
	c := *r
	c.rebind(&c)
	return &c
}
