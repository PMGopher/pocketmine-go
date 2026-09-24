package entity

import (
	"reflect"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

// Entity metadata property keys used by this port, named after
// pocketmine\network\mcpe\protocol\types\entity\EntityMetadataProperties (BedrockProtocol) and
// mapped onto gophertunnel's EntityDataKey* wire values.
const (
	MetadataFlags                               = protocol.EntityDataKeyFlags
	MetadataFlags2                              = protocol.EntityDataKeyFlagsTwo
	MetadataVariant                             = protocol.EntityDataKeyVariant
	MetadataColor                               = protocol.EntityDataKeyColorIndex
	MetadataNametag                             = protocol.EntityDataKeyName
	MetadataOwnerEID                            = protocol.EntityDataKeyOwner
	MetadataTargetEID                           = protocol.EntityDataKeyTarget
	MetadataAir                                 = protocol.EntityDataKeyAirSupply
	MetadataPotionColor                         = protocol.EntityDataKeyEffectColor
	MetadataPotionAmbient                       = protocol.EntityDataKeyEffectAmbience
	MetadataExperienceValue                     = protocol.EntityDataKeyValue
	MetadataFireworkItem                        = protocol.EntityDataKeyDisplayFirework
	MetadataPotionAuxValue                      = protocol.EntityDataKeyAuxValueData
	MetadataLeadHolderEID                       = protocol.EntityDataKeyLeashHolder
	MetadataScale                               = protocol.EntityDataKeyScale
	MetadataMaxAir                              = protocol.EntityDataKeyAirSupplyMax
	MetadataBlockTarget                         = protocol.EntityDataKeyBlockTarget
	MetadataBoundingBoxWidth                    = protocol.EntityDataKeyWidth
	MetadataBoundingBoxHeight                   = protocol.EntityDataKeyHeight
	MetadataFuseLength                          = protocol.EntityDataKeyFuseTime
	MetadataAreaEffectCloudRadius               = protocol.EntityDataKeyDataRadius
	MetadataAreaEffectCloudWaiting              = protocol.EntityDataKeyDataWaiting
	MetadataAlwaysShowNametag                   = protocol.EntityDataKeyAlwaysShowNameTag
	MetadataScoreTag                            = protocol.EntityDataKeyScore
	MetadataAreaEffectCloudDuration             = protocol.EntityDataKeyDataDuration
	MetadataAreaEffectCloudSpawnTime            = protocol.EntityDataKeyDataSpawnTime
	MetadataAreaEffectCloudRadiusPerTick        = protocol.EntityDataKeyDataChangeRate
	MetadataAreaEffectCloudRadiusChangeOnPickup = protocol.EntityDataKeyDataChangeOnPickup
	MetadataAreaEffectCloudPickupCount          = protocol.EntityDataKeyDataPickupCount
	MetadataVisibleMobEffects                   = protocol.EntityDataKeyVisibleMobEffects
	MetadataPlayerFlags                         = protocol.EntityDataKeyPlayerFlags
	MetadataPlayerBedPosition                   = protocol.EntityDataKeyBedPosition
)

// Generic entity flags used by this port, named after EntityMetadataFlags and mapped onto
// gophertunnel's EntityDataFlag* indices (flags >= 64 live in the FLAGS2 property).
const (
	FlagOnFire            = protocol.EntityDataFlagOnFire
	FlagSneaking          = protocol.EntityDataFlagSneaking
	FlagSprinting         = protocol.EntityDataFlagSprinting
	FlagAction            = protocol.EntityDataFlagUsingItem
	FlagInvisible         = protocol.EntityDataFlagInvisible
	FlagIgnited           = protocol.EntityDataFlagIgnited
	FlagBaby              = protocol.EntityDataFlagBaby
	FlagCritical          = protocol.EntityDataFlagCritical
	FlagCanShowNametag    = protocol.EntityDataFlagShowName
	FlagAlwaysShowNametag = protocol.EntityDataFlagAlwaysShowName
	FlagNoAI              = protocol.EntityDataFlagNoAI
	FlagSilent            = protocol.EntityDataFlagSilent
	FlagWallClimbing      = protocol.EntityDataFlagWallClimbing
	FlagCanClimb          = protocol.EntityDataFlagClimb
	FlagGliding           = protocol.EntityDataFlagGliding
	FlagBreathing         = protocol.EntityDataFlagBreathing
	FlagShowBase          = protocol.EntityDataFlagShowBottom
	FlagLinger            = protocol.EntityDataFlagLingering
	FlagHasCollision      = protocol.EntityDataFlagHasCollision
	FlagAffectedByGravity = protocol.EntityDataFlagHasGravity
	FlagEnchanted         = protocol.EntityDataFlagEnchanted
	FlagSwimming          = protocol.EntityDataFlagSwimming
	FlagSleeping          = protocol.EntityDataFlagSleeping
)

// MetadataCollection is a port of
// pocketmine\network\mcpe\protocol\types\entity\EntityMetadataCollection: an entity's network
// properties, tracking which ones changed since they were last sent.
//
// Values are stored as the Go types gophertunnel's EntityMetadata encoder expects (byte, int16,
// int32, int64, float32, string, map[string]any, protocol.BlockPos, mgl32.Vec3).
type MetadataCollection struct {
	properties protocol.EntityMetadata
	dirty      protocol.EntityMetadata
}

func NewMetadataCollection() *MetadataCollection {
	return &MetadataCollection{properties: protocol.EntityMetadata{}, dirty: protocol.EntityMetadata{}}
}

// set is a port of EntityMetadataCollection::set: unchanged values aren't marked dirty unless
// force is true.
func (c *MetadataCollection) set(key uint32, value any, force bool) {
	if existing, ok := c.properties[key]; !force && ok && reflect.DeepEqual(existing, value) {
		return
	}
	c.properties[key] = value
	c.dirty[key] = value
}

func (c *MetadataCollection) SetByte(key uint32, value byte)      { c.set(key, value, false) }
func (c *MetadataCollection) SetShort(key uint32, value int16)    { c.set(key, value, false) }
func (c *MetadataCollection) SetInt(key uint32, value int32)      { c.set(key, value, false) }
func (c *MetadataCollection) SetLong(key uint32, value int64)     { c.set(key, value, false) }
func (c *MetadataCollection) SetFloat(key uint32, value float32)  { c.set(key, value, false) }
func (c *MetadataCollection) SetString(key uint32, value string)  { c.set(key, value, false) }
func (c *MetadataCollection) SetVector3(key uint32, v mgl32.Vec3) { c.set(key, v, false) }

func (c *MetadataCollection) SetCompoundTag(key uint32, value map[string]any) {
	c.set(key, value, false)
}

func (c *MetadataCollection) SetBlockPos(key uint32, value protocol.BlockPos) {
	c.set(key, value, false)
}

// SetGenericFlag is a port of EntityMetadataCollection::setGenericFlag: flags 0-63 live in FLAGS,
// 64+ in FLAGS2.
func (c *MetadataCollection) SetGenericFlag(flagID int, value bool) {
	propertyID := uint32(MetadataFlags)
	if flagID >= 64 {
		propertyID = MetadataFlags2
	}
	realFlagID := uint(flagID % 64)

	var flagSet int64
	if existing, ok := c.properties[propertyID].(int64); ok {
		flagSet = existing
	}

	current := (flagSet >> realFlagID) & 1
	want := int64(0)
	if value {
		want = 1
	}
	if current != want {
		flagSet ^= 1 << realFlagID
		c.SetLong(propertyID, flagSet)
	}
}

// SetPlayerFlag is a port of EntityMetadataCollection::setPlayerFlag.
func (c *MetadataCollection) SetPlayerFlag(flagID int, value bool) {
	var flagSet byte
	if existing, ok := c.properties[MetadataPlayerFlags].(byte); ok {
		flagSet = existing
	}
	current := (flagSet >> uint(flagID)) & 1
	want := byte(0)
	if value {
		want = 1
	}
	if current != want {
		flagSet ^= 1 << uint(flagID)
		c.SetByte(MetadataPlayerFlags, flagSet)
	}
}

// GetAll returns a copy of every property.
func (c *MetadataCollection) GetAll() protocol.EntityMetadata { return copyMetadata(c.properties) }

// GetDirty returns a copy of the properties changed since ClearDirtyProperties was last called.
func (c *MetadataCollection) GetDirty() protocol.EntityMetadata { return copyMetadata(c.dirty) }

// ClearDirtyProperties clears the dirty set (called after the dirty properties have been sent).
func (c *MetadataCollection) ClearDirtyProperties() { c.dirty = protocol.EntityMetadata{} }

func copyMetadata(m protocol.EntityMetadata) protocol.EntityMetadata {
	result := make(protocol.EntityMetadata, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}
