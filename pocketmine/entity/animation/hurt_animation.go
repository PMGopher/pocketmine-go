package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// HurtAnimation is a port of pocketmine\entity\animation\HurtAnimation. The entity is a Living.
type HurtAnimation struct{ Entity Entity }

func (a HurtAnimation) Encode() []packet.Packet {
	return actorEvent(a.Entity, actorEventHurtAnimation, 0)
}
