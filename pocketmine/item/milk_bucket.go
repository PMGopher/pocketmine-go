package item

import (
	"pocketmine-go/pocketmine/entity/effect"
)

// MilkBucket is a port of pocketmine\item\MilkBucket. GetResidue (VanillaItems.BUCKET()) and
// CanStartUsingItem aren't ported - see Food's and the Item interface's doc comments.
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

func (m *MilkBucket) GetAdditionalEffects() []*effect.EffectInstance { return nil }

// OnConsume is a port of MilkBucket::onConsume: clears every effect.
func (m *MilkBucket) OnConsume(consumer effect.Living) { consumer.GetEffects().Clear() }
