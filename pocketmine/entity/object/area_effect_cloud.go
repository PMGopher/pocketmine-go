package object

import (
	stdmath "math"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/effect"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/particle"
)

// AreaEffectCloud constants, a port of AreaEffectCloud's public constants.
const (
	AreaEffectCloudDefaultDuration            = 600 // in ticks
	AreaEffectCloudDefaultDurationChangeOnUse = 0   // in ticks
	AreaEffectCloudUpdateDelay                = 10  // in ticks
	AreaEffectCloudReapplicationDelay         = 40  // in ticks

	AreaEffectCloudDefaultRadius               = 3.0                                                              // in blocks
	AreaEffectCloudDefaultRadiusChangeOnPickup = -0.5                                                             // in blocks
	AreaEffectCloudDefaultRadiusChangeOnUse    = -0.5                                                             // in blocks
	AreaEffectCloudDefaultRadiusChangePerTick  = -(AreaEffectCloudDefaultRadius / AreaEffectCloudDefaultDuration) // in blocks
)

// AreaEffectCloud NBT keys, a port of AreaEffectCloud's TAG_* constants.
const (
	tagCloudPotionID             = "PotionId"             //TAG_Short
	tagCloudSpawnTick            = "SpawnTick"            //TAG_Long
	tagCloudDuration             = "Duration"             //TAG_Int
	tagCloudPickupCount          = "PickupCount"          //TAG_Int
	tagCloudDurationOnUse        = "DurationOnUse"        //TAG_Int
	tagCloudReapplicationDelay   = "ReapplicationDelay"   //TAG_Int
	tagCloudInitialRadius        = "InitialRadius"        //TAG_Float
	tagCloudRadius               = "Radius"               //TAG_Float
	tagCloudRadiusChangeOnPickup = "RadiusChangeOnPickup" //TAG_Float
	tagCloudRadiusOnUse          = "RadiusOnUse"          //TAG_Float
	tagCloudRadiusPerTick        = "RadiusPerTick"        //TAG_Float
	tagCloudEffects              = "mobEffects"           //TAG_List
)

// effectLiving is the surface an area effect cloud needs from a Living victim.
type effectLiving interface {
	effect.Living
	IsLiving() bool
}

// AreaEffectCloud is a port of pocketmine\entity\object\AreaEffectCloud.
type AreaEffectCloud struct {
	entity.Entity

	age int

	effectCollection *effect.EffectCollection

	// victims maps entity ID -> the age at which it can next be affected.
	victims map[int]int

	maxAge             int
	maxAgeChangeOnUse  int
	reapplicationDelay int

	pickupCount          int
	radiusChangeOnPickup float64

	initialRadius       float64
	radius              float64
	radiusChangeOnUse   float64
	radiusChangePerTick float64
}

// NewAreaEffectCloud is a port of AreaEffectCloud::__construct.
func NewAreaEffectCloud(location entity.Location, tag *nbt.CompoundTag) *AreaEffectCloud {
	c := &AreaEffectCloud{
		victims:              map[int]int{},
		maxAge:               AreaEffectCloudDefaultDuration,
		maxAgeChangeOnUse:    AreaEffectCloudDefaultDurationChangeOnUse,
		reapplicationDelay:   AreaEffectCloudReapplicationDelay,
		radiusChangeOnPickup: AreaEffectCloudDefaultRadiusChangeOnPickup,
		initialRadius:        AreaEffectCloudDefaultRadius,
		radius:               AreaEffectCloudDefaultRadius,
		radiusChangeOnUse:    AreaEffectCloudDefaultRadiusChangeOnUse,
		radiusChangePerTick:  AreaEffectCloudDefaultRadiusChangePerTick,
	}
	c.Construct(c, location, tag)
	return c
}

func (c *AreaEffectCloud) GetNetworkTypeID() string { return entity.EntityIDAreaEffectCloud }

func (c *AreaEffectCloud) GetInitialSizeInfo() entity.EntitySizeInfo {
	return entity.NewEntitySizeInfo(0.5, c.radius*2)
}

func (c *AreaEffectCloud) GetInitialDragMultiplier() float64 { return 0.0 }

func (c *AreaEffectCloud) GetInitialGravity() float64 { return 0.0 }

