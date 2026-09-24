package item

import (
	"pocketmine-go/pocketmine/entity/effect"
	entityevent "pocketmine-go/pocketmine/event/entity"
)

// Living is the surface item hooks (OnTickWorn, Consumable.OnConsume) need from
// pocketmine\entity\Living - declared locally because pocketmine/entity imports this package (a
// Living wears items, a Human holds an inventory of them). *entity.Living satisfies it.
type Living interface {
	effect.Living

	IsUnderwater() bool
}

// isHuman reports whether l is a pocketmine\entity\Human (see effect.HumanHungerManager).
func isHuman(l effect.Living) bool {
	if effect.HumanHungerManager == nil {
		return false
	}
	_, ok := effect.HumanHungerManager(l)
	return ok
}

func init() {
	// EntityItemPickupEvent protects its item with `clone`, which the event package can only do
	// through this hook - see entityevent.CloneItem.
	entityevent.CloneItem = func(it entityevent.Item) entityevent.Item {
		if i, ok := it.(Item); ok {
			return i.Clone()
		}
		return it
	}
}
