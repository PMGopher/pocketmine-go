package projectile

import (
	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/particle"
)

// itemTranslator backs itemBreakParticle - convert.ItemTranslator is stateless.
var itemTranslator = convert.NewItemTranslator()

// itemBreakParticle is PHP's `new ItemBreakParticle($item)`: the particle is keyed by the item's
// network ID (see ItemBreakParticle::encode).
func itemBreakParticle(it item.Item) particle.ItemBreakParticle {
	networkID, _, _, _ := itemTranslator.ToNetworkID(it)
	return particle.ItemBreakParticle{ItemTypeID: int(networkID)}
}

// Egg is a port of pocketmine\entity\projectile\Egg.
type Egg struct {
	Throwable
}

// NewEgg is a port of Egg::__construct (shootingEntity may be nil).
func NewEgg(location entity.Location, shootingEntity world.Entity, tag *nbt.CompoundTag) *Egg {
	e := &Egg{}
	e.ConstructProjectile(e, location, shootingEntity, tag)
	return e
}

func (e *Egg) GetNetworkTypeID() string { return entity.EntityIDEgg }

//TODO: spawn chickens on collision

// OnHit is a port of Egg::onHit.
func (e *Egg) OnHit(event entityevent.ProjectileHit) {
	for i := 0; i < 6; i++ {
		e.GetWorld().AddParticle(e.GetPosition(), itemBreakParticle(item.VanillaEgg()))
	}
}
