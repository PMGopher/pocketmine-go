package mcpe

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/network/mcpe/convert"
)

// EntityEventBroadcaster is a port of pocketmine\network\mcpe\EntityEventBroadcaster: turns entity
// events into packets for the given sessions.
//
// PHP routes every entity event through each viewer's session broadcaster. In this port entities
// send most of their packets straight to their viewers (world.EntityViewer.SendPacket, see
// entity/network.go), so NetworkSession only uses this for the events it's the source of: the
// player's own attributes and effects.
type EntityEventBroadcaster interface {
	SyncAttributes(recipients []*NetworkSession, e entityWithID, attributes []*entity.Attribute)
	SyncActorData(recipients []*NetworkSession, e entityWithID, properties protocol.EntityMetadata)
	OnEntityEffectAdded(recipients []*NetworkSession, e entityWithID, instance *effect.EffectInstance, replacesOldEffect bool)
	OnEntityEffectRemoved(recipients []*NetworkSession, e entityWithID, instance *effect.EffectInstance)
	OnEntityRemoved(recipients []*NetworkSession, e entityWithID)
	OnMobMainHandItemChange(recipients []*NetworkSession, mob *entity.Human)
	OnMobOffHandItemChange(recipients []*NetworkSession, mob *entity.Human)
	OnMobArmorChange(recipients []*NetworkSession, mob *entity.Living)
	OnPickUpItem(recipients []*NetworkSession, collector, pickedUp entityWithID)
	OnEmote(recipients []*NetworkSession, from entityWithID, emoteID string)
}

// entityWithID is the part of an entity the broadcaster needs (PHP passes the Entity itself).
type entityWithID interface{ GetID() int }

// StandardEntityEventBroadcaster is a port of
// pocketmine\network\mcpe\StandardEntityEventBroadcaster.
type StandardEntityEventBroadcaster struct {
	broadcaster PacketBroadcaster
}

func NewStandardEntityEventBroadcaster(broadcaster PacketBroadcaster) *StandardEntityEventBroadcaster {
	return &StandardEntityEventBroadcaster{broadcaster: broadcaster}
}

func (b *StandardEntityEventBroadcaster) sendDataPacket(recipients []*NetworkSession, pk packet.Packet) {
	b.broadcaster.BroadcastPackets(recipients, []packet.Packet{pk})
}

func (b *StandardEntityEventBroadcaster) SyncAttributes(recipients []*NetworkSession, e entityWithID, attributes []*entity.Attribute) {
	if len(attributes) > 0 {
		b.sendDataPacket(recipients, entity.SyncAttributesPacket(e.GetID(), attributes))
	}
}

func (b *StandardEntityEventBroadcaster) SyncActorData(recipients []*NetworkSession, e entityWithID, properties protocol.EntityMetadata) {
	//TODO: HACK! as of 1.18.10, the client responds differently to the same data ordered in different orders - for
	//example, sending HEIGHT in the list before FLAGS when unsetting the SWIMMING flag results in a hitbox glitch
	// (gophertunnel writes EntityMetadata in ascending key order already.)
	b.sendDataPacket(recipients, &packet.SetActorData{EntityRuntimeID: uint64(e.GetID()), EntityMetadata: properties, Tick: 0})
}

func (b *StandardEntityEventBroadcaster) OnEntityEffectAdded(recipients []*NetworkSession, e entityWithID, instance *effect.EffectInstance, replacesOldEffect bool) {
	b.sendDataPacket(recipients, entity.EntityEffectAddedPacket(e.GetID(), instance, replacesOldEffect))
}

func (b *StandardEntityEventBroadcaster) OnEntityEffectRemoved(recipients []*NetworkSession, e entityWithID, instance *effect.EffectInstance) {
	b.sendDataPacket(recipients, entity.EntityEffectRemovedPacket(e.GetID(), instance))
}

func (b *StandardEntityEventBroadcaster) OnEntityRemoved(recipients []*NetworkSession, e entityWithID) {
	b.sendDataPacket(recipients, &packet.RemoveActor{EntityUniqueID: int64(e.GetID())})
}

func (b *StandardEntityEventBroadcaster) OnMobMainHandItemChange(recipients []*NetworkSession, mob *entity.Human) {
	//TODO: we could send zero for slot here because remote players don't need to know which slot was selected
	inv := mob.GetInventory()
	b.sendDataPacket(recipients, &packet.MobEquipment{
		EntityRuntimeID: uint64(mob.GetID()),
		NewItem:         convert.ItemStackWrapperLegacy(inv.GetItemInHand()),
		InventorySlot:   byte(inv.GetHeldItemIndex()),
		HotBarSlot:      byte(inv.GetHeldItemIndex()),
		WindowID:        ContainerIDInventory,
	})
}

func (b *StandardEntityEventBroadcaster) OnMobOffHandItemChange(recipients []*NetworkSession, mob *entity.Human) {
	b.sendDataPacket(recipients, &packet.MobEquipment{
		EntityRuntimeID: uint64(mob.GetID()),
		NewItem:         convert.ItemStackWrapperLegacy(mob.GetOffHandInventory().GetItem(0)),
		WindowID:        ContainerIDOffhand,
	})
}

func (b *StandardEntityEventBroadcaster) OnMobArmorChange(recipients []*NetworkSession, mob *entity.Living) {
	inv := mob.GetArmorInventory()
	b.sendDataPacket(recipients, &packet.MobArmourEquipment{
		EntityRuntimeID: uint64(mob.GetID()),
		Helmet:          convert.ItemStackWrapperLegacy(inv.GetHelmet()),
		Chestplate:      convert.ItemStackWrapperLegacy(inv.GetChestplate()),
		Leggings:        convert.ItemStackWrapperLegacy(inv.GetLeggings()),
		Boots:           convert.ItemStackWrapperLegacy(inv.GetBoots()),
	})
}

func (b *StandardEntityEventBroadcaster) OnPickUpItem(recipients []*NetworkSession, collector, pickedUp entityWithID) {
	b.sendDataPacket(recipients, &packet.TakeItemActor{ItemEntityRuntimeID: uint64(pickedUp.GetID()), TakerEntityRuntimeID: uint64(collector.GetID())})
}

func (b *StandardEntityEventBroadcaster) OnEmote(recipients []*NetworkSession, from entityWithID, emoteID string) {
	b.sendDataPacket(recipients, &packet.Emote{
		EntityRuntimeID: uint64(from.GetID()),
		EmoteID:         emoteID,
		EmoteLength:     0, //seems to be irrelevant for the client, we cannot risk rebroadcasting random values received
		Flags:           packet.EmoteFlagServerSide | packet.EmoteFlagMuteChat,
	})
}
