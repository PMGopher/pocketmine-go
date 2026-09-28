package item

// Fertilizer is a port of pocketmine\item\Fertilizer (bone meal). It adds no state or overrides of
// its own in PHP either: blocks that grow with bone meal check for it with `instanceof Fertilizer`.
type Fertilizer struct {
	ItemBase
}

func NewFertilizer(identifier ItemIdentifier, name string) *Fertilizer {
	f := &Fertilizer{}
	f.Init(f, identifier, name)
	return f
}

func (f *Fertilizer) Clone() Item {
	c := *f
	c.rebind(&c)
	return &c
}