// InitEntity is a port of AreaEffectCloud::initEntity.
func (c *AreaEffectCloud) InitEntity(tag *nbt.CompoundTag) {
	c.Entity.InitEntity(tag)

	c.effectCollection = effect.NewEffectCollection()
	onAdd := effect.EffectAddHook(func(*effect.EffectInstance, bool) { c.MarkNetworkPropertiesDirty() })
	onRemove := effect.EffectRemoveHook(func(*effect.EffectInstance) { c.MarkNetworkPropertiesDirty() })
	c.effectCollection.GetEffectAddHooks().Add(&onAdd)
	c.effectCollection.GetEffectRemoveHooks().Add(&onRemove)
	c.effectCollection.SetEffectFilterForBubbles(func(e *effect.EffectInstance) bool { return e.IsVisible() })

	worldTime := c.GetWorld().GetTime()
	c.age = int(max(worldTime-int64(tag.GetLongOr(tagCloudSpawnTick, nbt.LongTag(worldTime))), 0))
	c.maxAge = int(tag.GetIntOr(tagCloudDuration, AreaEffectCloudDefaultDuration))
	c.maxAgeChangeOnUse = int(tag.GetIntOr(tagCloudDurationOnUse, AreaEffectCloudDefaultDurationChangeOnUse))
	c.pickupCount = int(tag.GetIntOr(tagCloudPickupCount, 0))
	c.reapplicationDelay = int(tag.GetIntOr(tagCloudReapplicationDelay, AreaEffectCloudReapplicationDelay))

	c.initialRadius = float64(tag.GetFloatOr(tagCloudInitialRadius, AreaEffectCloudDefaultRadius))
	c.SetRadius(float64(tag.GetFloatOr(tagCloudRadius, nbt.FloatTag(c.initialRadius))))
	c.radiusChangeOnPickup = float64(tag.GetFloatOr(tagCloudRadiusChangeOnPickup, AreaEffectCloudDefaultRadiusChangeOnPickup))
	c.radiusChangeOnUse = float64(tag.GetFloatOr(tagCloudRadiusOnUse, AreaEffectCloudDefaultRadiusChangeOnUse))
	c.radiusChangePerTick = float64(tag.GetFloatOr(tagCloudRadiusPerTick, AreaEffectCloudDefaultRadiusChangePerTick))

	if effectsTag, ok, _ := tag.GetListTag(tagCloudEffects); ok {
		for _, t := range effectsTag.Values() {
			e, ok := t.(*nbt.CompoundTag)
			if !ok {
				continue
			}
			effectType, ok := bedrock.EffectIdMap().FromID(int(e.GetByteOr("Id", 0)))
			if !ok {
				continue
			}
			duration := int(e.GetIntOr("Duration", 0))
			c.effectCollection.Add(effect.NewEffectInstanceFull(
				effectType,
				&duration,
				int(uint8(e.GetByteOr("Amplifier", 0))),
				e.GetByteOr("ShowParticles", 1) != 0,
				e.GetByteOr("Ambient", 0) != 0,
				nil,
				false,
			))
		}
	}
}

