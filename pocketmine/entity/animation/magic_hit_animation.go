package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// DefaultMagicHitParticleCount is MagicHitAnimation's default $particleCount.
const DefaultMagicHitParticleCount = 15

// MagicHitAnimation is a port of pocketmine\entity\animation\MagicHitAnimation. The entity is a
// Living.
type MagicHitAnimation struct {
	Entity        Entity
	ParticleCount int
}

func (a MagicHitAnimation) Encode() []packet.Packet {
	return []packet.Packet{&packet.Animate{
		ActionType:      packet.AnimateActionMagicCriticalHit,
		EntityRuntimeID: uint64(a.Entity.GetID()),
		Data:            float32(a.ParticleCount),
	}}
}
