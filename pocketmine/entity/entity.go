// Package entity is a port of pocketmine\entity: the Entity base class, Living, Human, the mobs
// PocketMine-MP ships (Squid, Villager, Zombie), attributes, the hunger/experience managers and
// EntityFactory. Sub-namespaces are their own packages: entity/effect, entity/animation,
// entity/object, entity/projectile and entity/utils.
//
// # Porting conventions for the class hierarchy
//
// PHP's Entity calls many overridable methods on $this from its own method bodies. Go embedding
// doesn't dispatch virtually, so every concrete entity passes itself to Construct and the base
// methods call those hooks through e.self (the same pattern as block.Block). Because subtypes live
// in other packages (entity/object, entity/projectile, player), every PHP protected hook they can
// override (EntityBaseTick, InitEntity, SyncNetworkData, OnHitGround, Move, ...) is an exported
// method here. They're still "protected" in spirit: code outside the entity hierarchy should call
// the public API instead.
//
// PHP's public fields ($onGround, $boundingBox, $fallDistance, $ticksLived, $noDamageTicks,
// $keepMovement, $isCollided*, $size, $lastUpdate) keep getters/setters here, and the protected
// fields subtypes write directly ($motion, $location, $blocksAround, $networkPropertiesDirty) get
// explicitly-named "direct" accessors (SetMotionDirect, SetLocationDirect, ClearBlocksAround,
// MarkNetworkPropertiesDirty) that skip the events/side effects of the public setters, exactly like
// PHP's direct property writes.
//
// Not ported: timings (Timings::getEntityTimings) and the debug log line for a non-positive tick
// difference.
package entity