// SaveNBT is a port of AreaEffectCloud::saveNBT.
func (c *AreaEffectCloud) SaveNBT() *nbt.CompoundTag {
	tag := c.Entity.SaveNBT()

	tag.SetLong(tagCloudSpawnTick, nbt.LongTag(c.GetWorld().GetTime()-int64(c.age)))
	tag.SetShort(tagCloudPotionID, nbt.ShortTag(item.PotionTypeIdMapInstance.ToID(item.PotionTypeWater))) //not used, mobEffects is used exclusively in Bedrock
	tag.SetInt(tagCloudDuration, nbt.IntTag(c.maxAge))
	tag.SetInt(tagCloudDurationOnUse, nbt.IntTag(c.maxAgeChangeOnUse))
	tag.SetInt(tagCloudPickupCount, nbt.IntTag(c.pickupCount))
	tag.SetInt(tagCloudReapplicationDelay, nbt.IntTag(c.reapplicationDelay))
	tag.SetFloat(tagCloudInitialRadius, nbt.FloatTag(c.initialRadius))
	tag.SetFloat(tagCloudRadius, nbt.FloatTag(c.radius))
	tag.SetFloat(tagCloudRadiusChangeOnPickup, nbt.FloatTag(c.radiusChangeOnPickup))
	tag.SetFloat(tagCloudRadiusOnUse, nbt.FloatTag(c.radiusChangeOnUse))
	tag.SetFloat(tagCloudRadiusPerTick, nbt.FloatTag(c.radiusChangePerTick))

	if effects := c.effectCollection.All(); len(effects) > 0 {
		values := make([]nbt.Tag, 0, len(effects))
		for _, e := range effects {
			duration := e.GetDuration()
			if e.IsInfinite() {
				duration = -1
			}
			ambient := nbt.ByteTag(0)
			if e.IsAmbient() {
				ambient = 1
			}
			visible := nbt.ByteTag(0)
			if e.IsVisible() {
				visible = 1
			}
			values = append(values, nbt.NewCompoundTag().
				SetByte("Id", nbt.ByteTag(bedrock.EffectIdMap().ToID(e.GetType()))).
				SetByte("Amplifier", nbt.ByteTag(int8(e.GetAmplifier()))).
				SetInt("Duration", nbt.IntTag(duration)).
				SetByte("Ambient", ambient).
				SetByte("ShowParticles", visible))
		}
		list, err := nbt.NewListTag(values, nbt.TagCompound)
		if err != nil {
			panic(err)
		}
		tag.SetTag(tagCloudEffects, list)
	}

	return tag
}

func (c *AreaEffectCloud) IsFireProof() bool { return true }

func (c *AreaEffectCloud) CanBeCollidedWith() bool { return false }

// GetAge returns the current age of the cloud (in ticks).
func (c *AreaEffectCloud) GetAge() int { return c.age }

func (c *AreaEffectCloud) GetEffects() *effect.EffectCollection { return c.effectCollection }

// GetInitialRadius returns the initial radius (in blocks).
func (c *AreaEffectCloud) GetInitialRadius() float64 { return c.initialRadius }

// GetRadius returns the current radius (in blocks).
func (c *AreaEffectCloud) GetRadius() float64 { return c.radius }

// SetRadius sets the current radius (in blocks) - protected in PHP.
func (c *AreaEffectCloud) SetRadius(radius float64) {
	c.radius = radius
	c.SetSize(c.GetInitialSizeInfo())
	c.MarkNetworkPropertiesDirty()
}

func (c *AreaEffectCloud) GetRadiusChangeOnPickup() float64 { return c.radiusChangeOnPickup }

func (c *AreaEffectCloud) SetRadiusChangeOnPickup(v float64) { c.radiusChangeOnPickup = v }

func (c *AreaEffectCloud) GetRadiusChangeOnUse() float64 { return c.radiusChangeOnUse }

func (c *AreaEffectCloud) SetRadiusChangeOnUse(v float64) { c.radiusChangeOnUse = v }

func (c *AreaEffectCloud) GetRadiusChangePerTick() float64 { return c.radiusChangePerTick }

func (c *AreaEffectCloud) SetRadiusChangePerTick(v float64) { c.radiusChangePerTick = v }

// GetMaxAge returns the age at which the cloud will despawn.
func (c *AreaEffectCloud) GetMaxAge() int { return c.maxAge }

func (c *AreaEffectCloud) SetMaxAge(maxAge int) { c.maxAge = maxAge }

func (c *AreaEffectCloud) GetMaxAgeChangeOnUse() int { return c.maxAgeChangeOnUse }

func (c *AreaEffectCloud) SetMaxAgeChangeOnUse(v int) { c.maxAgeChangeOnUse = v }

// GetReapplicationDelay returns the delay (in ticks) before an entity can be affected again.
func (c *AreaEffectCloud) GetReapplicationDelay() int { return c.reapplicationDelay }

func (c *AreaEffectCloud) SetReapplicationDelay(delay int) { c.reapplicationDelay = delay }

