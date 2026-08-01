package item

// MelonSeeds is a port of pocketmine\item\MelonSeeds. GetBlock (should return VanillaBlocks.MELON_STEM())
// isn't ported - see StringItem's doc comment for why.
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
