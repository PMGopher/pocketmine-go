package object

import (
	"fmt"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/animation"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world"
)

// ItemEntity NBT keys, a port of ItemEntity's TAG_* constants.
const (
	tagItemHealth      = "Health"      //TAG_Short
	tagItemAge         = "Age"         //TAG_Short
	tagItemPickupDelay = "PickupDelay" //TAG_Short
	tagItemOwner       = "Owner"       //TAG_String
	tagItemThrower     = "Thrower"     //TAG_String
	TagItem            = "Item"        //TAG_Compound
)

// ItemEntity constants, a port of ItemEntity's public constants.
const (
	MergeCheckPeriod    = 2    //0.1 seconds
	DefaultDespawnDelay = 6000 //5 minutes
	NeverDespawn        = -1
	MaxDespawnDelay     = 32767 + DefaultDespawnDelay //max value storable by mojang NBT :(
)

// ItemEntity is a port of pocketmine\entity\object\ItemEntity.
type ItemEntity struct {
	entity.Entity

	owner        string
	thrower      string
	pickupDelay  int
	despawnDelay int
	item         item.Item
}

// NewItemEntity is a port of ItemEntity::__construct. Panics for an empty item (PHP's
// InvalidArgumentException).
func NewItemEntity(location entity.Location, it item.Item, tag *nbt.CompoundTag) *ItemEntity {
	if it.IsNull() {
		panic("Item entity must have a non-air item with a count of at least 1")
	}
	i := &ItemEntity{item: it.Clone(), despawnDelay: DefaultDespawnDelay}
	i.Construct(i, location, tag)
	return i
}

func (i *ItemEntity) GetNetworkTypeID() string { return entity.EntityIDItem }

func (i *ItemEntity) GetInitialSizeInfo() entity.EntitySizeInfo {
	return entity.NewEntitySizeInfo(0.25, 0.25)
}

func (i *ItemEntity) GetInitialDragMultiplier() float64 { return 0.02 }

func (i *ItemEntity) GetInitialGravity() float64 { return 0.04 }

// InitEntity is a port of ItemEntity::initEntity.
func (i *ItemEntity) InitEntity(tag *nbt.CompoundTag) {
	i.Entity.InitEntity(tag)

	i.SetMaxHealth(5)
	i.SetHealth(float64(tag.GetShortOr(tagItemHealth, nbt.ShortTag(int(i.GetHealth())))))

	age := int(tag.GetShortOr(tagItemAge, 0))
	if age == -32768 {
		i.despawnDelay = NeverDespawn
	} else {
		i.despawnDelay = max(0, DefaultDespawnDelay-age)
	}
	i.pickupDelay = int(tag.GetShortOr(tagItemPickupDelay, nbt.ShortTag(i.pickupDelay)))
	i.owner = string(tag.GetStringOr(tagItemOwner, nbt.StringTag(i.owner)))
	i.thrower = string(tag.GetStringOr(tagItemThrower, nbt.StringTag(i.thrower)))
}

// OnFirstUpdate is a port of ItemEntity::onFirstUpdate.
func (i *ItemEntity) OnFirstUpdate(currentTick int64) {
	entityevent.NewItemSpawnEvent(i).Call() //this must be called before EntitySpawnEvent, to maintain backwards compatibility
	i.Entity.OnFirstUpdate(currentTick)
}

