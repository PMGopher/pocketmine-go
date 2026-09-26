package blockinventory

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

// packetBroadcaster is World::broadcastPacketToViewers.
type packetBroadcaster interface {
	BroadcastPacketToViewers(pos math.Vector3, pk packet.Packet)
}

// broadcastChestEvent is what ChestInventory, EnderChestInventory and ShulkerBoxInventory's
// animateBlock do: a BlockEventPacket opening or closing the lid (event ID is always 1 for a chest).
func broadcastChestEvent(holder block.Position, isOpen bool) {
	world, err := holder.GetWorld()
	if err != nil {
		return
	}
	broadcaster, ok := world.(packetBroadcaster)
	if !ok {
		return
	}
	data := int32(0)
	if isOpen {
		data = 1
	}
	broadcaster.BroadcastPacketToViewers(holder.Vector3, &packet.BlockEvent{
		Position:  protocol.BlockPos{int32(holder.FloorX()), int32(holder.FloorY()), int32(holder.FloorZ())},
		EventType: 1,
		EventData: data,
	})
}
