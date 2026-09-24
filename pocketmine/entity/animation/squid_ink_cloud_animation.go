package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// SquidInkCloudAnimation is a port of pocketmine\entity\animation\SquidInkCloudAnimation. The
// entity is a Squid.
type SquidInkCloudAnimation struct{ Squid Entity }

func (a SquidInkCloudAnimation) Encode() []packet.Packet {
	return actorEvent(a.Squid, actorEventSquidInkCloud, 0)
}