import (
	"fmt"
	stdmath "math"
	"reflect"
	"sync/atomic"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/binaryutils"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity/animation"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// MotionThreshold mirrors Entity::MOTION_THRESHOLD.
const MotionThreshold = 0.00001

// stepClipMultiplier mirrors Entity::STEP_CLIP_MULTIPLIER.
const stepClipMultiplier = 0.4

var entityCount atomic.Int64

func init() { entityCount.Store(1) }

// NextRuntimeID is a port of Entity::nextRuntimeId: a new runtime entity ID for a new entity.
func NextRuntimeID() int { return int(entityCount.Add(1) - 1) }

// Hooks is the self-dispatch surface: every method Entity's own bodies call on $this that a
// subtype may override (PHP's abstract and protected methods included). It's exported so subtypes
// in other packages can extend it for their own self-dispatch (e.g. projectile.Projectile).
type Hooks interface {
	world.Entity

	GetInitialSizeInfo() EntitySizeInfo
	GetInitialDragMultiplier() float64
	GetInitialGravity() float64
	GetNetworkTypeID() string

	InitEntity(tag *nbt.CompoundTag)
	AddAttributes()
	RecalculateBoundingBox()
	// SendData, SetMotion, BroadcastSound and BroadcastAnimation are overridden by Player, which
	// also sends them to its own client; PHP calls them through $this.
	SendData(targets []world.EntityViewer, data protocol.EntityMetadata)
	SetMotion(motion math.Vector3) bool
	BroadcastSound(s sound.Sound)
	BroadcastAnimation(anim animation.Animation, targets []world.EntityViewer)

	OnDeath()
	OnDeathUpdate(tickDiff int) bool
	EntityBaseTick(tickDiff int) bool
	OnFirstUpdate(currentTick int64)

	IsFireProof() bool
	DoOnFireTick(tickDiff int) bool
	DealFireDamage()

	GetHealth() float64
	SetHealth(amount float64)
	GetMaxHealth() int
	Kill()

	HasMovementUpdate() bool
	TryChangeMovement()
	ApplyDragBeforeGravity() bool
	Move(dx, dy, dz float64)
	CheckGroundState(wantedX, wantedY, wantedZ, dx, dy, dz float64)
	OnHitGround() *float64
	UpdateMovement(teleport bool)
	GetOffsetPosition(v math.Vector3) math.Vector3
	BroadcastMovement(teleport bool)
	BroadcastMotion()
	GetEyeHeight() float64
	SetPositionInWorld(pos math.Vector3, w *world.World) bool

	SendSpawnPacket(player world.EntityViewer)
	SyncNetworkData(properties *MetadataCollection)
	OnDispose()
	DestroyCycles()
}

// Entity is a port of pocketmine\entity\Entity.
type Entity struct {
	self Hooks

	hasSpawned      map[world.EntityViewer]bool
	hasSpawnedOrder []world.EntityViewer

	id int

	networkProperties *MetadataCollection

	lastDamageCause entityevent.DamageSource

	blocksAround      []block.Behavior
	blocksAroundValid bool

	location                        Location
	lastLocation                    Location
	motion                          math.Vector3
	lastMotion                      math.Vector3
	forceMovementUpdate             bool
	checkBlockIntersectionsNextTick bool

	BoundingBox math.AxisAlignedBB
	OnGround    bool

	Size EntitySizeInfo

	health    float64
	maxHealth int

	ySize      float64
	stepHeight float64
	// KeepMovement mirrors Entity::$keepMovement: when set, Move skips collision entirely.
	KeepMovement bool

	FallDistance float64
	TicksLived   int
	LastUpdate   int64
	fireTicks    int

	savedWithChunk bool

	IsCollided             bool
	IsCollidedHorizontally bool
	IsCollidedVertically   bool

	NoDamageTicks int
	justCreated   bool

	attributeMap *AttributeMap

	gravity        float64
	drag           float64
	gravityEnabled bool

	closed        bool
	closeInFlight bool
	needsDespawn  bool

	networkPropertiesDirty bool

	nameTag           string
	nameTagVisible    bool
	alwaysShowNameTag bool
	scoreTag          string
	scale             float64

	canClimb            bool
	canClimbWalls       bool
	noClientPredictions bool
	invisible           bool
	silent              bool

	ownerID  *int
	targetID *int

	constructorCalled bool
}

// Construct is a port of Entity::__construct. Concrete entity constructors set their own fields
// first (PHP subclasses do the same before calling parent::__construct, since the base constructor
// calls overridable hooks like GetInitialSizeInfo/InitEntity), then call Construct with themselves
// as self.
//
// Invalid saved data in tag makes Construct panic with a *data.SavedDataLoadingError (PHP throws
// SavedDataLoadingException); EntityFactory.CreateFromData recovers that into an error. A nil tag
// means a freshly created entity.
func (e *Entity) Construct(self Hooks, location Location, tag *nbt.CompoundTag) {
	if e.constructorCalled {
		panic("Attempted to call constructor for an Entity multiple times")
	}
	e.constructorCalled = true
	checkLocationNotInfOrNaN(location)

	e.self = self
	e.hasSpawned = map[world.EntityViewer]bool{}
	e.health = 20
	e.maxHealth = 20
	e.savedWithChunk = true
	e.justCreated = true
	e.gravityEnabled = true
	e.nameTagVisible = true
	e.scale = 1.0
	e.checkBlockIntersectionsNextTick = true

	e.Size = self.GetInitialSizeInfo()
	e.drag = self.GetInitialDragMultiplier()
	e.gravity = self.GetInitialGravity()

	e.id = NextRuntimeID()

	e.location = location.AsLocation()

	e.BoundingBox = math.AxisAlignedBB{}
	self.RecalculateBoundingBox()

	if tag != nil {
		motion, err := ParseVec3(tag, TagMotion, true)
		if err != nil {
			panic(err)
		}
		e.motion = motion
	} else {
		e.motion = math.Vector3Zero()
	}

	e.resetLastMovements()

	e.networkProperties = NewMetadataCollection()

	e.attributeMap = NewAttributeMap()
	self.AddAttributes()

	if tag == nil {
		tag = nbt.NewCompoundTag()
	}
	self.InitEntity(tag)

	if _, neverSaved := self.(NeverSavedWithChunkEntity); !neverSaved && !GetEntityFactory().IsRegistered(reflect.TypeOf(self)) {
		//canSaveWithChunk is mutable, so that means it could be toggled after adding the entity and
		//cause a crash later on. Better we just force all entities to have a save ID, even if it
		//might not be needed.
		panic(fmt.Sprintf("Entity %T is not registered for a save ID in EntityFactory", self))
	}
	e.GetWorld().AddEntity(self)

	e.LastUpdate = e.GetWorld().GetCurrentTick()

	e.ScheduleUpdate()
}

func checkLocationNotInfOrNaN(l Location) {
	for _, v := range []float64{l.X, l.Y, l.Z, l.Yaw, l.Pitch} {
		checkFloatNotInfOrNaN("location", v)
	}
}

func checkFloatNotInfOrNaN(name string, v float64) {
	if stdmath.IsNaN(v) || stdmath.IsInf(v, 0) {
		panic(fmt.Sprintf("%s cannot be NaN or Inf", name))
	}
}

// IsJustCreated reports PHP's protected $justCreated: true until the entity's first base tick.
func (e *Entity) IsJustCreated() bool { return e.justCreated }

// Self returns the concrete entity embedding this Entity (PHP's $this).
func (e *Entity) Self() world.Entity { return e.self }

// GetInitialDragMultiplier/GetInitialGravity have no base implementation in PHP (they're
// abstract); subtypes must provide them. The same goes for GetInitialSizeInfo and
// GetNetworkTypeID.

func (e *Entity) GetNameTag() string { return e.nameTag }

func (e *Entity) IsNameTagVisible() bool { return e.nameTagVisible }

func (e *Entity) IsNameTagAlwaysVisible() bool { return e.alwaysShowNameTag }

// CanBeRenamed returns whether players can rename this entity using a name tag. Note that plugins
// can still name entities using SetNameTag().
func (e *Entity) CanBeRenamed() bool { return false }

func (e *Entity) SetNameTag(name string) {
	e.nameTag = name
	e.networkPropertiesDirty = true
}

func (e *Entity) SetNameTagVisible(value bool) {
	e.nameTagVisible = value
	e.networkPropertiesDirty = true
}

func (e *Entity) SetNameTagAlwaysVisible(value bool) {
	e.alwaysShowNameTag = value
	e.networkPropertiesDirty = true
}

func (e *Entity) GetScoreTag() string { return e.scoreTag }

func (e *Entity) SetScoreTag(score string) {
	e.scoreTag = score
	e.networkPropertiesDirty = true
}

func (e *Entity) GetScale() float64 { return e.scale }

// SetScale is a port of Entity::setScale (panicking on a non-positive scale).
func (e *Entity) SetScale(value float64) {
	if value <= 0 {
		panic("Scale must be greater than 0")
	}
	e.scale = value
	e.SetSize(e.self.GetInitialSizeInfo().Scale(value))
}

func (e *Entity) GetBoundingBox() math.AxisAlignedBB { return e.BoundingBox }

// RecalculateBoundingBox is a port of Entity::recalculateBoundingBox.
func (e *Entity) RecalculateBoundingBox() {
	halfWidth := e.Size.GetWidth() / 2
	e.BoundingBox = math.AxisAlignedBB{
		MinX: e.location.X - halfWidth,
		MinY: e.location.Y + e.ySize,
		MinZ: e.location.Z - halfWidth,
		MaxX: e.location.X + halfWidth,
		MaxY: e.location.Y + e.Size.GetHeight() + e.ySize,
		MaxZ: e.location.Z + halfWidth,
	}
}

func (e *Entity) GetSize() EntitySizeInfo { return e.Size }

// SetSize is a port of Entity::setSize (protected in PHP).
func (e *Entity) SetSize(size EntitySizeInfo) {
	e.Size = size
	e.self.RecalculateBoundingBox()
	e.networkPropertiesDirty = true
}

// HasNoClientPredictions returns whether clients may predict this entity's behaviour and movement.
func (e *Entity) HasNoClientPredictions() bool { return e.noClientPredictions }

// SetNoClientPredictions is a port of Entity::setNoClientPredictions.
func (e *Entity) SetNoClientPredictions(value bool) {
	e.noClientPredictions = value
	e.networkPropertiesDirty = true
}

func (e *Entity) IsInvisible() bool { return e.invisible }

func (e *Entity) SetInvisible(value bool) {
	e.invisible = value
	e.networkPropertiesDirty = true
}

func (e *Entity) IsSilent() bool { return e.silent }

func (e *Entity) SetSilent(value bool) {
	e.silent = value
	e.networkPropertiesDirty = true
}

// CanClimb returns whether the entity is able to climb blocks such as ladders or vines.
func (e *Entity) CanClimb() bool { return e.canClimb }

func (e *Entity) SetCanClimb(value bool) {
	e.canClimb = value
	e.networkPropertiesDirty = true
}

// CanClimbWalls returns whether this entity is climbing a block.
func (e *Entity) CanClimbWalls() bool { return e.canClimbWalls }

func (e *Entity) SetCanClimbWalls(value bool) {
	e.canClimbWalls = value
	e.networkPropertiesDirty = true
}

// GetOwningEntityID returns the entity ID of the owning entity, or false if it has no owner.
func (e *Entity) GetOwningEntityID() (int, bool) {
	if e.ownerID == nil {
		return 0, false
	}
	return *e.ownerID, true
}

// GetOwningEntity is a port of Entity::getOwningEntity: the owning entity, or nil if there's none
// or it no longer exists. PHP searches every world through WorldManager::findEntity; this port has
// no Server, so the owner is looked up in this entity's own world (an owner in another world
// reads as gone).
func (e *Entity) GetOwningEntity() world.Entity {
	if e.ownerID == nil {
		return nil
	}
	return e.findEntity(*e.ownerID)
}

func (e *Entity) findEntity(id int) world.Entity {
	if !e.location.IsValid() {
		return nil
	}
	if found, ok := e.location.World.GetEntity(id); ok {
		return found
	}
	return nil
}

// SetOwningEntity is a port of Entity::setOwningEntity. nil removes the owner; a closed owner
// panics (PHP's InvalidArgumentException).
func (e *Entity) SetOwningEntity(owner world.Entity) {
	if owner == nil {
		e.ownerID = nil
	} else if owner.IsClosed() {
		panic("Supplied owning entity is garbage and cannot be used")
	} else {
		id := owner.GetID()
		e.ownerID = &id
	}
	e.networkPropertiesDirty = true
}

// GetTargetEntityID returns the entity ID of the entity's target, or false if it has none.
func (e *Entity) GetTargetEntityID() (int, bool) {
	if e.targetID == nil {
		return 0, false
	}
	return *e.targetID, true
}

// GetTargetEntity returns the entity's target entity, or nil if not found (see GetOwningEntity for
// the lookup scope).
func (e *Entity) GetTargetEntity() world.Entity {
	if e.targetID == nil {
		return nil
	}
	return e.findEntity(*e.targetID)
}

// SetTargetEntity is a port of Entity::setTargetEntity.
func (e *Entity) SetTargetEntity(target world.Entity) {
	if target == nil {
		e.targetID = nil
	} else if target.IsClosed() {
		panic("Supplied target entity is garbage and cannot be used")
	} else {
		id := target.GetID()
		e.targetID = &id
	}
	e.networkPropertiesDirty = true
}

// CanSaveWithChunk returns whether this entity will be saved when its chunk is unloaded.
func (e *Entity) CanSaveWithChunk() bool { return e.savedWithChunk }

// SetCanSaveWithChunk sets whether this entity will be saved when its chunk is unloaded.
func (e *Entity) SetCanSaveWithChunk(value bool) { e.savedWithChunk = value }

func doubleList(values ...float64) *nbt.ListTag {
	tags := make([]nbt.Tag, len(values))
	for i, v := range values {
		tags[i] = nbt.DoubleTag(v)
	}
	list, err := nbt.NewListTag(tags, nbt.TagDouble)
	if err != nil {
		panic(err)
	}
	return list
}

func floatList(values ...float64) *nbt.ListTag {
	tags := make([]nbt.Tag, len(values))
	for i, v := range values {
		tags[i] = nbt.FloatTag(v)
	}
	list, err := nbt.NewListTag(tags, nbt.TagFloat)
	if err != nil {
		panic(err)
	}
	return list
}

// SaveNBT is a port of Entity::saveNBT.
func (e *Entity) SaveNBT() *nbt.CompoundTag {
	tag := nbt.NewCompoundTag().
		SetTag(TagPos, doubleList(e.location.X, e.location.Y, e.location.Z)).
		SetTag(TagMotion, doubleList(e.motion.X, e.motion.Y, e.motion.Z)).
		SetTag(TagRotation, floatList(e.location.Yaw, e.location.Pitch))

	if _, neverSaved := e.self.(NeverSavedWithChunkEntity); !neverSaved {
		GetEntityFactory().InjectSaveID(reflect.TypeOf(e.self), tag)

		if e.GetNameTag() != "" {
			tag.SetString(tagCustomName, nbt.StringTag(e.GetNameTag()))
			tag.SetByte(tagCustomNameVisible, boolByte(e.IsNameTagVisible()))
		}
	}

	tag.SetFloat(tagFallDistance, nbt.FloatTag(e.FallDistance))
	tag.SetShort(tagFire, nbt.ShortTag(e.fireTicks))
	tag.SetByte(tagOnGround, boolByte(e.OnGround))

	tag.SetLong(pocketmine.TagWorldDataVersion, pocketmine.WorldDataVersion)

	return tag
}

func boolByte(v bool) nbt.ByteTag {
	if v {
		return 1
	}
	return 0
}

// InitEntity is a port of Entity::initEntity.
func (e *Entity) InitEntity(tag *nbt.CompoundTag) {
	e.fireTicks = int(tag.GetShortOr(tagFire, 0))

	e.OnGround = tag.GetByteOr(tagOnGround, 0) != 0

	e.FallDistance = float64(tag.GetFloatOr(tagFallDistance, 0.0))

	if customNameTag, ok := tag.GetTag(tagCustomName); ok {
		if name, ok := customNameTag.(nbt.StringTag); ok {
			e.SetNameTag(string(name))

			if visibleTag, ok := tag.GetTag(tagCustomNameVisible); ok {
				if s, ok := visibleTag.(nbt.StringTag); ok {
					//Older versions incorrectly saved this as a string (see 890f72dbf23a77f294169b79590770470041adc4)
					e.SetNameTagVisible(string(s) != "")
				} else {
					e.SetNameTagVisible(tag.GetByteOr(tagCustomNameVisible, 1) != 0)
				}
			} else {
				e.SetNameTagVisible(true)
			}
		}
	}
}

// AddAttributes is Entity::addAttributes' default: no attributes.
func (e *Entity) AddAttributes() {}

// Attack is a port of Entity::attack.
func (e *Entity) Attack(source entityevent.DamageSource) {
	if e.self.IsFireProof() {
		switch source.GetCause() {
		case entityevent.CauseFire, entityevent.CauseFireTick, entityevent.CauseLava:
			source.Cancel()
		}
	}
	source.Call()
	if source.IsCancelled() {
		return
	}

	e.SetLastDamageCause(source)

	e.self.SetHealth(e.self.GetHealth() - source.GetFinalDamage())
}

// Heal is a port of Entity::heal.
func (e *Entity) Heal(source *entityevent.EntityRegainHealthEvent) {
	source.Call()
	if source.IsCancelled() {
		return
	}

	e.self.SetHealth(e.self.GetHealth() + source.GetAmount())
}

// Kill is a port of Entity::kill.
func (e *Entity) Kill() {
	if e.IsAlive() {
		e.health = 0
		e.self.OnDeath()
		e.ScheduleUpdate()
	}
}

// OnDeath is Entity::onDeath's default: nothing happens. Override this to do actions on death.
func (e *Entity) OnDeath() {}

// OnDeathUpdate is called to tick entities while dead. Returns whether the entity should be flagged
// for despawn yet.
func (e *Entity) OnDeathUpdate(tickDiff int) bool { return true }

func (e *Entity) IsAlive() bool { return e.health > 0 }

func (e *Entity) GetHealth() float64 { return e.health }

// SetHealth is a port of Entity::setHealth. This won't send any update to the players.
func (e *Entity) SetHealth(amount float64) {
	if amount == e.health {
		return
	}

	if amount <= 0 {
		if e.IsAlive() {
			if !e.justCreated {
				e.self.Kill()
			} else {
				e.health = 0
			}
		}
	} else if amount <= float64(e.self.GetMaxHealth()) || amount < e.health {
		e.health = amount
	} else {
		e.health = float64(e.self.GetMaxHealth())
	}
}

func (e *Entity) GetMaxHealth() int { return e.maxHealth }

func (e *Entity) SetMaxHealth(amount int) { e.maxHealth = amount }

func (e *Entity) SetLastDamageCause(source entityevent.DamageSource) { e.lastDamageCause = source }

// GetLastDamageCause returns the last damage event received, or nil.
func (e *Entity) GetLastDamageCause() entityevent.DamageSource { return e.lastDamageCause }

func (e *Entity) GetAttributeMap() *AttributeMap { return e.attributeMap }

func (e *Entity) GetNetworkProperties() *MetadataCollection { return e.networkProperties }

// EntityBaseTick is a port of Entity::entityBaseTick.
func (e *Entity) EntityBaseTick(tickDiff int) bool {
	//TODO: check vehicles

	if e.justCreated {
		e.justCreated = false
		if !e.IsAlive() {
			e.self.Kill()
		}
	}

	changedProperties := e.GetDirtyNetworkData()
	if len(changedProperties) > 0 {
		e.self.SendData(nil, changedProperties) // $this->sendData(null, $changedProperties)
		e.networkProperties.ClearDirtyProperties()
	}

	hasUpdate := false

	if e.checkBlockIntersectionsNextTick {
		e.CheckBlockIntersections()
	}
	e.checkBlockIntersectionsNextTick = true

	if e.location.Y <= world.YMin-16 && e.IsAlive() {
		ev := entityevent.NewEntityDamageEvent(e.self, entityevent.CauseVoid, 10, nil)
		e.self.Attack(ev)
		hasUpdate = true
	}

	if e.IsOnFire() && e.self.DoOnFireTick(tickDiff) {
		hasUpdate = true
	}

	if e.NoDamageTicks > 0 {
		e.NoDamageTicks -= tickDiff
		if e.NoDamageTicks < 0 {
			e.NoDamageTicks = 0
		}
	}

	e.TicksLived += tickDiff

	return hasUpdate
}

func (e *Entity) IsOnFire() bool { return e.fireTicks > 0 }

// SetOnFire is a port of Entity::setOnFire.
func (e *Entity) SetOnFire(seconds int) {
	ticks := seconds * 20
	if ticks > e.GetFireTicks() {
		e.SetFireTicks(ticks)
	}
	e.networkPropertiesDirty = true
}

func (e *Entity) GetFireTicks() int { return e.fireTicks }

// SetFireTicks is a port of Entity::setFireTicks. Panics on a negative value (PHP's
// InvalidArgumentException); values above the save format's limit are truncated.
func (e *Entity) SetFireTicks(fireTicks int) {
	if fireTicks < 0 {
		panic("Fire ticks cannot be negative")
	}

	//Since the max value is not externally obvious or intuitive, many plugins use this without being aware that
	//reasonably large values are not accepted. We even have such usages within PM itself. It doesn't make sense
	//to force all those calls to be aware of this limitation, as it's not a functional limit but a limitation of
	//the Mojang save format. Truncating this to the max acceptable value is the next best thing we can do.
	fireTicks = min(fireTicks, binaryutils.Int16Max)

	if !e.self.IsFireProof() {
		e.fireTicks = fireTicks
		e.networkPropertiesDirty = true
	}
}

// Extinguish is Entity::extinguish() with its default cause (CAUSE_CUSTOM).
func (e *Entity) Extinguish() { e.ExtinguishWithCause(entityevent.ExtinguishCauseCustom) }

// ExtinguishWithCause is a port of Entity::extinguish.
func (e *Entity) ExtinguishWithCause(cause int) {
	ev := entityevent.NewEntityExtinguishEvent(e.self, cause)
	ev.Call()

	e.fireTicks = 0
	e.networkPropertiesDirty = true
}

func (e *Entity) IsFireProof() bool { return false }

// DoOnFireTick is a port of Entity::doOnFireTick.
func (e *Entity) DoOnFireTick(tickDiff int) bool {
	if e.self.IsFireProof() && e.IsOnFire() {
		e.ExtinguishWithCause(entityevent.ExtinguishCauseFireProof)
		return false
	}

	e.fireTicks -= tickDiff

	if e.fireTicks%20 == 0 || tickDiff > 20 {
		e.self.DealFireDamage()
	}

	if !e.IsOnFire() {
		e.ExtinguishWithCause(entityevent.ExtinguishCauseTicking)
	} else {
		return true
	}

	return false
}

// DealFireDamage is called to deal damage to entities when they are on fire.
func (e *Entity) DealFireDamage() {
	ev := entityevent.NewEntityDamageEvent(e.self, entityevent.CauseFireTick, 1, nil)
	e.self.Attack(ev)
}

// CanCollideWith is a port of Entity::canCollideWith.
func (e *Entity) CanCollideWith(other world.Entity) bool {
	return !e.justCreated && other != world.Entity(e.self)
}

func (e *Entity) CanBeCollidedWith() bool { return e.IsAlive() }

// UpdateMovement is a port of Entity::updateMovement.
func (e *Entity) UpdateMovement(teleport bool) {
	diffPosition := e.location.DistanceSquared(e.lastLocation.Vector3)
	diffRotation := (e.location.Yaw-e.lastLocation.Yaw)*(e.location.Yaw-e.lastLocation.Yaw) +
		(e.location.Pitch-e.lastLocation.Pitch)*(e.location.Pitch-e.lastLocation.Pitch)

	diffMotion := e.motion.SubtractVector(e.lastMotion).LengthSquared()

	still := e.motion.LengthSquared() == 0.0
	wasStill := e.lastMotion.LengthSquared() == 0.0
	if wasStill != still {
		//TODO: hack for client-side AI interference: prevent client sided movement when motion is 0
		e.SetNoClientPredictions(still)
	}

	if teleport || diffPosition > 0.0001 || diffRotation > 1.0 || (!wasStill && still) {
		e.lastLocation = e.location.AsLocation()

		e.self.BroadcastMovement(teleport)
	}

	if diffMotion > 0.0025 || wasStill != still { //0.05 ** 2
		e.lastMotion = e.motion

		e.self.BroadcastMotion()
	}
}

// GetOffsetPosition is Entity::getOffsetPosition's default: no offset.
func (e *Entity) GetOffsetPosition(v math.Vector3) math.Vector3 { return v }

func vec32(v math.Vector3) mgl32.Vec3 {
	return mgl32.Vec3{float32(v.X), float32(v.Y), float32(v.Z)}
}

// BroadcastMovement is a port of Entity::broadcastMovement.
func (e *Entity) BroadcastMovement(teleport bool) {
	var flags byte
	//TODO: We should be setting FLAG_TELEPORT here to disable client-side movement interpolation, but it
	//breaks player teleporting (observers see the player rubberband back to the pre-teleport position while
	//the teleported player sees themselves at the correct position), and does nothing whatsoever for
	//non-player entities (movement is still interpolated). Both of these are client bugs.
	//See https://github.com/pmmp/PocketMine-MP/issues/4394
	if e.OnGround {
		flags |= packet.MoveFlagOnGround
	}
	BroadcastPackets(e.hasSpawnedOrder, &packet.MoveActorAbsolute{
		EntityRuntimeID: uint64(e.id),
		Flags:           flags,
		Position:        vec32(e.self.GetOffsetPosition(e.location.Vector3)),
		Rotation:        mgl32.Vec3{float32(e.location.Pitch), float32(e.location.Yaw), float32(e.location.Yaw)},
	})
}

// BroadcastMotion is a port of Entity::broadcastMotion.
func (e *Entity) BroadcastMotion() {
	BroadcastPackets(e.hasSpawnedOrder, &packet.SetActorMotion{
		EntityRuntimeID: uint64(e.id),
		Velocity:        vec32(e.GetMotion()),
		Tick:            0,
	})
}

func (e *Entity) GetGravity() float64 { return e.gravity }

func (e *Entity) SetGravity(gravity float64) {
	checkFloatNotInfOrNaN("gravity", gravity)
	e.gravity = gravity
}

func (e *Entity) HasGravity() bool { return e.gravityEnabled }

func (e *Entity) SetHasGravity(v bool) { e.gravityEnabled = v }

// ApplyDragBeforeGravity is Entity::applyDragBeforeGravity's default.
func (e *Entity) ApplyDragBeforeGravity() bool { return false }

// TryChangeMovement is a port of Entity::tryChangeMovement: gravity, drag and ground friction.
func (e *Entity) TryChangeMovement() {
	friction := 1 - e.drag

	mY := e.motion.Y

	if e.self.ApplyDragBeforeGravity() {
		mY *= friction
	}

	if e.gravityEnabled {
		mY -= e.gravity
	}

	if !e.self.ApplyDragBeforeGravity() {
		mY *= friction
	}

	if e.OnGround {
		below := e.GetWorld().GetBlockAtIfLoaded(int(stdmath.Floor(e.location.X)), int(stdmath.Floor(e.location.Y-1)), int(stdmath.Floor(e.location.Z)))
		friction *= below.GetFrictionFactor()
	}

	e.motion = math.NewVector3(e.motion.X*friction, mY, e.motion.Z*friction)
}

// CheckObstruction is a port of Entity::checkObstruction: pushes the entity out of a solid block
// it's stuck inside, towards the nearest open side.
func (e *Entity) CheckObstruction(x, y, z float64) bool {
	w := e.GetWorld()
	if len(w.GetBlockCollisionBoxes(e.BoundingBox)) == 0 {
		return false
	}

	floorX := int(stdmath.Floor(x))
	floorY := int(stdmath.Floor(y))
	floorZ := int(stdmath.Floor(z))

	diffX := x - float64(floorX)
	diffY := y - float64(floorY)
	diffZ := z - float64(floorZ)

	if w.GetBlockAtIfLoaded(floorX, floorY, floorZ).IsSolid() {
		westNonSolid := !w.GetBlockAtIfLoaded(floorX-1, floorY, floorZ).IsSolid()
		eastNonSolid := !w.GetBlockAtIfLoaded(floorX+1, floorY, floorZ).IsSolid()
		downNonSolid := !w.GetBlockAtIfLoaded(floorX, floorY-1, floorZ).IsSolid()
		upNonSolid := !w.GetBlockAtIfLoaded(floorX, floorY+1, floorZ).IsSolid()
		northNonSolid := !w.GetBlockAtIfLoaded(floorX, floorY, floorZ-1).IsSolid()
		southNonSolid := !w.GetBlockAtIfLoaded(floorX, floorY, floorZ+1).IsSolid()

		direction := math.Facing(-1)
		limit := 9999.0

		if westNonSolid {
			limit = diffX
			direction = math.West
		}

		if eastNonSolid && 1-diffX < limit {
			limit = 1 - diffX
			direction = math.East
		}

		if downNonSolid && diffY < limit {
			limit = diffY
			direction = math.Down
		}

		if upNonSolid && 1-diffY < limit {
			limit = 1 - diffY
			direction = math.Up
		}

		if northNonSolid && diffZ < limit {
			limit = diffZ
			direction = math.North
		}

		if southNonSolid && 1-diffZ < limit {
			direction = math.South
		}

		if direction == -1 {
			return false
		}

		force := utils.GetRandomFloat()*0.2 + 0.1

		switch direction {
		case math.West:
			e.motion.X = -force
		case math.East:
			e.motion.X = force
		case math.Down:
			e.motion.Y = -force
		case math.Up:
			e.motion.Y = force
		case math.North:
			e.motion.Z = -force
		case math.South:
			e.motion.Z = force
		}
		return true
	}

	return false
}

// GetHorizontalFacing is a port of Entity::getHorizontalFacing.
func (e *Entity) GetHorizontalFacing() math.Facing {
	angle := stdmath.Mod(e.location.Yaw, 360)
	if angle < 0 {
		angle += 360.0
	}

	if (0 <= angle && angle < 45) || (315 <= angle && angle < 360) {
		return math.South
	}
	if 45 <= angle && angle < 135 {
		return math.West
	}
	if 135 <= angle && angle < 225 {
		return math.North
	}

	return math.East
}

func deg2rad(d float64) float64 { return d * stdmath.Pi / 180 }

// GetDirectionVector is a port of Entity::getDirectionVector.
func (e *Entity) GetDirectionVector() math.Vector3 {
	y := -stdmath.Sin(deg2rad(e.location.Pitch))
	xz := stdmath.Cos(deg2rad(e.location.Pitch))
	x := -xz * stdmath.Sin(deg2rad(e.location.Yaw))
	z := xz * stdmath.Cos(deg2rad(e.location.Yaw))

	return math.NewVector3(x, y, z).Normalize()
}

// GetDirectionPlane is a port of Entity::getDirectionPlane.
func (e *Entity) GetDirectionPlane() math.Vector2 {
	return math.NewVector2(-stdmath.Cos(deg2rad(e.location.Yaw)-stdmath.Pi/2), -stdmath.Sin(deg2rad(e.location.Yaw)-stdmath.Pi/2)).Normalize()
}

// OnFirstUpdate is called from OnUpdate on the first tick of a new entity, before any movement
// processing or main ticking logic. Use this to fire any events related to spawning the entity.
func (e *Entity) OnFirstUpdate(currentTick int64) {
	entityevent.NewEntitySpawnEvent(e.self).Call()
}

// OnUpdate is a port of Entity::onUpdate.
func (e *Entity) OnUpdate(currentTick int64) bool {
	if e.closed {
		return false
	}

	tickDiff := int(currentTick - e.LastUpdate)
	if tickDiff <= 0 {
		return true
	}

	e.LastUpdate = currentTick

	if e.justCreated {
		e.self.OnFirstUpdate(currentTick)
	}

	if !e.IsAlive() {
		if e.self.OnDeathUpdate(tickDiff) {
			e.FlagForDespawn()
		}

		return true
	}

	if e.self.HasMovementUpdate() {
		e.self.TryChangeMovement()

		if stdmath.Abs(e.motion.X) <= MotionThreshold {
			e.motion.X = 0
		}
		if stdmath.Abs(e.motion.Y) <= MotionThreshold {
			e.motion.Y = 0
		}
		if stdmath.Abs(e.motion.Z) <= MotionThreshold {
			e.motion.Z = 0
		}

		if e.motion.X != 0 || e.motion.Y != 0 || e.motion.Z != 0 {
			e.self.Move(e.motion.X, e.motion.Y, e.motion.Z)
		}

		e.forceMovementUpdate = false
	}

	e.self.UpdateMovement(false)

	hasUpdate := e.self.EntityBaseTick(tickDiff)

	return hasUpdate || e.self.HasMovementUpdate()
}

// ScheduleUpdate is a port of Entity::scheduleUpdate: the entity will be ticked on the next world
// tick. Panics on a closed entity (PHP's LogicException).
func (e *Entity) ScheduleUpdate() {
	if e.closed {
		panic(fmt.Sprintf("Cannot schedule update on garbage entity %T", e.self))
	}
	e.GetWorld().ScheduleEntityUpdate(e.self)
}

// OnNearbyBlockChange is a port of Entity::onNearbyBlockChange.
func (e *Entity) OnNearbyBlockChange() {
	e.SetForceMovementUpdate(true)
	e.ScheduleUpdate()
}

// OnRandomUpdate is called when a random update is performed on the chunk the entity is in.
func (e *Entity) OnRandomUpdate() { e.ScheduleUpdate() }

// SetForceMovementUpdate flags the entity as needing a movement update on the next tick, even if
// its motion is zero. Used to trigger movement updates when blocks change near entities.
func (e *Entity) SetForceMovementUpdate(value bool) {
	e.forceMovementUpdate = value
	e.ClearBlocksAround()
}

// HasMovementUpdate returns whether the entity needs a movement update on the next tick.
func (e *Entity) HasMovementUpdate() bool {
	return e.forceMovementUpdate || e.motion.X != 0 || e.motion.Y != 0 || e.motion.Z != 0 || !e.OnGround
}

func (e *Entity) GetFallDistance() float64 { return e.FallDistance }

func (e *Entity) SetFallDistance(fallDistance float64) { e.FallDistance = fallDistance }

func (e *Entity) ResetFallDistance() { e.FallDistance = 0.0 }

// UpdateFallState is a port of Entity::updateFallState: returns the new vertical velocity from
// OnHitGround when landing, if it provides one.
func (e *Entity) UpdateFallState(distanceThisTick float64, onGround bool) *float64 {
	if distanceThisTick < e.FallDistance {
		//we've fallen some distance (distanceThisTick is negative)
		//or we ascended back towards where fall distance was measured from initially (distanceThisTick is positive but less than existing fallDistance)
		e.FallDistance -= distanceThisTick
	} else {
		//we ascended past the apex where fall distance was originally being measured from
		//reset it so it will be measured starting from the new, higher position
		e.FallDistance = 0
	}
	if onGround && e.FallDistance > 0 {
		newVerticalVelocity := e.self.OnHitGround()
		e.ResetFallDistance()
		return newVerticalVelocity
	}
	return nil
}

// OnHitGround is called when a falling entity hits the ground.
func (e *Entity) OnHitGround() *float64 { return nil }

func (e *Entity) GetEyeHeight() float64 { return e.Size.GetEyeHeight() }

func (e *Entity) GetEyePos() math.Vector3 {
	return math.NewVector3(e.location.X, e.location.Y+e.self.GetEyeHeight(), e.location.Z)
}

// OnCollideWithPlayer is called when a player collides with this entity (Player::checkNearEntities).
func (e *Entity) OnCollideWithPlayer(player Player) {}

// OnInteract is called when interacted or tapped by a Player. Returns whether something happened
// as a result of the interaction.
func (e *Entity) OnInteract(player Player, clickPos math.Vector3) bool { return false }

// fluidHeightProvider is the surface IsUnderwater needs from block.Water (promoted from Liquid).
type fluidHeightProvider interface {
	GetFluidHeightPercent() float64
}

// IsUnderwater is a port of Entity::isUnderwater.
func (e *Entity) IsUnderwater() bool {
	y := e.location.Y + e.self.GetEyeHeight()
	blockY := int(stdmath.Floor(y))
	blk := e.GetWorld().GetBlockAtIfLoaded(int(stdmath.Floor(e.location.X)), blockY, int(stdmath.Floor(e.location.Z)))

	if water, ok := blk.(*block.Water); ok {
		f := float64(blockY+1) - (fluidHeightProvider(water).GetFluidHeightPercent() - 0.1111111)
		return y < f
	}

	return false
}

// bbCollider is the promoted-from-*block.Block surface IsInsideOfSolid/projectiles need.
type bbCollider interface {
	CollidesWithBB(bb math.AxisAlignedBB) bool
}

// IsInsideOfSolid is a port of Entity::isInsideOfSolid.
func (e *Entity) IsInsideOfSolid() bool {
	blk := e.GetWorld().GetBlockAtIfLoaded(int(stdmath.Floor(e.location.X)), int(stdmath.Floor(e.location.Y+e.self.GetEyeHeight())), int(stdmath.Floor(e.location.Z)))

	collider, ok := blk.(bbCollider)
	return blk.IsSolid() && !blk.IsTransparent() && ok && collider.CollidesWithBB(e.GetBoundingBox())
}

// Move is a port of Entity::move: moves the entity by the given amount, colliding with blocks and
// stepping up ledges up to its step height.
func (e *Entity) Move(dx, dy, dz float64) {
	e.ClearBlocksAround()

	wantedX := dx
	wantedY := dy
	wantedZ := dz

	if e.KeepMovement {
		e.BoundingBox.Offset(dx, dy, dz)
	} else {
		e.ySize *= stepClipMultiplier

		moveBB := e.BoundingBox

		w := e.GetWorld()
		list := w.GetBlockCollisionBoxes(moveBB.AddCoord(dx, dy, dz))

		for _, bb := range list {
			dy = bb.CalculateYOffset(moveBB, dy)
		}

		moveBB.Offset(0, dy, 0)

		fallingFlag := e.OnGround || (dy != wantedY && wantedY < 0)

		for _, bb := range list {
			dx = bb.CalculateXOffset(moveBB, dx)
		}

		moveBB.Offset(dx, 0, 0)

		for _, bb := range list {
			dz = bb.CalculateZOffset(moveBB, dz)
		}

		moveBB.Offset(0, 0, dz)

		stepHeight := e.GetStepHeight()

		if stepHeight > 0 && fallingFlag && (wantedX != dx || wantedZ != dz) {
			cx := dx
			cy := dy
			cz := dz
			dx = wantedX
			dy = stepHeight
			dz = wantedZ

			stepBB := e.BoundingBox

			list = w.GetBlockCollisionBoxes(stepBB.AddCoord(dx, dy, dz))
			for _, bb := range list {
				dy = bb.CalculateYOffset(stepBB, dy)
			}

			stepBB.Offset(0, dy, 0)

			for _, bb := range list {
				dx = bb.CalculateXOffset(stepBB, dx)
			}

			stepBB.Offset(dx, 0, 0)

			for _, bb := range list {
				dz = bb.CalculateZOffset(stepBB, dz)
			}

			stepBB.Offset(0, 0, dz)

			reverseDY := -dy
			for _, bb := range list {
				reverseDY = bb.CalculateYOffset(stepBB, reverseDY)
			}
			dy += reverseDY
			stepBB.Offset(0, reverseDY, 0)

			if cx*cx+cz*cz >= dx*dx+dz*dz {
				dx = cx
				dy = cy
				dz = cz
			} else {
				moveBB = stepBB
				e.ySize += dy
			}
		}

		e.BoundingBox = moveBB
	}

	e.location = NewLocation(
		(e.BoundingBox.MinX+e.BoundingBox.MaxX)/2,
		e.BoundingBox.MinY-e.ySize,
		(e.BoundingBox.MinZ+e.BoundingBox.MaxZ)/2,
		e.location.World,
		e.location.Yaw,
		e.location.Pitch,
	)

	e.GetWorld().OnEntityMoved(e.self)
	e.CheckBlockIntersections()
	e.self.CheckGroundState(wantedX, wantedY, wantedZ, dx, dy, dz)
	postFallVerticalVelocity := e.UpdateFallState(dy, e.OnGround)

	if wantedX != dx {
		e.motion.X = 0
	}
	if postFallVerticalVelocity != nil {
		e.motion.Y = *postFallVerticalVelocity
	} else if wantedY != dy {
		e.motion.Y = 0
	}
	if wantedZ != dz {
		e.motion.Z = 0
	}

	//TODO: vehicle collision events (first we need to spawn them!)
}

func (e *Entity) SetStepHeight(stepHeight float64) { e.stepHeight = stepHeight }

func (e *Entity) GetStepHeight() float64 { return e.stepHeight }

// CheckGroundState is a port of Entity::checkGroundState.
func (e *Entity) CheckGroundState(wantedX, wantedY, wantedZ, dx, dy, dz float64) {
	e.IsCollidedVertically = wantedY != dy
	e.IsCollidedHorizontally = wantedX != dx || wantedZ != dz
	e.IsCollided = e.IsCollidedHorizontally || e.IsCollidedVertically
	e.OnGround = wantedY != dy && wantedY < 0
}

// GetBlocksIntersected is a port of Entity::getBlocksIntersected: every block whose full-cube area
// is intersected by the entity's AABB (shrunk by inset).
func (e *Entity) GetBlocksIntersected(inset float64) []block.Behavior {
	minX := int(stdmath.Floor(e.BoundingBox.MinX + inset))
	minY := int(stdmath.Floor(e.BoundingBox.MinY + inset))
	minZ := int(stdmath.Floor(e.BoundingBox.MinZ + inset))
	maxX := int(stdmath.Floor(e.BoundingBox.MaxX - inset))
	maxY := int(stdmath.Floor(e.BoundingBox.MaxY - inset))
	maxZ := int(stdmath.Floor(e.BoundingBox.MaxZ - inset))

	w := e.GetWorld()

	var blocks []block.Behavior
	for z := minZ; z <= maxZ; z++ {
		for x := minX; x <= maxX; x++ {
			for y := minY; y <= maxY; y++ {
				blocks = append(blocks, w.GetBlockAtIfLoaded(x, y, z))
			}
		}
	}
	return blocks
}

// GetBlocksAroundWithEntityInsideActions is a port of Entity::getBlocksAroundWithEntityInsideActions.
func (e *Entity) GetBlocksAroundWithEntityInsideActions() []block.Behavior {
	if !e.blocksAroundValid {
		e.blocksAround = nil
		e.blocksAroundValid = true

		inset := 0.001 //Offset against floating-point errors
		for _, blk := range e.GetBlocksIntersected(inset) {
			if blk.HasEntityCollision() {
				e.blocksAround = append(e.blocksAround, blk)
			}
		}
	}

	return e.blocksAround
}

// ClearBlocksAround is PHP's `$this->blocksAround = null`: the cached intersecting blocks are
// recalculated on next use.
func (e *Entity) ClearBlocksAround() {
	e.blocksAround = nil
	e.blocksAroundValid = false
}

// CanBeMovedByCurrents returns whether this entity can be moved by currents in liquids.
func (e *Entity) CanBeMovedByCurrents() bool { return true }

// CheckBlockIntersections is a port of Entity::checkBlockIntersections.
func (e *Entity) CheckBlockIntersections() {
	e.checkBlockIntersectionsNextTick = false
	var vectors []math.Vector3

	for _, blk := range e.GetBlocksAroundWithEntityInsideActions() {
		if !blk.OnEntityInside(e.self) {
			e.ClearBlocksAround()
		}
		if v, ok := blk.AddVelocityToEntity(e.self); ok {
			vectors = append(vectors, v)
		}
	}

	if len(vectors) > 0 {
		vector := math.SumVector3(vectors...)
		if vector.LengthSquared() > 0 {
			d := 0.014
			e.motion = e.motion.AddVector(vector.Normalize().Multiply(d))
		}
	}
}

// GetPosition returns the entity's position (PHP's getPosition(), minus the World, which is
// GetWorld()).
func (e *Entity) GetPosition() math.Vector3 { return e.location.Vector3 }

// GetLocation is a port of Entity::getLocation (a copy).
func (e *Entity) GetLocation() Location { return e.location.AsLocation() }

func (e *Entity) GetWorld() *world.World { return e.location.GetWorld() }

// SetLocationDirect is PHP's direct `$this->location = ...` write: no world bookkeeping or bounding
// box update (subtypes that do this, like Projectile::move, follow it with RecalculateBoundingBox).
func (e *Entity) SetLocationDirect(l Location) { e.location = l }

// SetPosition is Entity::setPosition with a plain Vector3 (the entity stays in its world).
func (e *Entity) SetPosition(pos math.Vector3) bool {
	return e.self.SetPositionInWorld(pos, nil)
}

// SetPositionInWorld is a port of Entity::setPosition: w nil means the entity's current world (PHP's
// plain-Vector3 overload); a different world moves the entity across worlds.
func (e *Entity) SetPositionInWorld(pos math.Vector3, w *world.World) bool {
	if e.closed {
		return false
	}

	oldWorld := e.GetWorld()
	newWorld := oldWorld
	if w != nil {
		newWorld = w
	}
	if oldWorld != newWorld {
		e.DespawnFromAll()
		oldWorld.RemoveEntity(e.self)
	}

	e.location = LocationFromObject(pos, newWorld, e.location.Yaw, e.location.Pitch)

	e.self.RecalculateBoundingBox()

	e.ClearBlocksAround()

	if oldWorld != newWorld {
		newWorld.AddEntity(e.self)
	} else {
		newWorld.OnEntityMoved(e.self)
	}

	return true
}

// SetRotation is a port of Entity::setRotation.
func (e *Entity) SetRotation(yaw, pitch float64) {
	checkFloatNotInfOrNaN("yaw", yaw)
	checkFloatNotInfOrNaN("pitch", pitch)
	e.location.Yaw = yaw
	e.location.Pitch = pitch
	e.ScheduleUpdate()
}

// SetPositionAndRotation is a port of Entity::setPositionAndRotation.
func (e *Entity) SetPositionAndRotation(pos math.Vector3, w *world.World, yaw, pitch float64) bool {
	if e.self.SetPositionInWorld(pos, w) {
		e.SetRotation(yaw, pitch)
		return true
	}
	return false
}

// ResetLastMovements is Entity::resetLastMovements (Player::teleport needs it).
func (e *Entity) ResetLastMovements() { e.resetLastMovements() }

// SetYSize sets Entity::$ySize (Player::sendPosition resets it).
func (e *Entity) SetYSize(ySize float64) { e.ySize = ySize }

func (e *Entity) resetLastMovements() {
	e.lastLocation = e.location.AsLocation()
	e.lastMotion = e.motion
}

func (e *Entity) GetMotion() math.Vector3 { return e.motion }

// SetMotion is a port of Entity::setMotion.
func (e *Entity) SetMotion(motion math.Vector3) bool {
	checkFloatNotInfOrNaN("x", motion.X)
	checkFloatNotInfOrNaN("y", motion.Y)
	checkFloatNotInfOrNaN("z", motion.Z)
	if !e.justCreated {
		ev := entityevent.NewEntityMotionEvent(e.self, motion)
		ev.Call()
		if ev.IsCancelled() {
			return false
		}
	}

	e.motion = motion

	if !e.justCreated {
		e.self.UpdateMovement(false)
	}

	return true
}

// SetMotionDirect is PHP's direct `$this->motion = ...` write: no EntityMotionEvent and no
// movement broadcast.
func (e *Entity) SetMotionDirect(motion math.Vector3) { e.motion = motion }

// AddMotion adds the given values to the entity's motion vector.
func (e *Entity) AddMotion(x, y, z float64) {
	checkFloatNotInfOrNaN("x", x)
	checkFloatNotInfOrNaN("y", y)
	checkFloatNotInfOrNaN("z", z)
	e.motion = e.motion.Add(x, y, z)
}

func (e *Entity) IsOnGround() bool { return e.OnGround }

func (e *Entity) SetOnGround(onGround bool) { e.OnGround = onGround }

// Teleport is Entity::teleport with a plain Vector3: the entity stays in its world and keeps its
// rotation.
func (e *Entity) Teleport(pos math.Vector3) bool { return e.TeleportTo(pos, nil, nil, nil) }

// TeleportTo is a port of Entity::teleport: w nil means the current world; yaw/pitch nil keep the
// current rotation (pass a Location's own yaw/pitch to use them, as PHP does for a Location
// argument).
func (e *Entity) TeleportTo(pos math.Vector3, w *world.World, yaw, pitch *float64) bool {
	checkFloatNotInfOrNaN("x", pos.X)
	checkFloatNotInfOrNaN("y", pos.Y)
	checkFloatNotInfOrNaN("z", pos.Z)
	if yaw != nil {
		checkFloatNotInfOrNaN("yaw", *yaw)
	}
	if pitch != nil {
		checkFloatNotInfOrNaN("pitch", *pitch)
	}

	targetWorld := e.GetWorld()
	if w != nil {
		targetWorld = w
	}
	from := entityevent.Position{Vector3: e.location.Vector3, World: e.location.World}
	to := entityevent.Position{Vector3: pos, World: targetWorld}
	ev := entityevent.NewEntityTeleportEvent(e.self, from, to)
	ev.Call()
	if ev.IsCancelled() {
		return false
	}
	e.ySize = 0
	target := ev.GetTo()
	if tw, ok := target.World.(*world.World); ok {
		targetWorld = tw
	}

	newYaw := e.location.Yaw
	if yaw != nil {
		newYaw = *yaw
	}
	newPitch := e.location.Pitch
	if pitch != nil {
		newPitch = *pitch
	}

	e.self.SetMotion(math.NewVector3(0, 0, 0))
	if e.SetPositionAndRotation(target.Vector3, targetWorld, newYaw, newPitch) {
		e.ResetFallDistance()
		e.SetForceMovementUpdate(true)

		e.self.UpdateMovement(true)

		return true
	}

	return false
}

func (e *Entity) GetID() int { return e.id }

// GetViewers returns the players this entity has been spawned to.
func (e *Entity) GetViewers() []world.EntityViewer {
	return append(make([]world.EntityViewer, 0, len(e.hasSpawnedOrder)), e.hasSpawnedOrder...)
}

// GetViewerCount is count(getViewers()) - used by EntityShootBowEvent::setProjectile.
func (e *Entity) GetViewerCount() int { return len(e.hasSpawnedOrder) }

// SendSpawnPacket is a port of Entity::sendSpawnPacket: sends whatever packets are needed to spawn
// the entity to the client.
func (e *Entity) SendSpawnPacket(player world.EntityViewer) {
	attributes := e.attributeMap.GetAll()
	networkAttributes := make([]protocol.AttributeValue, 0, len(attributes))
	for _, attr := range attributes {
		networkAttributes = append(networkAttributes, protocol.AttributeValue{
			Name:  attr.GetID(),
			Value: float32(attr.GetValue()),
			Max:   float32(attr.GetMaxValue()),
			Min:   float32(attr.GetMinValue()),
		})
	}
	player.SendPacket(&packet.AddActor{
		EntityUniqueID:  int64(e.GetID()), //TODO: actor unique ID
		EntityRuntimeID: uint64(e.GetID()),
		EntityType:      e.self.GetNetworkTypeID(),
		Position:        vec32(e.self.GetOffsetPosition(e.location.Vector3)),
		Velocity:        vec32(e.GetMotion()),
		Pitch:           float32(e.location.Pitch),
		Yaw:             float32(e.location.Yaw),
		HeadYaw:         float32(e.location.Yaw), //TODO: head yaw
		BodyYaw:         float32(e.location.Yaw), //TODO: body yaw (wtf mojang?)
		Attributes:      networkAttributes,
		EntityMetadata:  e.GetAllNetworkData(),
		//TODO: entity links
	})
}

// SpawnTo is a port of Entity::spawnTo.
func (e *Entity) SpawnTo(player world.EntityViewer) {
	//TODO: this will cause some visible lag during chunk resends; if the player uses a spawn egg in a chunk, the
	//created entity won't be visible until after the resend arrives. However, this is better than possibly crashing
	//the player by sending them entities too early.
	if !e.hasSpawned[player] && player.GetWorld() == e.GetWorld() && player.HasReceivedChunk(e.location.FloorX()>>4, e.location.FloorZ()>>4) {
		e.hasSpawned[player] = true
		e.hasSpawnedOrder = append(e.hasSpawnedOrder, player)

		e.self.SendSpawnPacket(player)
	}
}

// SpawnToAll is a port of Entity::spawnToAll.
func (e *Entity) SpawnToAll() {
	if e.closed {
		return
	}
	for _, player := range e.GetWorld().GetViewersForPosition(e.location.Vector3) {
		e.self.SpawnTo(player)
	}
}

// RespawnToAll is a port of Entity::respawnToAll.
func (e *Entity) RespawnToAll() {
	viewers := e.GetViewers()
	for _, player := range viewers {
		e.forgetViewer(player)
		e.self.SpawnTo(player)
	}
}

func (e *Entity) forgetViewer(player world.EntityViewer) {
	delete(e.hasSpawned, player)
	for i, v := range e.hasSpawnedOrder {
		if v == player {
			e.hasSpawnedOrder = append(e.hasSpawnedOrder[:i:i], e.hasSpawnedOrder[i+1:]...)
			break
		}
	}
}

// DespawnFrom is a port of Entity::despawnFrom.
//
// Deprecated (in PHP too): this does NOT permanently hide the entity from the player. As soon as
// the entity or player moves, the player will once again be able to see the entity.
func (e *Entity) DespawnFrom(player world.EntityViewer, send bool) {
	if e.hasSpawned[player] {
		if send {
			player.SendPacket(removeActorPacket(e.id))
		}
		e.forgetViewer(player)
	}
}

// DespawnFromAll is a port of Entity::despawnFromAll.
//
// Deprecated (in PHP too): see DespawnFrom.
func (e *Entity) DespawnFromAll() {
	BroadcastPackets(e.hasSpawnedOrder, removeActorPacket(e.id))
	e.hasSpawned = map[world.EntityViewer]bool{}
	e.hasSpawnedOrder = nil
}

// GetPickedItem returns the item that players will equip when middle-clicking on this entity (nil
// if none).
func (e *Entity) GetPickedItem() item.Item { return nil }

// FlagForDespawn flags the entity to be removed from the world on the next tick.
func (e *Entity) FlagForDespawn() {
	e.needsDespawn = true
	e.ScheduleUpdate()
}

func (e *Entity) IsFlaggedForDespawn() bool { return e.needsDespawn }

// IsClosed returns whether the entity has been "closed".
func (e *Entity) IsClosed() bool { return e.closed }

// Close is a port of Entity::close: closes the entity and frees attached references. Entities are
// unusable after this has been executed!
func (e *Entity) Close() {
	if e.closeInFlight {
		return
	}

	if !e.closed {
		e.closeInFlight = true
		entityevent.NewEntityDespawnEvent(e.self).Call()

		e.self.OnDispose()
		e.closed = true
		e.self.DestroyCycles()
		e.closeInFlight = false
	}
}

// OnDispose is called when the entity is disposed to clean up things like viewers. This SHOULD NOT
// destroy internal state, because it may be needed by descendent classes.
func (e *Entity) OnDispose() {
	e.DespawnFromAll()
	if e.location.IsValid() {
		e.GetWorld().RemoveEntity(e.self)
	}
}

// DestroyCycles is called when the entity is disposed, after all events have been fired.
func (e *Entity) DestroyCycles() {
	e.lastDamageCause = nil
}

// SendData is a port of Entity::sendData: nil targets means every viewer; nil data means every
// property.
func (e *Entity) SendData(targets []world.EntityViewer, data protocol.EntityMetadata) {
	if targets == nil {
		targets = e.hasSpawnedOrder
	}
	if data == nil {
		data = e.GetAllNetworkData()
	}
	BroadcastPackets(targets, &packet.SetActorData{EntityRuntimeID: uint64(e.id), EntityMetadata: data})
}

// MarkNetworkPropertiesDirty is PHP's `$this->networkPropertiesDirty = true`: SyncNetworkData runs
// again before the next network update.
func (e *Entity) MarkNetworkPropertiesDirty() { e.networkPropertiesDirty = true }

// GetDirtyNetworkData is a port of Entity::getDirtyNetworkData.
func (e *Entity) GetDirtyNetworkData() protocol.EntityMetadata {
	if e.networkPropertiesDirty {
		e.self.SyncNetworkData(e.networkProperties)
		e.networkPropertiesDirty = false
	}
	return e.networkProperties.GetDirty()
}

// GetAllNetworkData is a port of Entity::getAllNetworkData.
func (e *Entity) GetAllNetworkData() protocol.EntityMetadata {
	if e.networkPropertiesDirty {
		e.self.SyncNetworkData(e.networkProperties)
		e.networkPropertiesDirty = false
	}
	return e.networkProperties.GetAll()
}

func optionalID(id *int, fallback int64) int64 {
	if id == nil {
		return fallback
	}
	return int64(*id)
}

// SyncNetworkData is a port of Entity::syncNetworkData.
func (e *Entity) SyncNetworkData(properties *MetadataCollection) {
	properties.SetByte(MetadataAlwaysShowNametag, byte(boolByte(e.alwaysShowNameTag)))
	properties.SetFloat(MetadataBoundingBoxHeight, float32(e.Size.GetHeight()/e.scale))
	properties.SetFloat(MetadataBoundingBoxWidth, float32(e.Size.GetWidth()/e.scale))
	properties.SetFloat(MetadataScale, float32(e.scale))
	properties.SetLong(MetadataLeadHolderEID, -1)
	properties.SetLong(MetadataOwnerEID, optionalID(e.ownerID, -1))
	properties.SetLong(MetadataTargetEID, optionalID(e.targetID, 0))
	properties.SetString(MetadataNametag, e.nameTag)
	properties.SetString(MetadataScoreTag, e.scoreTag)
	properties.SetByte(MetadataColor, 0)

	properties.SetGenericFlag(FlagAffectedByGravity, e.gravityEnabled)
	properties.SetGenericFlag(FlagCanClimb, e.canClimb)
	properties.SetGenericFlag(FlagCanShowNametag, e.nameTagVisible)
	properties.SetGenericFlag(FlagHasCollision, true)
	properties.SetGenericFlag(FlagNoAI, e.noClientPredictions)
	properties.SetGenericFlag(FlagInvisible, e.invisible)
	properties.SetGenericFlag(FlagSilent, e.silent)
	properties.SetGenericFlag(FlagOnFire, e.IsOnFire())
	properties.SetGenericFlag(FlagWallClimbing, e.canClimbWalls)
}

// BroadcastAnimation is a port of Entity::broadcastAnimation (nil targets means every viewer).
func (e *Entity) BroadcastAnimation(anim animation.Animation, targets []world.EntityViewer) {
	if targets == nil {
		targets = e.hasSpawnedOrder
	}
	BroadcastPackets(targets, anim.Encode()...)
}

// BroadcastSound is a port of Entity::broadcastSound to every viewer: dropped if the entity is
// silent.
func (e *Entity) BroadcastSound(s sound.Sound) { e.BroadcastSoundTo(s, nil) }

// BroadcastSoundTo is Entity::broadcastSound with explicit targets (nil means every viewer).
func (e *Entity) BroadcastSoundTo(s sound.Sound, targets []world.EntityViewer) {
	if !e.silent {
		if targets == nil {
			targets = e.GetViewers()
		}
		e.GetWorld().AddSoundFor(e.location.Vector3, s, targets)
	}
}

func (e *Entity) String() string {
	return fmt.Sprintf("%s(%d)", reflect.TypeOf(e.self).Elem().Name(), e.GetID())
}
