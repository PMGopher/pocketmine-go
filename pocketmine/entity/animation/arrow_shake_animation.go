package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// ArrowShakeAnimation is a port of pocketmine\entity\animation\ArrowShakeAnimation. The entity is
// an Arrow.
type ArrowShakeAnimation struct {
	Arrow           Entity
	DurationInTicks int
}

func (a ArrowShakeAnimation) Encode() []packet.Packet {
	return actorEvent(a.Arrow, actorEventArrowShake, int32(a.DurationInTicks))
}
