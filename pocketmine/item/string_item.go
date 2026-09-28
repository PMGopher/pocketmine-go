package item

import "pocketmine-go/pocketmine/block"

// StringItem is a port of pocketmine\item\StringItem (named to avoid colliding with Go's string
// type).
type StringItem struct {
	ItemBase
}

func NewStringItem(identifier ItemIdentifier, name string) *StringItem {
	s := &StringItem{}
	s.Init(s, identifier, name)
	return s
}

func (s *StringItem) Clone() Item {
	c := *s
	c.rebind(&c)
	return &c
}

// GetBlock is a port of StringItem::getBlock.
func (x *StringItem) GetBlock() block.Behavior { return block.VanillaBlock("tripwire") }
