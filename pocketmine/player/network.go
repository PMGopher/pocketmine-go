package player

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/math"
)

// levelEvent builds a real LevelEventPacket - a small shared helper for the handful of call sites
// in this package (SurvivalBlockBreakHandler's break-progress broadcasts) that need to construct
// one directly rather than going through a world/sound.Sound or world/particle.Particle value.
func levelEvent(eventType int32, eventData int32, pos math.Vector3) packet.Packet {
	return &packet.LevelEvent{
		EventType: eventType,
		Position:  mgl32.Vec3{float32(pos.X), float32(pos.Y), float32(pos.Z)},
		EventData: eventData,
	}
}

// PacketSender is the real delivery mechanism SendPacket uses - a thin closure around whatever
// connection cmd/pocketmine-go actually holds for this player (see SetPacketSender). Real PHP
// reaches this player's own NetworkSession::sendDataPacket; this port has no NetworkSession type,
// so the caller that owns the actual gophertunnel *minecraft.Conn supplies this directly instead.
type PacketSender func(pk packet.Packet)

// SetPacketSender wires this player up to an actual network connection - called once by
// cmd/pocketmine-go right after constructing the Player for a real session. Until this is called,
// SendPacket is a harmless no-op (matches how a Player isn't reachable over the network until a
// session exists for it in real PHP either).
func (p *Player) SetPacketSender(sender PacketSender) { p.packetSender = sender }

// SendPacket is a port of the network-facing half of calls like
// NetworkSession::sendDataPacket/World::broadcastPacketToViewers's per-viewer dispatch - the thing
// that actually receives a packet.Packet a *Player is a viewer/target of (see world.viewer, which
// this method exists to satisfy).
func (p *Player) SendPacket(pk packet.Packet) {
	if p.packetSender != nil {
		p.packetSender(pk)
	}
}
