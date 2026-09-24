package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// TotemUseAnimation is a port of pocketmine\entity\animation\TotemUseAnimation. The entity is a
// Human.
//
// TODO (from PHP): check if this can be expanded to more than just humans.
type TotemUseAnimation struct{ Human Entity }

func (a TotemUseAnimation) Encode() []packet.Packet {
	return actorEvent(a.Human, actorEventConsumeTotem, 0)
}
