package item

// WheatSeeds is a port of pocketmine\item\WheatSeeds. GetBlock (should return VanillaBlocks.WHEAT())
// isn't ported - see StringItem's doc comment for why.
type WheatSeeds struct {
	ItemBase
}

func NewWheatSeeds(identifier ItemIdentifier, name string) *WheatSeeds {
	w := &WheatSeeds{}
	w.Init(w, identifier, name)
	return w
}

func (w *WheatSeeds) Clone() Item {
	c := *w
	c.rebind(&c)
	return &c
}