// EntityBaseTick is a port of AreaEffectCloud::entityBaseTick.
func (c *AreaEffectCloud) EntityBaseTick(tickDiff int) bool {
	hasUpdate := c.Entity.EntityBaseTick(tickDiff)

	c.age += tickDiff
	radius := c.radius + (c.radiusChangePerTick * float64(tickDiff))
	if radius < 0.5 {
		c.FlagForDespawn()
		return true
	}
	c.SetRadius(radius)
	if c.age >= AreaEffectCloudUpdateDelay && (c.age%AreaEffectCloudUpdateDelay) == 0 {
		if c.age > c.maxAge {
			c.FlagForDespawn()
			return true
		}
		for entityID, expiration := range c.victims {
			if c.age >= expiration {
				delete(c.victims, entityID)
			}
		}

		var affected []entityevent.Entity
		radiusChange := 0.0
		maxAgeChange := 0
		pos := c.GetPosition()
		for _, e := range c.GetWorld().GetCollidingEntities(c.GetBoundingBox(), c) {
			living, ok := e.(effectLiving)
			if !ok {
				continue
			}
			if _, isVictim := c.victims[living.GetID()]; isVictim {
				continue
			}
			entityPosition := living.GetPosition()
			xDiff := entityPosition.X - pos.X
			zDiff := entityPosition.Z - pos.Z
			if xDiff*xDiff+zDiff*zDiff > c.radius*c.radius {
				continue
			}
			affected = append(affected, living)
			if c.radiusChangeOnUse != 0.0 {
				radiusChange += c.radiusChangeOnUse
				if c.radius+radiusChange <= 0 {
					break
				}
			}
			if c.maxAgeChangeOnUse != 0 {
				maxAgeChange += c.maxAgeChangeOnUse
				if c.maxAge+maxAgeChange <= 0 {
					break
				}
			}
		}
		if len(affected) == 0 {
			return hasUpdate
		}

		ev := entityevent.NewAreaEffectCloudApplyEvent(c, affected)
		ev.Call()
		if ev.IsCancelled() {
			return hasUpdate
		}

		for _, e := range ev.GetAffectedEntities() {
			living, ok := e.(effectLiving)
			if !ok {
				continue
			}
			for _, instance := range c.effectCollection.All() {
				instance = instance.Clone() //avoid accidental modification
				if effect.IsInstant(instance.GetType()) {
					instance.GetType().ApplyEffect(living, instance, 0.5, c)
				} else {
					living.GetEffects().Add(instance.SetDuration(int(stdmath.Round(float64(instance.GetDuration()) / 4))))
				}
			}
			if c.reapplicationDelay != 0 {
				c.victims[living.GetID()] = c.age + c.reapplicationDelay
			}
		}

		newRadius := c.radius + radiusChange
		newMaxAge := c.maxAge + maxAgeChange
		if newRadius <= 0 || newMaxAge <= 0 {
			c.FlagForDespawn()
			return true
		}
		c.SetRadius(newRadius)
		c.SetMaxAge(newMaxAge)
		hasUpdate = true
	}

	return hasUpdate
}

// SyncNetworkData is a port of AreaEffectCloud::syncNetworkData.
func (c *AreaEffectCloud) SyncNetworkData(properties *entity.MetadataCollection) {
	c.Entity.SyncNetworkData(properties)

	//visual properties
	properties.SetFloat(entity.MetadataAreaEffectCloudRadius, float32(c.radius))
	bubbleColor := particle.DefaultPotionSplashColor
	if len(c.effectCollection.All()) > 0 {
		bubbleColor = c.effectCollection.GetBubbleColor()
	}
	properties.SetInt(entity.MetadataPotionColor, bubbleColor.ToARGB())

	//these are properties the client expects, and are used for client-sided logic, which we don't want
	properties.SetByte(entity.MetadataPotionAmbient, 0)
	properties.SetInt(entity.MetadataAreaEffectCloudDuration, -1)
	properties.SetFloat(entity.MetadataAreaEffectCloudRadiusChangeOnPickup, 0)
	properties.SetFloat(entity.MetadataAreaEffectCloudRadiusPerTick, 0)
	properties.SetInt(entity.MetadataAreaEffectCloudSpawnTime, 0)
	properties.SetFloat(entity.MetadataAreaEffectCloudPickupCount, 0)
	properties.SetInt(entity.MetadataAreaEffectCloudWaiting, 0)
}

// DestroyCycles is a port of AreaEffectCloud::destroyCycles (wipe out callback refs).
func (c *AreaEffectCloud) DestroyCycles() {
	c.effectCollection = effect.NewEffectCollection()
	c.Entity.DestroyCycles()
}
