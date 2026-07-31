package block

// HardenedGlass is a port of pocketmine\block\HardenedGlass.
type HardenedGlass struct {
	Transparent
}

func NewHardenedGlass(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *HardenedGlass {
	h := &HardenedGlass{Transparent{NewBlock(idInfo, name, typeInfo)}}
	h.Init(h)
	return h
}

func (h *HardenedGlass) Clone() Behavior {
	c := *h
	c.rebind(&c)
	return &c
}
