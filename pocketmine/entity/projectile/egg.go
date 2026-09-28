package projectile

import (
	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world"
)

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
		e.GetWorld().AddParticle(e.GetPosition(), convert.NewItemBreakParticle(item.VanillaEgg()))
	}
}
