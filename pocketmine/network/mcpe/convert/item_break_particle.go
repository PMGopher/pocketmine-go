package convert

import (
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/world/particle"
)

var particleItemTranslator = NewItemTranslator()

// NewItemBreakParticle is PHP's `new ItemBreakParticle($item)`: the network ID and meta
// ItemBreakParticle::encode gets from TypeConverter::getItemTranslator()->toNetworkId(). An item
// with no network mapping gets ID 0 (PHP's encode would throw).
func NewItemBreakParticle(it item.Item) particle.ItemBreakParticle {
	networkID, meta, _, _ := particleItemTranslator.ToNetworkIDQuiet(it)
	return particle.ItemBreakParticle{NetworkID: int(networkID), NetworkMeta: int(meta)}
}
