package item

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// Potion is a port of pocketmine\item\Potion.
//
// GetAdditionalEffects (potionType.GetEffects()), GetResidue (VanillaItems.GLASS_BOTTLE()), and
// OnConsume/CanStartUsingItem all need pieces that aren't ported (EffectInstance, the item
// registry, and a real Player/Living respectively - see PotionType's and the Item interface's
// doc comments), so none of those are ported here; only the PotionType state and its describeState
// round trip are real.
type Potion struct {
	ItemBase

	PotionTypeValue PotionType
}

func NewPotion(identifier ItemIdentifier, name string) *Potion {
	p := &Potion{PotionTypeValue: PotionTypeWater}
	p.Init(p, identifier, name)
	return p
}

func (p *Potion) Clone() Item {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *Potion) GetType() PotionType { return p.PotionTypeValue }

func (p *Potion) SetType(t PotionType) { p.PotionTypeValue = t }

func (p *Potion) GetMaxStackSize() int { return 1 }

func (p *Potion) describeState(w runtime.DataDescriber) {
	t := int(p.PotionTypeValue)
	w.BoundedIntAuto(int(PotionTypeWater), int(PotionTypeStrongSlowness), &t)
	p.PotionTypeValue = PotionType(t)
}
