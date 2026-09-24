package projectile

import (
	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/particle"
)

// Snowball is a port of pocketmine\entity\projectile\Snowball.
type Snowball struct {
	Throwable
}

// NewSnowball is a port of Snowball::__construct (shootingEntity may be nil).
func NewSnowball(location entity.Location, shootingEntity world.Entity, tag *nbt.CompoundTag) *Snowball {
	s := &Snowball{}
	s.ConstructProjectile(s, location, shootingEntity, tag)
	return s
}

func (s *Snowball) GetNetworkTypeID() string { return entity.EntityIDSnowball }

// OnHit is a port of Snowball::onHit.
func (s *Snowball) OnHit(event entityevent.ProjectileHit) {
	w := s.GetWorld()
	for i := 0; i < 6; i++ {
		w.AddParticle(s.GetPosition(), particle.SnowballPoofParticle{})
	}
}
