package item

// Fertilizer is a port of pocketmine\item\Fertilizer (bone meal) - adds no state or overrides of
// its own in PHP either, existing solely to be checked with `instanceof Fertilizer` by
// fertilizer-driven block growth (Sapling, TorchflowerCrop, PitcherCrop, Crops, SweetBerryBush,
// CocoaBlock, etc.) - none of which wire this up yet, since their grow() logic is still blocked
// on the unported block registry regardless (see e.g. Sapling.grow's doc comment).
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
