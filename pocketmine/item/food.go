package item

// Food is a port of pocketmine\item\Food. Concrete food items (Apple, Bread, Carrot, etc.) embed
// this and must define their own GetFoodRestore/GetSaturationRestore - FoodSource declares those
// as abstract in PHP too (Food itself doesn't implement them), so there's no default here either.
//
// Not ported: GetResidue (should return VanillaItems.AIR() - needs the unported item registry,
// same category of gap as ItemBlock's registry-dependent pieces), GetAdditionalEffects's real
// return value (EffectInstance isn't ported, so this is a void marker method like
// block.BaseCake.GetAdditionalEffects), OnConsume (a no-op in the PHP base anyway, and its Living
// parameter isn't ported), and CanStartUsingItem (needs a real Player - see the Item interface's
// doc comment on Player/Entity-interaction methods).
type Food struct {
	ItemBase
}

func (f *Food) RequiresHunger() bool { return true }

func (f *Food) GetAdditionalEffects() {}
