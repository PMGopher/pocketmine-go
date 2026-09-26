package item

import (
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/world/sound"

	runtime "pocketmine-go/pocketmine/data/runtime"
)

// Potion is a port of pocketmine\item\Potion. GetResidue (VanillaItems.GLASS_BOTTLE()) and
// CanStartUsingItem (needs a real Player) aren't ported - see the Item interface's doc comment.
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

// soundBroadcaster is the optional surface OnConsume needs to play a sound from the consumer
// (Entity::broadcastSound). *entity.Entity satisfies it.
type soundBroadcaster interface {
	BroadcastSound(s sound.Sound)
}

// OnConsume is a port of Potion::onConsume.
func (p *Potion) OnConsume(consumer effect.Living) {
	if b, ok := consumer.(soundBroadcaster); ok {
		b.BroadcastSound(sound.BottleEmptySound{})
	}
}

// GetAdditionalEffects is a port of Potion::getAdditionalEffects.
func (p *Potion) GetAdditionalEffects() []*effect.EffectInstance {
	//TODO: check CustomPotionEffects NBT
	return p.PotionTypeValue.GetEffects()
}

// IsWaterPotion reports whether this is a water bottle (getType() === PotionType::WATER); the
// block package's cauldrons check it through a small interface since they can't import item.
func (p *Potion) IsWaterPotion() bool { return p.PotionTypeValue == PotionTypeWater }
