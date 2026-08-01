package item

// BeetrootSeeds is a port of pocketmine\item\BeetrootSeeds. GetBlock (should return VanillaBlocks.BEETROOTS())
// isn't ported - see StringItem's doc comment for why.
type BeetrootSeeds struct {
	ItemBase
}

func NewBeetrootSeeds(identifier ItemIdentifier, name string) *BeetrootSeeds {
	b := &BeetrootSeeds{}
	b.Init(b, identifier, name)
	return b
}

func (b *BeetrootSeeds) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}
