package entity

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/entity/effect"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world"
)

// This file ports the parts of pocketmine\network\mcpe\NetworkBroadcastUtils and
// StandardEntityEventBroadcaster that entities use to talk to their viewers. PHP routes these
// through each viewer's NetworkSession; this port has no NetworkSession yet, so packets go straight
// to world.EntityViewer.SendPacket.

// BroadcastPackets is a port of NetworkBroadcastUtils::broadcastPackets.
func BroadcastPackets(targets []world.EntityViewer, pks ...packet.Packet) {
	for _, target := range targets {
		for _, pk := range pks {
			target.SendPacket(pk)
		}
	}
}

func removeActorPacket(id int) packet.Packet {
	return &packet.RemoveActor{EntityUniqueID: int64(id)}
}

// SyncAttributesPacket is StandardEntityEventBroadcaster::syncAttributes' packet (nil if there's
// nothing to sync).
func SyncAttributesPacket(entityID int, attributes []*Attribute) packet.Packet {
	if len(attributes) == 0 {
		return nil
	}
	networkAttributes := make([]protocol.Attribute, 0, len(attributes))
	for _, attr := range attributes {
		networkAttributes = append(networkAttributes, protocol.Attribute{
			AttributeValue: protocol.AttributeValue{
				Name:  attr.GetID(),
				Value: float32(attr.GetValue()),
				Max:   float32(attr.GetMaxValue()),
				Min:   float32(attr.GetMinValue()),
			},
			DefaultMin: float32(attr.GetMinValue()),
			DefaultMax: float32(attr.GetMaxValue()),
			Default:    float32(attr.GetDefaultValue()),
		})
	}
	return &packet.UpdateAttributes{EntityRuntimeID: uint64(entityID), Attributes: networkAttributes, Tick: 0}
}

// EntityEffectAddedPacket is StandardEntityEventBroadcaster::onEntityEffectAdded's packet.
func EntityEffectAddedPacket(entityID int, instance *effect.EffectInstance, replacesOldEffect bool) packet.Packet {
	//TODO: we may need yet another effect <=> ID map in the future depending on protocol changes
	operation := byte(packet.MobEffectAdd)
	if replacesOldEffect {
		operation = packet.MobEffectModify
	}
	duration := int32(instance.GetDuration())
	if instance.IsInfinite() {
		duration = -1
	}
	return &packet.MobEffect{
		EntityRuntimeID: uint64(entityID),
		Operation:       operation,
		EffectType:      int32(bedrock.EffectIdMap().ToID(instance.GetType())),
		Amplifier:       int32(instance.GetAmplifier()),
		Particles:       instance.IsVisible(),
		Duration:        duration,
		Tick:            0,
		Ambient:         instance.IsAmbient(),
	}
}

// EntityEffectRemovedPacket is StandardEntityEventBroadcaster::onEntityEffectRemoved's packet.
func EntityEffectRemovedPacket(entityID int, instance *effect.EffectInstance) packet.Packet {
	return &packet.MobEffect{
		EntityRuntimeID: uint64(entityID),
		Operation:       packet.MobEffectRemove,
		EffectType:      int32(bedrock.EffectIdMap().ToID(instance.GetType())),
		Tick:            0,
	}
}

// mobArmorChangePacket is StandardEntityEventBroadcaster::onMobArmorChange's packet.
func mobArmorChangePacket(l *Living) packet.Packet {
	inv := l.GetArmorInventory()
	return &packet.MobArmourEquipment{
		EntityRuntimeID: uint64(l.GetID()),
		Helmet:          convert.ItemStackWrapperLegacy(inv.GetHelmet()),
		Chestplate:      convert.ItemStackWrapperLegacy(inv.GetChestplate()),
		Leggings:        convert.ItemStackWrapperLegacy(inv.GetLeggings()),
		Boots:           convert.ItemStackWrapperLegacy(inv.GetBoots()),
		Body:            protocol.ItemInstance{},
	}
}

// mobMainHandItemChangePacket is StandardEntityEventBroadcaster::onMobMainHandItemChange's packet.
func mobMainHandItemChangePacket(h *Human) packet.Packet {
	//TODO: we could send zero for slot here because remote players don't need to know which slot was selected
	inv := h.GetInventory()
	return &packet.MobEquipment{
		EntityRuntimeID: uint64(h.GetID()),
		NewItem:         convert.ItemStackWrapperLegacy(inv.GetItemInHand()),
		InventorySlot:   byte(inv.GetHeldItemIndex()),
		HotBarSlot:      byte(inv.GetHeldItemIndex()),
		WindowID:        protocol.WindowIDInventory,
	}
}

// mobOffHandItemChangePacket is StandardEntityEventBroadcaster::onMobOffHandItemChange's packet.
func mobOffHandItemChangePacket(h *Human) packet.Packet {
	return &packet.MobEquipment{
		EntityRuntimeID: uint64(h.GetID()),
		NewItem:         convert.ItemStackWrapperLegacy(h.GetOffHandInventory().GetItem(0)),
		InventorySlot:   0,
		HotBarSlot:      0,
		WindowID:        protocol.WindowIDOffHand,
	}
}

// BroadcastPickUpItem is StandardEntityEventBroadcaster::onPickUpItem: the "collector picked up
// pickedUp" animation.
func BroadcastPickUpItem(targets []world.EntityViewer, collectorID, pickedUpID int) {
	BroadcastPackets(targets, &packet.TakeItemActor{ItemEntityRuntimeID: uint64(pickedUpID), TakerEntityRuntimeID: uint64(collectorID)})
}
