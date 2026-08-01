package item

// WritableBook is a port of pocketmine\item\WritableBook - an empty class body in the PHP
// original too; all its behavior comes from the embedded WritableBookBase.
type WritableBook struct {
	WritableBookBase
}

func NewWritableBook(identifier ItemIdentifier, name string) *WritableBook {
	w := &WritableBook{}
	w.Init(w, identifier, name)
	return w
}

func (w *WritableBook) Clone() Item {
	c := *w
	c.Pages = append([]WritableBookPage(nil), w.Pages...)
	c.rebind(&c)
	return &c
}