// EntityBaseTick is a port of ItemEntity::entityBaseTick: pickup delay, merging and despawning.
func (i *ItemEntity) EntityBaseTick(tickDiff int) bool {
	if i.IsClosed() {
		return false
	}

	hasUpdate := i.Entity.EntityBaseTick(tickDiff)

	if i.IsFlaggedForDespawn() {
		return hasUpdate
	}

	if i.pickupDelay != NeverDespawn && i.pickupDelay > 0 { //Infinite delay
		hasUpdate = true
		i.pickupDelay -= tickDiff
		if i.pickupDelay < 0 {
			i.pickupDelay = 0
		}
	}

	if i.HasMovementUpdate() && i.isMergeCandidate() && i.despawnDelay%MergeCheckPeriod == 0 {
		mergeable := []*ItemEntity{i} //in case the merge target ends up not being this
		mergeTarget := i
		for _, e := range i.GetWorld().GetNearbyEntitiesExcept(i.BoundingBox.ExpandedCopy(0.5, 0.5, 0.5), i) {
			other, ok := e.(*ItemEntity)
			if !ok || other.IsFlaggedForDespawn() {
				continue
			}

			if other.IsMergeable(i) {
				mergeable = append(mergeable, other)
				if other.item.GetCount() > mergeTarget.item.GetCount() {
					mergeTarget = other
				}
			}
		}
		for _, itemEntity := range mergeable {
			if itemEntity != mergeTarget {
				itemEntity.TryMergeInto(mergeTarget)
			}
		}
	}

	if !i.IsFlaggedForDespawn() && i.despawnDelay != NeverDespawn {
		hasUpdate = true
		i.despawnDelay -= tickDiff
		if i.despawnDelay <= 0 {
			ev := entityevent.NewItemDespawnEvent(i)
			ev.Call()
			if ev.IsCancelled() {
				i.despawnDelay = DefaultDespawnDelay
			} else {
				i.FlagForDespawn()
			}
		}
	}

	return hasUpdate
}

func (i *ItemEntity) isMergeCandidate() bool {
	return i.pickupDelay != NeverDespawn && i.item.GetCount() < i.item.GetMaxStackSize()
}

// IsMergeable returns whether this item entity can merge with the given one.
func (i *ItemEntity) IsMergeable(other *ItemEntity) bool {
	if !i.isMergeCandidate() || !other.isMergeCandidate() {
		return false
	}
	it := other.item
	return other != i && it.CanStackWith(i.item) && it.GetCount()+i.item.GetCount() <= it.GetMaxStackSize()
}

// TryMergeInto attempts to merge this item entity into the given item entity. Returns true if it
// was successful.
func (i *ItemEntity) TryMergeInto(consumer *ItemEntity) bool {
	if !i.IsMergeable(consumer) {
		return false
	}

	ev := entityevent.NewItemMergeEvent(i, consumer)
	ev.Call()

	if ev.IsCancelled() {
		return false
	}

	consumer.SetStackSize(consumer.item.GetCount() + i.item.GetCount())
	i.FlagForDespawn()
	consumer.pickupDelay = max(consumer.pickupDelay, i.pickupDelay)
	consumer.despawnDelay = max(consumer.despawnDelay, i.despawnDelay)

	return true
}

// TryChangeMovement is a port of ItemEntity::tryChangeMovement.
func (i *ItemEntity) TryChangeMovement() {
	pos := i.GetPosition()
	i.CheckObstruction(pos.X, pos.Y, pos.Z)
	i.Entity.TryChangeMovement()
}

func (i *ItemEntity) ApplyDragBeforeGravity() bool { return true }

// CanSaveWithChunk is a port of ItemEntity::canSaveWithChunk.
func (i *ItemEntity) CanSaveWithChunk() bool {
	return !i.item.IsNull() && i.Entity.CanSaveWithChunk()
}

// SaveNBT is a port of ItemEntity::saveNBT.
func (i *ItemEntity) SaveNBT() *nbt.CompoundTag {
	tag := i.Entity.SaveNBT()
	if itemTag, err := item.NbtSerialize(i.item, -1); err == nil {
		tag.SetTag(TagItem, itemTag)
	}
	tag.SetShort(tagItemHealth, nbt.ShortTag(int(i.GetHealth())))
	age := -32768
	if i.despawnDelay != NeverDespawn {
		age = DefaultDespawnDelay - i.despawnDelay
	}
	tag.SetShort(tagItemAge, nbt.ShortTag(age))
	tag.SetShort(tagItemPickupDelay, nbt.ShortTag(i.pickupDelay))
	tag.SetString(tagItemOwner, nbt.StringTag(i.owner))
	tag.SetString(tagItemThrower, nbt.StringTag(i.thrower))

	return tag
}

func (i *ItemEntity) GetItem() item.Item { return i.item }

func (i *ItemEntity) IsFireProof() bool { return i.item.IsFireProof() }

