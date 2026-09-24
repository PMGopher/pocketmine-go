package entity

import "pocketmine-go/pocketmine/entity/effect"

// Consumable is a port of pocketmine\entity\Consumable: things which can be consumed by an entity
// (food and potion items, cake blocks). The consumer is effect.Living rather than *Living so
// consumables in the item and block packages (which pocketmine/entity imports) can implement it.
type Consumable interface {
	// GetAdditionalEffects returns effects to be added to the consumer on consumption.
	GetAdditionalEffects() []*effect.EffectInstance

	// OnConsume is called when this Consumable is consumed by mob, after standard resulting effects
	// have been applied.
	OnConsume(consumer effect.Living)
}
