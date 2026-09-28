package item

import "pocketmine-go/pocketmine/block"

// MelonSeeds is a port of pocketmine\item\MelonSeeds.
type MelonSeeds struct {
	ItemBase
}

func NewMelonSeeds(identifier ItemIdentifier, name string) *MelonSeeds {
	m := &MelonSeeds{}
	m.Init(m, identifier, name)
	return m
}

func (m *MelonSeeds) Clone() Item {
	c := *m
	c.rebind(&c)
	return &c
}

// GetBlock is a port of MelonSeeds::getBlock.
func (x *MelonSeeds) GetBlock() block.Behavior { return block.VanillaBlock("melon_stem") }
