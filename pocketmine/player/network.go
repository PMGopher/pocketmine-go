package player

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// SendPacket sends pk to this player's client (world.EntityViewer): what PHP does with
// $player->getNetworkSession()->sendDataPacket() when broadcasting to viewers.
func (p *Player) SendPacket(pk packet.Packet) {
	if p.networkSession != nil {
		p.networkSession.SendDataPacket(pk)
	}
}

// GetYaw is Player's shorthand for getLocation()->getYaw() (block.Player needs it).
func (p *Player) GetYaw() float64 { return p.GetLocation().Yaw }

// GetPitch is Player's shorthand for getLocation()->getPitch().
func (p *Player) GetPitch() float64 { return p.GetLocation().Pitch }
