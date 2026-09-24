package projectile

import (
	stdmath "math"
	"math/rand/v2"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/animation"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// Arrow pickup modes, a port of Arrow::PICKUP_*.
const (
	ArrowPickupNone     = 0
	ArrowPickupAny      = 1
	ArrowPickupCreative = 2
)

// Arrow NBT keys, a port of Arrow's TAG_* constants.
const (
	tagPickup = "pickup" //TAG_Byte
	TagCrit   = "crit"   //TAG_Byte
	tagLife   = "life"   //TAG_Short
)

// Arrow is a port of pocketmine\entity\projectile\Arrow.
type Arrow struct {
	Projectile

	pickupMode     int
	punchKnockback float64
	collideTicks   int
	critical       bool
}

// NewArrow is a port of Arrow::__construct (shootingEntity may be nil).
func NewArrow(location entity.Location, shootingEntity world.Entity, critical bool, tag *nbt.CompoundTag) *Arrow {
	a := &Arrow{pickupMode: ArrowPickupAny}
	a.damage = 2.0
	a.ConstructProjectile(a, location, shootingEntity, tag)
	a.SetCritical(critical)
	return a
}

func (a *Arrow) GetNetworkTypeID() string { return entity.EntityIDArrow }

func (a *Arrow) GetInitialSizeInfo() entity.EntitySizeInfo {
	return entity.NewEntitySizeInfo(0.25, 0.25)
}

func (a *Arrow) GetInitialDragMultiplier() float64 { return 0.01 }

func (a *Arrow) GetInitialGravity() float64 { return 0.05 }

// InitEntity is a port of Arrow::initEntity.
func (a *Arrow) InitEntity(tag *nbt.CompoundTag) {
	a.Projectile.InitEntity(tag)

	a.pickupMode = int(tag.GetByteOr(tagPickup, ArrowPickupAny))
	a.critical = tag.GetByteOr(TagCrit, 0) == 1
	a.collideTicks = int(tag.GetShortOr(tagLife, nbt.ShortTag(a.collideTicks)))
}

// SaveNBT is a port of Arrow::saveNBT.
func (a *Arrow) SaveNBT() *nbt.CompoundTag {
	tag := a.Projectile.SaveNBT()
	tag.SetByte(tagPickup, nbt.ByteTag(a.pickupMode))
	crit := nbt.ByteTag(0)
	if a.critical {
		crit = 1
	}
	tag.SetByte(TagCrit, crit)
	tag.SetShort(tagLife, nbt.ShortTag(a.collideTicks))
	return tag
}

func (a *Arrow) IsCritical() bool { return a.critical }

func (a *Arrow) SetCritical(value bool) {
	a.critical = value
	a.MarkNetworkPropertiesDirty()
}

// GetResultDamage is a port of Arrow::getResultDamage: damage scales with speed, and critical
// arrows deal a random bonus.
func (a *Arrow) GetResultDamage() int {
	base := int(stdmath.Ceil(a.GetMotion().Length() * float64(a.Projectile.GetResultDamage())))
	if a.IsCritical() {
		return base + rand.IntN(base/2+2)
	}
	return base
}

func (a *Arrow) GetPunchKnockback() float64 { return a.punchKnockback }

func (a *Arrow) SetPunchKnockback(punchKnockback float64) { a.punchKnockback = punchKnockback }

// EntityBaseTick is a port of Arrow::entityBaseTick: stuck arrows despawn after a minute.
func (a *Arrow) EntityBaseTick(tickDiff int) bool {
	if a.IsClosed() {
		return false
	}

	hasUpdate := a.Projectile.EntityBaseTick(tickDiff)

	if a.blockHit != nil {
		a.collideTicks += tickDiff
		if a.collideTicks > 1200 {
			a.FlagForDespawn()
			hasUpdate = true
		}
	} else {
		a.collideTicks = 0
	}

	return hasUpdate
}

// OnHit is a port of Arrow::onHit.
func (a *Arrow) OnHit(event entityevent.ProjectileHit) {
	a.SetCritical(false)
	a.BroadcastSound(sound.ArrowHitSound{})
}

// OnHitBlock is a port of Arrow::onHitBlock.
func (a *Arrow) OnHitBlock(blockHit block.Behavior, hitResult math.RayTraceResult) {
	a.Projectile.OnHitBlock(blockHit, hitResult)
	a.BroadcastAnimation(animation.ArrowShakeAnimation{Arrow: a, DurationInTicks: 7}, nil)
}

// OnHitEntity is a port of Arrow::onHitEntity (Punch knockback on top of the damage).
func (a *Arrow) OnHitEntity(entityHit world.Entity, hitResult math.RayTraceResult) {
	a.Projectile.OnHitEntity(entityHit, hitResult)
	if a.punchKnockback > 0 {
		motion := a.GetMotion()
		horizontalSpeed := stdmath.Sqrt(motion.X*motion.X + motion.Z*motion.Z)
		if horizontalSpeed > 0 {
			multiplier := a.punchKnockback * 0.6 / horizontalSpeed
			entityHit.SetMotion(entityHit.GetMotion().Add(motion.X*multiplier, 0.1, motion.Z*multiplier))
		}
	}
}

func (a *Arrow) GetPickupMode() int { return a.pickupMode }

func (a *Arrow) SetPickupMode(pickupMode int) { a.pickupMode = pickupMode }

// OnCollideWithPlayer is a port of Arrow::onCollideWithPlayer: a stuck arrow is picked up.
func (a *Arrow) OnCollideWithPlayer(player entity.Player) {
	if a.blockHit == nil {
		return
	}

	it := item.VanillaArrow()
	var playerInventory inventory.Inventory
	switch {
	case !player.HasFiniteResources(): //arrows are not picked up in creative
	case player.GetOffHandInventory().GetItem(0).CanStackWith(it) && player.GetOffHandInventory().CanAddItem(it):
		playerInventory = player.GetOffHandInventory()
	case player.GetInventory().CanAddItem(it):
		playerInventory = player.GetInventory()
	}

	var eventInventory entityevent.Inventory
	if playerInventory != nil {
		eventInventory = playerInventory
	}
	ev := entityevent.NewEntityItemPickupEvent(player, a, it, eventInventory)
	if player.HasFiniteResources() && playerInventory == nil {
		ev.Cancel()
	}
	if a.pickupMode == ArrowPickupNone || (a.pickupMode == ArrowPickupCreative && !player.IsCreative()) {
		ev.Cancel()
	}

	ev.Call()
	if ev.IsCancelled() {
		return
	}

	entity.BroadcastPickUpItem(a.GetViewers(), player.GetID(), a.GetID())

	if inv, ok := ev.GetInventory().(inventory.Inventory); ok && inv != nil {
		if picked, ok := ev.GetItem().(item.Item); ok {
			inv.AddItem(picked)
		}
	}
	a.FlagForDespawn()
}

// SyncNetworkData is a port of Arrow::syncNetworkData.
func (a *Arrow) SyncNetworkData(properties *entity.MetadataCollection) {
	a.Projectile.SyncNetworkData(properties)

	properties.SetGenericFlag(entity.FlagCritical, a.critical)
}
