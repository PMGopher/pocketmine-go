// Package entity is a port of pocketmine\event\entity - every event fired by (or about) an entity.
//
// This package sits BELOW block, item, inventory, world and pocketmine/entity in the import graph:
// block constructs EntityDamageByBlockEvent/EntityCombustByBlockEvent/EntityTrampleFarmlandEvent,
// entity/effect constructs damage/heal events, and item's armor/enchantment code inspects
// EntityDamageEvent - so none of those packages can be imported from here. The PHP-typed payloads
// (Entity, Block, Item, Inventory, World) are therefore declared as small local interfaces below,
// the same forward-compatible-local-interface convention used throughout this port (see e.g.
// block.World). Listeners type-assert to the concrete type they expect (block.Behavior, item.Item,
// *entity.Living, ...) exactly where PHP code would narrow with instanceof.
//
// Importers conventionally alias this package as entityevent, since pocketmine/entity already owns
// the plain "entity" name.
package entity

import (
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
)

// Entity is the surface every event in this package needs from pocketmine\entity\Entity. GetID is
// what EntityDamageByEntityEvent/EntityDamageByChildEntityEvent store (PHP keeps the runtime ID
// and resolves it through WorldManager::findEntity); IsClosed is what that lookup's "entity no
// longer exists" null result maps to here.
type Entity interface {
	GetID() int
	GetPosition() math.Vector3
	IsClosed() bool
}

// Block is the surface these events need from pocketmine\block\Block. block.Behavior satisfies it.
type Block interface {
	GetName() string
	GetTypeId() int
}

// Item is the surface these events need from pocketmine\item\Item. item.Item satisfies it.
type Item interface {
	GetTypeId() int
	GetCount() int
	IsNull() bool
}

// Inventory is the surface EntityItemPickupEvent needs from pocketmine\inventory\Inventory.
// inventory.Inventory satisfies it.
type Inventory interface {
	GetSize() int
}

// World is the surface Position needs from pocketmine\world\World. *world.World satisfies it.
type World interface {
	GetDisplayName() string
}

// Position is a port of the pocketmine\world\Position values carried by EntityTeleportEvent and
// EntityExplodeEvent - a Vector3 plus the World it's in.
type Position struct {
	math.Vector3
	World World
}

// CloneItem is how events that PHP protects with `clone $item` (EntityItemPickupEvent) copy an
// Item without importing the item package. The item package's init() replaces this identity
// default with a real Item.Clone call - the same dependency-inversion hook as
// block.NewItemBlockFunc.
var CloneItem = func(it Item) Item { return it }

// EntityEvent is a port of pocketmine\event\entity\EntityEvent - embedded by every concrete event
// type in this package.
type EntityEvent struct {
	entity Entity
}

// GetEntity is a port of EntityEvent::getEntity.
func (e *EntityEvent) GetEntity() Entity { return e.entity }

// Call dispatches e to registered handlers on the global event Manager - a thin re-export of
// event.Call so callers don't need to import the event package separately. Equivalent to PHP's
// $event->call().
func Call[E any](e *E) { event.Call(e) }

// HasHandlers is a port of Event::hasHandlers.
func HasHandlers[E any]() bool { return event.HasHandlers[E]() }

// NewEntityEventBase builds the embedded EntityEvent for event types declared outside this package
// (pocketmine/event/player's PlayerExhaustEvent, entity/effect's effect events) - PHP's
// `$this->entity = $entity` assignment from a subclass constructor.
func NewEntityEventBase(entity Entity) EntityEvent { return EntityEvent{entity: entity} }
