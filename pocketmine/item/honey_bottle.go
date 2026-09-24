package item

import (
	"pocketmine-go/pocketmine/entity/effect"
)

// HoneyBottle is a port of pocketmine\item\HoneyBottle. GetResidue (VanillaItems.GLASS_BOTTLE())
// isn't ported - see Food's doc comment.
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

// OnConsume is a port of HoneyBottle::onConsume: cures poison.
func (h *HoneyBottle) OnConsume(consumer effect.Living) {
	consumer.GetEffects().Remove(effect.VanillaPoison())
}
