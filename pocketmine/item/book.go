package item

// Book is a port of pocketmine\item\Book - an empty class body in the PHP original too.
type Book struct {
	ItemBase
}

func NewBook(identifier ItemIdentifier, name string) *Book {
	b := &Book{}
	b.Init(b, identifier, name)
	return b
}

func (b *Book) Clone() Item {
	c := *b
	c.rebind(&c)
	return &c
}
