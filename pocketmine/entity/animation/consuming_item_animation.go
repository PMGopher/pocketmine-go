package animation

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/network/mcpe/convert"
)

// itemTranslator is shared: convert.ItemTranslator is stateless (a lookup over vendored data).
var itemTranslator = convert.NewItemTranslator()

// ConsumingItemAnimation is a port of pocketmine\entity\animation\ConsumingItemAnimation. The
// entity is a Living.
type ConsumingItemAnimation struct {
	Entity Entity
	Item   item.Item
}

// Encode is a port of ConsumingItemAnimation::encode. An item with no network mapping encodes as ID
// 0 (PHP would throw from toNetworkId; an unmapped item can't be eaten client-side anyway).
func (a ConsumingItemAnimation) Encode() []packet.Packet {
	netID, netData, _, _ := itemTranslator.ToNetworkID(a.Item)
	//TODO: need to check the data values
	return actorEvent(a.Entity, actorEventEatingItem, (netID<<16)|int32(netData))
}
