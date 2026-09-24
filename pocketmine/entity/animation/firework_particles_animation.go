package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// FireworkParticlesAnimation is a port of pocketmine\entity\animation\FireworkParticlesAnimation.
// The entity is a FireworkRocket.
type FireworkParticlesAnimation struct{ Entity Entity }

func (a FireworkParticlesAnimation) Encode() []packet.Packet {
	return actorEvent(a.Entity, actorEventFireworkParticles, 0)
}
