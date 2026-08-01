package item

// FishingRod is a port of pocketmine\item\FishingRod. Casting/reeling logic is marked //TODO even
// in the PHP original (needs the unported entity package), so only its Durable durability state
// is ported here.
type FishingRod struct {
	Durable
}

func NewFishingRod(identifier ItemIdentifier, name string) *FishingRod {
	f := &FishingRod{}
	f.Init(f, identifier, name)
	return f
}

func (f *FishingRod) Clone() Item {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FishingRod) GetMaxStackSize() int { return 1 }

func (f *FishingRod) GetMaxDurability() int { return 384 }
