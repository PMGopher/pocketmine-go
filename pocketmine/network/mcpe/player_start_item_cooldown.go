package mcpe

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// PlayerStartItemCooldown is BedrockProtocol's PlayerStartItemCooldownPacket (0xb0), which
// gophertunnel doesn't define (it only has the serverbound ClientStartItemCooldown). It starts an
// item cooldown on the client, as NetworkSession::onItemCooldownChanged sends it.
type PlayerStartItemCooldown struct {
	ItemCategory  string
	CooldownTicks int32
}

// ID is ProtocolInfo::PLAYER_START_ITEM_COOLDOWN_PACKET.
func (*PlayerStartItemCooldown) ID() uint32 { return 0xb0 }

func (pk *PlayerStartItemCooldown) Marshal(io protocol.IO) {
	io.String(&pk.ItemCategory)
	io.Varint32(&pk.CooldownTicks)
}
