package item

// PlainItem is a plain pocketmine\item\Item instance - items PHP registers as `new Item($id,
// $name)` with no subclass (ink sacs, iron ingots, ...).
type PlainItem struct {
	ItemBase
}

// NewItem is a port of Item::__construct for a plain item.
func NewItem(identifier ItemIdentifier, name string) *PlainItem {
	i := &PlainItem{}
	i.Init(i, identifier, name)
	return i
}

func (i *PlainItem) Clone() Item {
	c := *i
	c.rebind(&c)
	return &c
}
