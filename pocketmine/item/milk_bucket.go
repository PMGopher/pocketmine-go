package item

// MilkBucket is a port of pocketmine\item\MilkBucket. GetResidue (VanillaItems.BUCKET()),
// OnConsume, and CanStartUsingItem aren't ported - see Food's and the Item interface's doc
// comments.
type MilkBucket struct {
	ItemBase
}

func NewMilkBucket(identifier ItemIdentifier, name string) *MilkBucket {
	m := &MilkBucket{}
	m.Init(m, identifier, name)
	return m
}

func (m *MilkBucket) Clone() Item {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *MilkBucket) GetMaxStackSize() int { return 1 }

func (m *MilkBucket) GetAdditionalEffects() {}
