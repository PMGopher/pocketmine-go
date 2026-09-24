package entity

import (
	stdmath "math"
	"math/rand/v2"

	"pocketmine-go/pocketmine/entity/animation"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// Squid is a port of pocketmine\entity\Squid.
type Squid struct {
	WaterAnimal

	// SwimDirection is nil while the squid has no direction (PHP's ?Vector3).
	SwimDirection *math.Vector3
	SwimSpeed     float64

	switchDirectionTicker int
}

// NewSquid is a port of Squid::__construct (Living's constructor).
func NewSquid(location Location, tag *nbt.CompoundTag) *Squid {
	s := &Squid{SwimSpeed: 0.1}
	s.ConstructLiving(s, location, tag)
	return s
}

func (s *Squid) GetNetworkTypeID() string { return EntityIDSquid }

func (s *Squid) GetInitialSizeInfo() EntitySizeInfo { return NewEntitySizeInfo(0.8, 0.8) }

// InitEntity is a port of Squid::initEntity.
func (s *Squid) InitEntity(tag *nbt.CompoundTag) {
	s.SetMaxHealth(10)
	s.WaterAnimal.InitEntity(tag)
}

func (s *Squid) GetName() string { return "Squid" }

// Attack is a port of Squid::attack: a hit squid flees and inks.
func (s *Squid) Attack(source entityevent.DamageSource) {
	s.WaterAnimal.Attack(source)
	if source.IsCancelled() {
		return
	}

	if byEntity, ok := AsDamageByEntity(source); ok {
		s.SwimSpeed = float64(150+rand.IntN(201)) / 2000
		if e := byEntity.GetDamager(); e != nil {
			direction := s.location.SubtractVector(e.GetPosition()).Normalize()
			s.SwimDirection = &direction
		}

		s.BroadcastAnimation(animation.SquidInkCloudAnimation{Squid: s}, nil)
	}
}

func (s *Squid) generateRandomDirection() math.Vector3 {
	return math.NewVector3(float64(rand.IntN(2001)-1000)/1000, float64(rand.IntN(1001)-500)/1000, float64(rand.IntN(2001)-1000)/1000)
}

// EntityBaseTick is a port of Squid::entityBaseTick.
func (s *Squid) EntityBaseTick(tickDiff int) bool {
	if s.closed {
		return false
	}

	s.switchDirectionTicker++
	if s.switchDirectionTicker == 100 {
		s.switchDirectionTicker = 0
		if rand.IntN(101) < 50 {
			s.SwimDirection = nil
		}
	}

	hasUpdate := s.WaterAnimal.EntityBaseTick(tickDiff)

	if s.IsAlive() {
		if s.location.Y > 62 && s.SwimDirection != nil {
			s.SwimDirection.Y = -0.5
		}

		inWater := s.IsUnderwater()
		s.SetHasGravity(!inWater)
		if !inWater {
			s.SwimDirection = nil
		} else if s.SwimDirection != nil {
			if s.motion.LengthSquared() <= s.SwimDirection.LengthSquared() {
				s.motion = s.SwimDirection.Multiply(s.SwimSpeed)
			}
		} else {
			direction := s.generateRandomDirection()
			s.SwimDirection = &direction
			s.SwimSpeed = float64(50+rand.IntN(51)) / 2000
		}

		f := stdmath.Sqrt(s.motion.X*s.motion.X + s.motion.Z*s.motion.Z)
		s.SetRotation(
			-stdmath.Atan2(s.motion.X, s.motion.Z)*180/stdmath.Pi,
			-stdmath.Atan2(f, s.motion.Y)*180/stdmath.Pi,
		)
	}

	return hasUpdate
}

// GetDrops is a port of Squid::getDrops.
func (s *Squid) GetDrops() []item.Item {
	inkSac := item.VanillaInkSac()
	inkSac.SetCount(1 + rand.IntN(3))
	return []item.Item{inkSac}
}

func (s *Squid) GetPickedItem() item.Item { return item.VanillaSquidSpawnEgg() }
