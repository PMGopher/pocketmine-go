// Package block is a port of pocketmine\event\block: events fired by (or about) blocks.
//
// Like pocketmine/event/entity, this package sits below block, item, world and player in the
// import graph (block fires most of these events), so the PHP-typed payloads are small local
// interfaces: block.Behavior satisfies Block, item.Item satisfies Item, *player.Player satisfies
// Player, and so on. Listeners type-assert to the concrete type they need, exactly where PHP code
// would narrow with instanceof.
//
// Importers conventionally alias this package as blockevent.
package block

import (
	stdmath "math"

	"pocketmine-go/pocketmine/event"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/math"
)

// Block is the surface these events need from pocketmine\block\Block.
type Block interface {
	GetName() string
	GetTypeId() int
}

// Item is the surface these events need from pocketmine\item\Item.
type Item = entityevent.Item

// Entity is the surface these events need from pocketmine\entity\Entity.
type Entity = entityevent.Entity

// Inventory is the surface these events need from pocketmine\inventory\Inventory.
type Inventory = entityevent.Inventory

// Player is the surface these events need from pocketmine\player\Player (block.Player and
// *player.Player satisfy it).
type Player interface {
	Entity
}

// BlockTransaction is pocketmine\world\BlockTransaction (block.BlockTransactionImpl). Listeners
// type-assert it to the concrete transaction type.
type BlockTransaction interface{}

// Position is pocketmine\world\Position.
type Position = entityevent.Position

// BlockEvent is a port of pocketmine\event\block\BlockEvent.
type BlockEvent struct {
	block Block
}

// NewBlockEventBase builds the embedded BlockEvent (PHP's parent::__construct($block)).
func NewBlockEventBase(block Block) BlockEvent { return BlockEvent{block: block} }

// GetBlock is a port of BlockEvent::getBlock.
func (e *BlockEvent) GetBlock() Block { return e.block }

// Call dispatches e to registered handlers on the global event Manager.
func Call[E any](e *E) { event.Call(e) }

// checkFinite is Utils::checkFloatNotInfOrNaN.
func checkFinite(name string, v float64) {
	if stdmath.IsNaN(v) || stdmath.IsInf(v, 0) {
		panic(name + " cannot be NaN or Inf")
	}
}

// checkVector3 is Utils::checkVector3NotInfOrNaN.
func checkVector3(v math.Vector3) {
	checkFinite("x", v.X)
	checkFinite("y", v.Y)
	checkFinite("z", v.Z)
}
