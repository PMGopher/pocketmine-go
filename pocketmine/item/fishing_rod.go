package item

// FishingRod is a port of pocketmine\item\FishingRod. Casting and reeling are //TODO in PHP too, so
// only the durability is ported.
type FishingRod struct {
	Durable
}

func NewFishingRod(identifier ItemIdentifier, name string, enchantmentTags ...string) *FishingRod {
	f := &FishingRod{}
	f.Init(f, identifier, name)
	f.enchantmentTags = enchantmentTags
	return f
}

func (f *FishingRod) Clone() Item {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FishingRod) GetMaxStackSize() int { return 1 }

func (f *FishingRod) GetMaxDurability() int { return 384 }
