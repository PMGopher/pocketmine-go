package item

// PaintingItem is a port of pocketmine\item\PaintingItem. Not ported: onInteractBlock (placing a
// Painting entity against the clicked block) - it needs the Player item-use flow (see the Item
// interface's doc comment).
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
