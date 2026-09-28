package entity

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/world"
)

// init gives the item package Entity::getWorld (see item.EntityWorldFunc).
func init() {
	item.EntityWorldFunc = func(e item.Entity) block.World {
		if we, ok := e.(interface{ GetWorld() *world.World }); ok {
			if w := we.GetWorld(); w != nil {
				return w
			}
		}
		return nil
	}
}
