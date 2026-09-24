package projectile

import (
	"math/rand/v2"

	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/particle"
	"pocketmine-go/pocketmine/world/sound"
)

// ExperienceBottle is a port of pocketmine\entity\projectile\ExperienceBottle.
type ExperienceBottle struct {
	Throwable
}

// NewExperienceBottle is a port of ExperienceBottle::__construct (shootingEntity may be nil).
func NewExperienceBottle(location entity.Location, shootingEntity world.Entity, tag *nbt.CompoundTag) *ExperienceBottle {
	b := &ExperienceBottle{}
	b.ConstructProjectile(b, location, shootingEntity, tag)
	return b
}

func (b *ExperienceBottle) GetNetworkTypeID() string { return entity.EntityIDXPBottle }

func (b *ExperienceBottle) GetInitialGravity() float64 { return 0.07 }

func (b *ExperienceBottle) GetResultDamage() int { return -1 }

// OnHit is a port of ExperienceBottle::onHit.
func (b *ExperienceBottle) OnHit(event entityevent.ProjectileHit) {
	w := b.GetWorld()
	w.AddParticle(b.GetPosition(), particle.PotionSplashParticle{Color: particle.DefaultPotionSplashColor})
	b.BroadcastSound(sound.PotionSplashSound{})

	w.DropExperience(b.GetPosition(), 3+rand.IntN(9))
}
