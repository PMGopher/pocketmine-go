package item

// StringItem is a port of pocketmine\item\StringItem (named to avoid colliding with Go's string
// type). GetBlock (should return VanillaBlocks.TRIPWIRE()) isn't ported - GetBlock isn't part of
// the Item interface here at all yet (see the Item interface's doc comment), and needs the
// unported block registry regardless.
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
