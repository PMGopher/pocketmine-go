package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// RespawnAnimation is a port of pocketmine\entity\animation\RespawnAnimation. The entity is a
// Living.
type RespawnAnimation struct{ Entity Entity }

func (a RespawnAnimation) Encode() []packet.Packet {
	return actorEvent(a.Entity, actorEventRespawn, 0)
}
