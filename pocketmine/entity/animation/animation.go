// Package animation is a port of pocketmine\entity\animation - one-shot visual effects broadcast to
// an entity's viewers (Entity.BroadcastAnimation).
//
// Every animation only needs its subject's runtime ID, so the PHP-typed constructor parameters
// (Living, Human, Arrow, Squid, ItemEntity, FireworkRocket) are all the local Entity interface
// here - this package sits below pocketmine/entity, which uses it.
package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// Animation is a port of pocketmine\entity\animation\Animation. Represents an animation such as an
// arm swing, or other visual effect done by entities.
type Animation interface {
	Encode() []packet.Packet
}

// Entity is the surface every animation needs from its subject.
type Entity interface {
	GetID() int
}

// ActorEvent IDs used by the animations in this package, named after
// pocketmine\network\mcpe\protocol\types\ActorEvent (BedrockProtocol). gophertunnel exposes the same
// wire values under different names; the equivalents are noted for cross-reference.
const (
	actorEventHurtAnimation     = 2  // packet.ActorEventHurt
	actorEventDeathAnimation    = 3  // packet.ActorEventDeath
	actorEventArmSwing          = 4  // packet.ActorEventStartAttacking
	actorEventSquidInkCloud     = 15 // packet.ActorEventSquidFleeing
	actorEventRespawn           = 18 // packet.ActorEventSpawnAlive
	actorEventFireworkParticles = 25 // packet.ActorEventFireworksExplode
	actorEventArrowShake        = 39 // packet.ActorEventShake
	actorEventEatingItem        = 57 // packet.ActorEventFeed
	actorEventConsumeTotem      = 65 // packet.ActorEventTalismanActivate
	actorEventItemEntityMerge   = 69 // packet.ActorEventUpdateStackSize
)

// actorEvent is ActorEventPacket::create($actorRuntimeId, $eventId, $eventData, null).
func actorEvent(e Entity, eventID byte, eventData int32) []packet.Packet {
	return []packet.Packet{&packet.ActorEvent{EntityRuntimeID: uint64(e.GetID()), EventType: eventID, EventData: eventData}}
}
