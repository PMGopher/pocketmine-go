package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// DefaultCriticalHitParticleCount is CriticalHitAnimation's default $particleCount.
const DefaultCriticalHitParticleCount = 55

// CriticalHitAnimation is a port of pocketmine\entity\animation\CriticalHitAnimation. The entity is
// a Living.
type CriticalHitAnimation struct {
	Entity        Entity
	ParticleCount int
}

func (a CriticalHitAnimation) Encode() []packet.Packet {
	return []packet.Packet{&packet.Animate{
		ActionType:      packet.AnimateActionCriticalHit,
		EntityRuntimeID: uint64(a.Entity.GetID()),
		Data:            float32(a.ParticleCount),
	}}
}
