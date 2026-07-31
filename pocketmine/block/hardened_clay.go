package block

// HardenedClay is a port of pocketmine\block\HardenedClay.
type HardenedClay struct {
	Opaque
}

func NewHardenedClay(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *HardenedClay {
	h := &HardenedClay{Opaque{NewBlock(idInfo, name, typeInfo)}}
	h.Init(h)
	return h
}

func (h *HardenedClay) Clone() Behavior {
	c := *h
	c.rebind(&c)
	return &c
}