func (i *ItemEntity) CanCollideWith(other world.Entity) bool { return false }

func (i *ItemEntity) CanBeCollidedWith() bool { return false }

func (i *ItemEntity) GetPickupDelay() int { return i.pickupDelay }

func (i *ItemEntity) SetPickupDelay(delay int) { i.pickupDelay = delay }

// GetDespawnDelay returns the number of ticks left before this item will despawn. If -1, the item
// will never despawn.
func (i *ItemEntity) GetDespawnDelay() int { return i.despawnDelay }

// SetDespawnDelay panics outside 0..MaxDespawnDelay (and != NeverDespawn), like PHP's
// InvalidArgumentException.
func (i *ItemEntity) SetDespawnDelay(despawnDelay int) {
	if (despawnDelay < 0 || despawnDelay > MaxDespawnDelay) && despawnDelay != NeverDespawn {
		panic(fmt.Sprintf("Despawn ticker must be in range 0 ... %d or %d, got %d", MaxDespawnDelay, NeverDespawn, despawnDelay))
	}
	i.despawnDelay = despawnDelay
}

func (i *ItemEntity) GetOwner() string { return i.owner }

func (i *ItemEntity) SetOwner(owner string) { i.owner = owner }

func (i *ItemEntity) GetThrower() string { return i.thrower }

func (i *ItemEntity) SetThrower(thrower string) { i.thrower = thrower }

// SendSpawnPacket is a port of ItemEntity::sendSpawnPacket.
func (i *ItemEntity) SendSpawnPacket(player world.EntityViewer) {
	player.SendPacket(&packet.AddItemActor{
		EntityUniqueID:  int64(i.GetID()), //TODO: entity unique ID
		EntityRuntimeID: uint64(i.GetID()),
		Item:            convert.ItemStackWrapperLegacy(i.GetItem()),
		Position:        vec32(i.GetPosition()),
		Velocity:        vec32(i.GetMotion()),
		EntityMetadata:  i.GetAllNetworkData(),
		FromFishing:     false, //TODO: I have no idea what this is needed for, but right now we don't support fishing anyway
	})
}

// SetStackSize is a port of ItemEntity::setStackSize (panicking below 1).
func (i *ItemEntity) SetStackSize(newCount int) {
	if newCount <= 0 {
		panic("Stack size must be at least 1")
	}
	i.item.SetCount(newCount)
	i.BroadcastAnimation(animation.ItemEntityStackSizeChangeAnimation{ItemEntity: i, NewStackSize: newCount}, nil)
}

func (i *ItemEntity) GetOffsetPosition(v math.Vector3) math.Vector3 { return v.Add(0, 0.125, 0) }

// OnCollideWithPlayer is a port of ItemEntity::onCollideWithPlayer: the player picks the item up.
func (i *ItemEntity) OnCollideWithPlayer(player entity.Player) {
	if i.GetPickupDelay() != 0 {
		return
	}

	it := i.GetItem()
	var playerInventory inventory.Inventory
	switch {
	case player.GetOffHandInventory().GetItem(0).CanStackWith(it) && player.GetOffHandInventory().GetAddableItemQuantity(it) > 0:
		playerInventory = player.GetOffHandInventory()
	case player.GetInventory().GetAddableItemQuantity(it) > 0:
		playerInventory = player.GetInventory()
	}

	var eventInventory entityevent.Inventory
	if playerInventory != nil {
		eventInventory = playerInventory
	}
	ev := entityevent.NewEntityItemPickupEvent(player, i, it, eventInventory)
	if player.HasFiniteResources() && playerInventory == nil {
		ev.Cancel()
	}

	ev.Call()
	if ev.IsCancelled() {
		return
	}

	entity.BroadcastPickUpItem(i.GetViewers(), player.GetID(), i.GetID())

	if inv, ok := ev.GetInventory().(inventory.Inventory); ok && inv != nil {
		if picked, ok := ev.GetItem().(item.Item); ok {
			for _, remains := range inv.AddItem(picked) {
				zero := math.Vector3Zero()
				i.GetWorld().DropItem(i.GetPosition(), remains, &zero, 10)
			}
		}
	}
	i.FlagForDespawn()
}
