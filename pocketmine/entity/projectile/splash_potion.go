package projectile

import (
	stdmath "math"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/entity/object"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/particle"
	"pocketmine-go/pocketmine/world/sound"
)

// TagPotionID mirrors SplashPotion::TAG_POTION_ID.
const TagPotionID = "PotionId" //TAG_Short

// SplashPotion is a port of pocketmine\entity\projectile\SplashPotion.
type SplashPotion struct {
	Throwable

	linger     bool
	potionType item.PotionType
}

// NewSplashPotion is a port of SplashPotion::__construct (shootingEntity may be nil).
func NewSplashPotion(location entity.Location, shootingEntity world.Entity, potionType item.PotionType, tag *nbt.CompoundTag) *SplashPotion {
	s := &SplashPotion{potionType: potionType}
	s.ConstructProjectile(s, location, shootingEntity, tag)
	return s
}

func (s *SplashPotion) GetNetworkTypeID() string { return entity.EntityIDSplashPotion }

func (s *SplashPotion) GetInitialGravity() float64 { return 0.05 }

// SaveNBT is a port of SplashPotion::saveNBT.
func (s *SplashPotion) SaveNBT() *nbt.CompoundTag {
	tag := s.Throwable.SaveNBT()
	tag.SetShort(TagPotionID, nbt.ShortTag(item.PotionTypeIdMapInstance.ToID(s.GetPotionType())))

	return tag
}

func (s *SplashPotion) GetResultDamage() int { return -1 } //no damage

// effectLiving is the surface a splash potion needs from a Living it affects.
type effectLiving interface {
	effect.Living
	IsLiving() bool
	GetEyePos() math.Vector3
}

// fireSides is the promoted-from-*block.Block surface water potions need to put out fire.
type fireSides interface {
	GetSide(side math.Facing, step int) block.Behavior
	HasTypeTag(tag string) bool
}

// OnHit is a port of SplashPotion::onHit: effects are applied to living entities within 4 blocks
// (scaled by distance), or a lingering potion leaves an area effect cloud; a water bottle puts out
// fire where it lands.
func (s *SplashPotion) OnHit(event entityevent.ProjectileHit) {
	effects := s.GetPotionEffects()
	hasEffects := true

	var splash particle.PotionSplashParticle
	if len(effects) == 0 {
		splash = particle.PotionSplashParticle{Color: particle.DefaultPotionSplashColor}
		hasEffects = false
	} else {
		var colors []color.Color
		for _, e := range effects {
			level := e.GetEffectLevel()
			for j := 0; j < level; j++ {
				colors = append(colors, e.GetColor())
			}
		}
		splash = particle.PotionSplashParticle{Color: color.Mix(colors[0], colors[1:]...)}
	}

	w := s.GetWorld()
	w.AddParticle(s.GetPosition(), splash)
	s.BroadcastSound(sound.PotionSplashSound{})

	if !s.WillLinger() {
		if hasEffects {
			for _, e := range w.GetCollidingEntities(s.BoundingBox.ExpandedCopy(4.125, 2.125, 4.125), s) {
				living, ok := e.(effectLiving)
				if !ok {
					continue
				}
				distanceSquared := living.GetEyePos().DistanceSquared(s.GetPosition())
				if distanceSquared > 16 { //4 blocks
					continue
				}

				distanceMultiplier := 1 - (stdmath.Sqrt(distanceSquared) / 4)
				if hitEntity, ok := event.(*entityevent.ProjectileHitEntityEvent); ok && hitEntity.GetEntityHit() == entityevent.Entity(living) {
					distanceMultiplier = 1.0
				}

				for _, instance := range s.GetPotionEffects() {
					//getPotionEffects() is used to get COPIES to avoid accidentally modifying the same effect instance already applied to another entity

					if !effect.IsInstant(instance.GetType()) {
						newDuration := int(stdmath.Round(float64(instance.GetDuration()) * 0.75 * distanceMultiplier))
						if newDuration < 20 {
							continue
						}
						instance.SetDuration(newDuration)
						living.GetEffects().Add(instance)
					} else {
						instance.GetType().ApplyEffect(living, instance, distanceMultiplier, s)
					}
				}
			}
		}
	} else {
		cloud := object.NewAreaEffectCloud(entity.LocationFromObject(s.GetPosition().Floor().Add(0.5, 0.5, 0.5), w, 0, 0), nil)
		for _, instance := range s.potionType.GetEffects() {
			cloud.GetEffects().Add(instance)
		}
		if owner := s.GetOwningEntity(); owner != nil && !owner.IsClosed() {
			cloud.SetOwningEntity(owner)
		}
		cloud.SpawnToAll()
	}

	if hitBlock, ok := event.(*entityevent.ProjectileHitBlockEvent); !hasEffects && ok && s.GetPotionType() == item.PotionTypeWater {
		if blockHit, ok := hitBlock.GetBlockHit().(fireSides); ok {
			blockIn := blockHit.GetSide(hitBlock.GetRayTraceResult().HitFace, 1)

			if in, ok := blockIn.(fireSides); ok {
				if in.HasTypeTag(block.BlockTypeTagsFire) {
					_ = w.SetBlock(blockIn.GetPosition(), block.VanillaAir())
				}
				for _, facing := range []math.Facing{math.North, math.South, math.West, math.East} {
					horizontalSide := in.GetSide(facing, 1)
					if side, ok := horizontalSide.(fireSides); ok && side.HasTypeTag(block.BlockTypeTagsFire) {
						_ = w.SetBlock(horizontalSide.GetPosition(), block.VanillaAir())
					}
				}
			}
		}
	}
}

func (s *SplashPotion) GetPotionType() item.PotionType { return s.potionType }

func (s *SplashPotion) SetPotionType(t item.PotionType) {
	s.potionType = t
	s.MarkNetworkPropertiesDirty()
}

// WillLinger returns whether this splash potion will create an area-effect cloud when it lands.
func (s *SplashPotion) WillLinger() bool { return s.linger }

// SetLinger sets whether this splash potion will create an area-effect-cloud when it lands.
func (s *SplashPotion) SetLinger(value bool) {
	s.linger = value
	s.MarkNetworkPropertiesDirty()
}

// GetPotionEffects returns fresh copies of the potion's effects.
func (s *SplashPotion) GetPotionEffects() []*effect.EffectInstance {
	return s.potionType.GetEffects()
}

// SyncNetworkData is a port of SplashPotion::syncNetworkData.
func (s *SplashPotion) SyncNetworkData(properties *entity.MetadataCollection) {
	s.Throwable.SyncNetworkData(properties)

	properties.SetShort(entity.MetadataPotionAuxValue, int16(item.PotionTypeIdMapInstance.ToID(s.potionType)))
	properties.SetGenericFlag(entity.FlagLinger, s.linger)
}

// IsWaterPotion reports `$projectile->getPotionType() === PotionType::WATER`, for blocks that react
// to water splash potions (Campfire::onProjectileHit).
func (s *SplashPotion) IsWaterPotion() bool { return s.potionType == item.PotionTypeWater }
