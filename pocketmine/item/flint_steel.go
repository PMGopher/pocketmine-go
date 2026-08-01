package item

// FlintSteel is a port of pocketmine\item\FlintSteel.
//
// OnInteractBlock (igniting fire on the clicked air block, then applyDamage(1)) needs
// VanillaBlocks.FIRE() and a real Player/Block/World - see the Item interface's doc comment for
// why those Player/Entity-interaction methods aren't part of Item here yet, so it's not ported.
// GetMaxDurability and the rest of the Durable/Tool machinery (damage tracking, NBT round trip)
// are fully real, though.
type FlintSteel struct {
	Tool
}

func NewFlintSteel(identifier ItemIdentifier, name string) *FlintSteel {
	f := &FlintSteel{}
	f.Init(f, identifier, name)
	return f
}

func (f *FlintSteel) Clone() Item {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FlintSteel) GetMaxDurability() int { return 65 }
