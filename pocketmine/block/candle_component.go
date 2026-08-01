package block

import (
	"pocketmine-go/pocketmine/world/sound"
)

// itemTypeIDsFireCharge and itemTypeIDsFlintAndSteel mirror item.FIRE_CHARGE/item.FLINT_AND_STEEL
// (pocketmine-go/pocketmine/item, not yet ported) - same reasoning as itemTypeIDsHoneycomb in
// copper_material.go.
const (
	itemTypeIDsFireCharge    = 20260
	itemTypeIDsFlintAndSteel = 20107
)

// Durable is a forward-compatible marker for pocketmine\item\Durable - same pattern as the Axe
// interface in wood.go.
type Durable interface {
	ApplyDamage(amount int) bool
}

// CandleComponent is a port of pocketmine\block\utils\CandleTrait (which itself uses
// LightableTrait - embedded here the same way).
type CandleComponent struct {
	LightableComponent
}

// GetBaseLightLevel is a port of CandleTrait::getLightLevel (aliased `getBaseLightLevel` in the
// PHP Candle class, which multiplies it by count).
func (c *CandleComponent) GetBaseLightLevel() int {
	if c.Lit {
		return 3
	}
	return 0
}

// OnInteractCandle is a port of CandleTrait::onInteract. Concrete candle block types call this
// from their own OnInteract - see CopperComponent.OnInteractCopper's doc comment for why this
// can't just be inherited the way a PHP trait method can.
//
// The `$item->hasEnchantment(VanillaEnchantments::FIRE_ASPECT())` branch of the lighting
// condition is dropped: the enchantment package isn't ported yet, so it's treated as always
// false, same as every other HasEnchantment check in this port.
func (c *CandleComponent) OnInteractCandle(self Behavior, position Position, item Item) bool {
	world, err := position.GetWorld()
	if err != nil {
		return false
	}

	if item.GetTypeId() == itemTypeIDsFireCharge || item.GetTypeId() == itemTypeIDsFlintAndSteel {
		if c.Lit {
			return true
		}
		if durable, ok := item.(Durable); ok {
			durable.ApplyDamage(1)
		} else if item.GetTypeId() == itemTypeIDsFireCharge {
			item.Pop()
			world.AddSound(position.AsVector3(), sound.BlazeShootSound{})
		}
		world.AddSound(position.AsVector3(), sound.FlintSteelSound{})
		c.Lit = true
		if err := world.SetBlock(position, self); err != nil {
			panic(err)
		}
		return true
	}

	if item.IsNull() {
		if !c.Lit {
			return true
		}
		world.AddSound(position.AsVector3(), sound.FireExtinguishSound{})
		c.Lit = false
		if err := world.SetBlock(position, self); err != nil {
			panic(err)
		}
		return true
	}

	return false
}

// OnProjectileHitCandle is a port of CandleTrait::onProjectileHit.
func (c *CandleComponent) OnProjectileHitCandle(self Behavior, position Position, projectile Projectile) {
	if !c.Lit && projectile.IsOnFire() {
		c.Lit = true
		world, err := position.GetWorld()
		if err != nil {
			return
		}
		if err := world.SetBlock(position, self); err != nil {
			panic(err)
		}
	}
}
