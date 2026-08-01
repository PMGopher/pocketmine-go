package item

// HoneyBottle is a port of pocketmine\item\HoneyBottle. GetResidue (VanillaItems.GLASS_BOTTLE())
// and OnConsume aren't ported - see Food's and the Item interface's doc comments.
type HoneyBottle struct {
	Food
}

func NewHoneyBottle(identifier ItemIdentifier, name string) *HoneyBottle {
	h := &HoneyBottle{}
	h.Init(h, identifier, name)
	return h
}

func (h *HoneyBottle) Clone() Item {
	c := *h
	c.rebind(&c)
	return &c
}

func (h *HoneyBottle) GetMaxStackSize() int { return 16 }

func (h *HoneyBottle) RequiresHunger() bool { return false }

func (h *HoneyBottle) GetFoodRestore() int { return 6 }

func (h *HoneyBottle) GetSaturationRestore() float64 { return 1.2 }
