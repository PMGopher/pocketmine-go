package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// ArmSwingAnimation is a port of pocketmine\entity\animation\ArmSwingAnimation. The entity is a
// Living.
//
// TODO (from PHP): not sure if this should be constrained to humanoids, but we don't have any
// concept of that right now.
type ArmSwingAnimation struct{ Entity Entity }

func (a ArmSwingAnimation) Encode() []packet.Packet {
	return actorEvent(a.Entity, actorEventArmSwing, 0)
}
