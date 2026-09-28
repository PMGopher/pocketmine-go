package item

// PaintingItem is a port of pocketmine\item\PaintingItem.
type PaintingItem struct {
	ItemBase
}

func NewPaintingItem(identifier ItemIdentifier, name string) *PaintingItem {
	p := &PaintingItem{}
	p.Init(p, identifier, name)
	return p
}

func (p *PaintingItem) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}
