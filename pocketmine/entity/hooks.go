package entity

import (
	"pocketmine-go/pocketmine/entity/effect"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/world"
)

// init installs this package's side of the dependency-inversion hooks lower packages declare (see
// each hook's own doc comment for why it exists).
func init() {
	effect.HumanHungerManager = func(l effect.Living) (effect.HungerManager, bool) {
		if h, ok := l.(interface{ GetHungerManager() *HungerManager }); ok {
			return h.GetHungerManager(), true
		}
		return nil, false
	}
	effect.OwningEntityOf = func(source entityevent.Entity) entityevent.Entity {
		if owned, ok := source.(interface{ GetOwningEntity() world.Entity }); ok {
			if owner := owned.GetOwningEntity(); owner != nil {
				return owner
			}
		}
		return nil
	}
}
