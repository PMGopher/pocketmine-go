package projectile

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/sound"
)

// Trident NBT keys, a port of Trident's TAG_* constants.
const (
	TagTridentItem       = "Trident"    //TAG_Compound
	tagSpawnedInCreative = "isCreative" //TAG_Byte
)

// Trident is a port of pocketmine\entity\projectile\Trident.
type Trident struct {
	Projectile

	item              item.Item
	canCollide        bool
	spawnedInCreative bool
}

// NewTrident is a port of Trident::__construct (panicking for an empty item; shootingEntity may be
// nil).
func NewTrident(location entity.Location, it item.Item, shootingEntity world.Entity, tag *nbt.CompoundTag) *Trident {
	if it.IsNull() {
		panic("Trident must have a count of at least 1")
	}
	t := &Trident{item: it.Clone(), canCollide: true}
	t.damage = 8.0
	t.ConstructProjectile(t, location, shootingEntity, tag)
	return t
}

func (t *Trident) GetNetworkTypeID() string { return entity.EntityIDThrownTrident }

func (t *Trident) GetInitialSizeInfo() entity.EntitySizeInfo {
	return entity.NewEntitySizeInfo(0.35, 0.25)
}

func (t *Trident) GetInitialDragMultiplier() float64 { return 0.01 }

func (t *Trident) GetInitialGravity() float64 { return 0.1 }

// InitEntity is a port of Trident::initEntity.
func (t *Trident) InitEntity(tag *nbt.CompoundTag) {
	t.Projectile.InitEntity(tag)

	t.spawnedInCreative = tag.GetByteOr(tagSpawnedInCreative, 0) == 1
}

// SaveNBT is a port of Trident::saveNBT.
func (t *Trident) SaveNBT() *nbt.CompoundTag {
	tag := t.Projectile.SaveNBT()
	if itemTag, err := item.NbtSerialize(t.item, -1); err == nil {
		tag.SetTag(TagTridentItem, itemTag)
	}
	creative := nbt.ByteTag(0)
	if t.spawnedInCreative {
		creative = 1
	}
	tag.SetByte(tagSpawnedInCreative, creative)
	return tag
}

// OnFirstUpdate is a port of Trident::onFirstUpdate.
func (t *Trident) OnFirstUpdate(currentTick int64) {
	owner, isPlayer := entity.AsPlayer(t.GetOwningEntity())
	t.spawnedInCreative = isPlayer && owner.IsCreative()

	t.Projectile.OnFirstUpdate(currentTick)
}

// EntityBaseTick is a port of Trident::entityBaseTick.
func (t *Trident) EntityBaseTick(tickDiff int) bool {
	if t.IsClosed() {
		return false
	}
	//TODO: Loyalty enchantment.

	return t.Projectile.EntityBaseTick(tickDiff)
}

func (t *Trident) DespawnsOnEntityHit() bool { return false }

// OnHitEntity is a port of Trident::onHitEntity.
func (t *Trident) OnHitEntity(entityHit world.Entity, hitResult math.RayTraceResult) {
	t.Projectile.OnHitEntity(entityHit, hitResult)
	t.canCollide = false
	t.BroadcastSound(sound.TridentHitEntitySound{})
	motion := t.GetMotion()
	t.SetMotion(math.NewVector3(motion.X*-0.01, motion.Y*-0.1, motion.Z*-0.01))
}

// OnHitBlock is a port of Trident::onHitBlock.
func (t *Trident) OnHitBlock(blockHit block.Behavior, hitResult math.RayTraceResult) {
	t.Projectile.OnHitBlock(blockHit, hitResult)
	t.canCollide = true
	t.BroadcastSound(sound.TridentHitBlockSound{})
}

func (t *Trident) GetItem() item.Item { return t.item.Clone() }

// SetItem panics for an empty item.
func (t *Trident) SetItem(it item.Item) {
	if it.IsNull() {
		panic("Trident must have a count of at least 1")
	}
	if t.item.HasEnchantments() != it.HasEnchantments() {
		t.MarkNetworkPropertiesDirty()
	}
	t.item = it.Clone()
}

// CanCollideWith is a port of Trident::canCollideWith.
func (t *Trident) CanCollideWith(other world.Entity) bool {
	ownerID, hasOwner := t.GetOwningEntityID()
	return t.canCollide && (!hasOwner || other.GetID() != ownerID) && t.Projectile.CanCollideWith(other)
}

// OnCollideWithPlayer is a port of Trident::onCollideWithPlayer.
func (t *Trident) OnCollideWithPlayer(player entity.Player) {
	if t.blockHit != nil {
		t.pickup(player)
	}
}

func (t *Trident) pickup(player entity.Player) {
	shouldDespawn := false

	playerInventory := player.GetInventory()
	ev := entityevent.NewEntityItemPickupEvent(player, t, t.GetItem(), playerInventory)
	if picked, ok := ev.GetItem().(item.Item); ok && player.HasFiniteResources() && !playerInventory.CanAddItem(picked) {
		ev.Cancel()
	}
	if t.spawnedInCreative {
		ev.Cancel()
		shouldDespawn = true
	}

	ev.Call()
	if !ev.IsCancelled() {
		if inv, ok := ev.GetInventory().(inventory.Inventory); ok && inv != nil {
			if picked, ok := ev.GetItem().(item.Item); ok {
				inv.AddItem(picked)
			}
		}
		shouldDespawn = true
	}

	if shouldDespawn {
		//even if the item was not actually picked up, the animation must be displayed.
		entity.BroadcastPickUpItem(t.GetViewers(), player.GetID(), t.GetID())
		t.FlagForDespawn()
	}
}

// SyncNetworkData is a port of Trident::syncNetworkData.
func (t *Trident) SyncNetworkData(properties *entity.MetadataCollection) {
	t.Projectile.SyncNetworkData(properties)

	properties.SetGenericFlag(entity.FlagEnchanted, t.item.HasEnchantments())
}
