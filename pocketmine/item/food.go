package item

import (
	"pocketmine-go/pocketmine/entity/effect"
)

// Food is a port of pocketmine\item\Food. Concrete food items (Apple, Bread, Carrot, etc.) embed
// this and must define their own GetFoodRestore/GetSaturationRestore - FoodSource declares those
// as abstract in PHP too (Food itself doesn't implement them), so there's no default here either.
//
// Not ported: GetResidue (should return VanillaItems.AIR() - needs the unported item registry,
// same category of gap as ItemBlock's registry-dependent pieces) and CanStartUsingItem (needs a
// real Player - see the Item interface's doc comment on Player/Entity-interaction methods).
type Food struct {
	ItemBase
}

func (f *Food) RequiresHunger() bool { return true }

// GetAdditionalEffects is Food::getAdditionalEffects' default: no extra effects.
func (f *Food) GetAdditionalEffects() []*effect.EffectInstance { return nil }

// OnConsume is Food::onConsume's default: nothing happens.
func (f *Food) OnConsume(consumer effect.Living) {}
