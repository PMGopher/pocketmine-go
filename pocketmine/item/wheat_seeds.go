package item

import "pocketmine-go/pocketmine/block"

// WheatSeeds is a port of pocketmine\item\WheatSeeds.
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

// GetBlock is a port of WheatSeeds::getBlock.
func (x *WheatSeeds) GetBlock() block.Behavior { return block.VanillaBlock("wheat") }
