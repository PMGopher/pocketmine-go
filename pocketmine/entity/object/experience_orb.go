package object

import (
	"fmt"
	stdmath "math"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
)

// ExperienceOrb NBT keys and constants, a port of ExperienceOrb's constants.
const (
	TagValuePC = "Value"            //short
	TagValuePE = "experience value" //int (WTF?)
	tagOrbAge  = "Age"              //TAG_Short

	// MaxTargetDistance is the max distance an orb will follow a player across.
	MaxTargetDistance = 8.0

	OrbDefaultDespawnDelay = 6000
	OrbNeverDespawn        = -1
	OrbMaxDespawnDelay     = 32767 + OrbDefaultDespawnDelay //max value storable by mojang NBT :(
)

// OrbSplitSizes mirrors ExperienceOrb::ORB_SPLIT_SIZES: split sizes used for dropping experience
// orbs, indexed biggest to smallest so that we can return as soon as we found the biggest value.
var OrbSplitSizes = []int{2477, 1237, 617, 307, 149, 73, 37, 17, 7, 3, 1}

// GetMaxOrbSize returns the largest size of normal XP orb that will be spawned for the specified
// amount of XP. Used to split XP into multiple orbs when spawning.
func GetMaxOrbSize(amount int) int {
	for _, split := range OrbSplitSizes {
		if amount >= split {
			return split
		}
	}
	return 1
}

// SplitIntoOrbSizes splits the specified amount of XP into an array of acceptable XP orb sizes.
func SplitIntoOrbSizes(amount int) []int {
	var result []int
	for amount > 0 {
		size := GetMaxOrbSize(amount)
		result = append(result, size)
		amount -= size
	}
	return result
}

// xpHuman is the surface of entity.Human experience orbs need (PHP's `instanceof Human`).
type xpHuman interface {
	world.Entity
	GetXpManager() *entity.ExperienceManager
	GetEyeHeight() float64
}

// ExperienceOrb is a port of pocketmine\entity\object\ExperienceOrb.
type ExperienceOrb struct {
	entity.Entity

	lookForTargetTime     int
	targetPlayerRuntimeID *int
	xpValue               int
	despawnDelay          int
}

// NewExperienceOrb is a port of ExperienceOrb::__construct.
func NewExperienceOrb(location entity.Location, xpValue int, tag *nbt.CompoundTag) *ExperienceOrb {
	o := &ExperienceOrb{xpValue: xpValue, despawnDelay: OrbDefaultDespawnDelay}
	o.Construct(o, location, tag)
	return o
}

func (o *ExperienceOrb) GetNetworkTypeID() string { return entity.EntityIDXPOrb }

func (o *ExperienceOrb) GetInitialSizeInfo() entity.EntitySizeInfo {
	return entity.NewEntitySizeInfo(0.25, 0.25)
}

func (o *ExperienceOrb) GetInitialDragMultiplier() float64 { return 0.02 }

func (o *ExperienceOrb) GetInitialGravity() float64 { return 0.04 }

// InitEntity is a port of ExperienceOrb::initEntity.
func (o *ExperienceOrb) InitEntity(tag *nbt.CompoundTag) {
	o.Entity.InitEntity(tag)

	age := int(tag.GetShortOr(tagOrbAge, 0))
	if age == -32768 {
		o.despawnDelay = OrbNeverDespawn
	} else {
		o.despawnDelay = max(0, OrbDefaultDespawnDelay-age)
	}
}

// SaveNBT is a port of ExperienceOrb::saveNBT.
func (o *ExperienceOrb) SaveNBT() *nbt.CompoundTag {
	tag := o.Entity.SaveNBT()
	age := -32768
	if o.despawnDelay != OrbNeverDespawn {
		age = OrbDefaultDespawnDelay - o.despawnDelay
	}
	tag.SetShort(tagOrbAge, nbt.ShortTag(age))

	tag.SetShort(TagValuePC, nbt.ShortTag(o.GetXpValue()))
	tag.SetInt(TagValuePE, nbt.IntTag(o.GetXpValue()))

	return tag
}

func (o *ExperienceOrb) GetDespawnDelay() int { return o.despawnDelay }

// SetDespawnDelay panics outside 0..OrbMaxDespawnDelay (and != OrbNeverDespawn).
func (o *ExperienceOrb) SetDespawnDelay(despawnDelay int) {
	if (despawnDelay < 0 || despawnDelay > OrbMaxDespawnDelay) && despawnDelay != OrbNeverDespawn {
		panic(fmt.Sprintf("Despawn ticker must be in range 0 ... %d or %d, got %d", OrbMaxDespawnDelay, OrbNeverDespawn, despawnDelay))
	}
	o.despawnDelay = despawnDelay
}

