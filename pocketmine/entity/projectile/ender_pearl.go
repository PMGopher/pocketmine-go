package projectile

import (
	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/particle"
	"pocketmine-go/pocketmine/world/sound"
)

// EnderPearl is a port of pocketmine\entity\projectile\EnderPearl.
type EnderPearl struct {
	Throwable
}

// NewEnderPearl is a port of EnderPearl::__construct (shootingEntity may be nil).
func NewEnderPearl(location entity.Location, shootingEntity world.Entity, tag *nbt.CompoundTag) *EnderPearl {
	e := &EnderPearl{}
	e.ConstructProjectile(e, location, shootingEntity, tag)
	return e
}

func (e *EnderPearl) GetNetworkTypeID() string { return entity.EntityIDEnderPearl }

// teleportable is the surface EnderPearl needs from its owner (Entity::teleport).
type teleportable interface {
	Teleport(pos math.Vector3) bool
}

// OnHit is a port of EnderPearl::onHit: the thrower is teleported to where the pearl landed and
// takes fall damage.
func (e *EnderPearl) OnHit(event entityevent.ProjectileHit) {
	owner := e.GetOwningEntity()
	if owner != nil {
		//TODO: check end gateways (when they are added)
		//TODO: spawn endermites at origin

		w := e.GetWorld()
		origin := owner.GetPosition()
		w.AddParticle(origin, particle.EndermanTeleportParticle{})
		w.AddSound(origin, sound.EndermanTeleportSound{})
		target := event.GetRayTraceResult().HitVector
		if t, ok := owner.(teleportable); ok {
			t.Teleport(target)
		}
		w.AddSound(target, sound.EndermanTeleportSound{})

		owner.Attack(entityevent.NewEntityDamageEvent(owner, entityevent.CauseFall, 5, nil))
	}
}
