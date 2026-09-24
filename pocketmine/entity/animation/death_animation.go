package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// DeathAnimation is a port of pocketmine\entity\animation\DeathAnimation. The entity is a Living.
type DeathAnimation struct{ Entity Entity }

func (a DeathAnimation) Encode() []packet.Packet {
	return actorEvent(a.Entity, actorEventDeathAnimation, 0)
}