func (o *ExperienceOrb) GetXpValue() int { return o.xpValue }

// SetXpValue panics for a non-positive amount (PHP's InvalidArgumentException).
func (o *ExperienceOrb) SetXpValue(amount int) {
	if amount <= 0 {
		panic(fmt.Sprintf("XP amount must be greater than 0, got %d", amount))
	}
	o.xpValue = amount
	o.MarkNetworkPropertiesDirty()
}

func (o *ExperienceOrb) HasTargetPlayer() bool { return o.targetPlayerRuntimeID != nil }

// GetTargetPlayer is a port of ExperienceOrb::getTargetPlayer (nil if none).
func (o *ExperienceOrb) GetTargetPlayer() xpHuman {
	if o.targetPlayerRuntimeID == nil {
		return nil
	}

	e, ok := o.GetWorld().GetEntity(*o.targetPlayerRuntimeID)
	//TODO: HACK! We really shouldn't be keeping disconnected players (and generally flagged-for-despawn entities)
	//in the world's entity table, but changing that is too risky for a hotfix. This workaround will do for now.
	if human, isHuman := e.(xpHuman); ok && isHuman && !human.IsFlaggedForDespawn() {
		return human
	}

	return nil
}

func (o *ExperienceOrb) SetTargetPlayer(player xpHuman) {
	if player == nil {
		o.targetPlayerRuntimeID = nil
		return
	}
	id := player.GetID()
	o.targetPlayerRuntimeID = &id
}

// EntityBaseTick is a port of ExperienceOrb::entityBaseTick: find, follow and get picked up by the
// nearest human.
func (o *ExperienceOrb) EntityBaseTick(tickDiff int) bool {
	hasUpdate := o.Entity.EntityBaseTick(tickDiff)

	o.despawnDelay -= tickDiff
	if o.despawnDelay <= 0 {
		o.FlagForDespawn()
		return true
	}

	pos := o.GetPosition()
	currentTarget := o.GetTargetPlayer()
	if currentTarget != nil && (!currentTarget.IsAlive() || !currentTarget.GetXpManager().CanAttractXpOrbs() || currentTarget.GetPosition().DistanceSquared(pos) > MaxTargetDistance*MaxTargetDistance) {
		currentTarget = nil
	}

	if o.lookForTargetTime >= 20 {
		if currentTarget == nil {
			newTarget := o.GetWorld().GetNearestEntity(pos, MaxTargetDistance, false, func(e world.Entity) bool {
				_, ok := e.(xpHuman)
				return ok
			})

			if human, ok := newTarget.(xpHuman); ok && human.GetXpManager().CanAttractXpOrbs() {
				if p, isPlayer := entity.AsPlayer(newTarget); !isPlayer || !p.IsSpectator() {
					currentTarget = human
				}
			}
		}

		o.lookForTargetTime = 0
	} else {
		o.lookForTargetTime += tickDiff
	}

	o.SetTargetPlayer(currentTarget)

	if currentTarget != nil {
		vector := currentTarget.GetPosition().Add(0, currentTarget.GetEyeHeight()/2, 0).SubtractVector(pos).Divide(MaxTargetDistance)

		distance := vector.LengthSquared()
		if distance < 1 {
			factor := 1 - stdmath.Sqrt(distance)
			o.SetMotionDirect(o.GetMotion().AddVector(vector.Normalize().Multiply(0.2 * factor * factor)))
		}

		if currentTarget.GetXpManager().CanPickupXp() && o.BoundingBox.IntersectsWith(currentTarget.GetBoundingBox(), 0.00001) {
			o.FlagForDespawn()

			currentTarget.GetXpManager().OnPickupXp(o.GetXpValue())
		}
	}

	return hasUpdate
}

// TryChangeMovement is a port of ExperienceOrb::tryChangeMovement.
func (o *ExperienceOrb) TryChangeMovement() {
	pos := o.GetPosition()
	o.CheckObstruction(pos.X, pos.Y, pos.Z)
	o.Entity.TryChangeMovement()
}

func (o *ExperienceOrb) CanBeCollidedWith() bool { return false }

// SyncNetworkData is a port of ExperienceOrb::syncNetworkData.
func (o *ExperienceOrb) SyncNetworkData(properties *entity.MetadataCollection) {
	o.Entity.SyncNetworkData(properties)

	properties.SetInt(entity.MetadataExperienceValue, int32(o.xpValue))
}
