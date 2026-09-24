package animation

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// ItemEntityStackSizeChangeAnimation is a port of
// pocketmine\entity\animation\ItemEntityStackSizeChangeAnimation. The entity is an ItemEntity.
type ItemEntityStackSizeChangeAnimation struct {
	ItemEntity   Entity
	NewStackSize int
}

func (a ItemEntityStackSizeChangeAnimation) Encode() []packet.Packet {
	return actorEvent(a.ItemEntity, actorEventItemEntityMerge, int32(a.NewStackSize))
}
