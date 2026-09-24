package entity

import "pocketmine-go/pocketmine/event"

// EntityFrostWalkerEvent is a port of pocketmine\event\entity\EntityFrostWalkerEvent - called when
// an entity with the Frost Walker enchantment converts water into frosted ice. PHP types the
// entity as Living and the liquid as Liquid.
type EntityFrostWalkerEvent struct {
	EntityEvent
	event.CancellableTrait

	radius      int
	liquid      Block
	targetBlock Block
}

func NewEntityFrostWalkerEvent(entity Entity, radius int, liquid, targetBlock Block) *EntityFrostWalkerEvent {
	return &EntityFrostWalkerEvent{EntityEvent: EntityEvent{entity: entity}, radius: radius, liquid: liquid, targetBlock: targetBlock}
}

func (e *EntityFrostWalkerEvent) Call() { event.Call(e) }

func (e *EntityFrostWalkerEvent) GetRadius() int { return e.radius }

func (e *EntityFrostWalkerEvent) SetRadius(radius int) { e.radius = radius }

func (e *EntityFrostWalkerEvent) GetLiquid() Block { return e.liquid }

func (e *EntityFrostWalkerEvent) SetLiquid(liquid Block) { e.liquid = liquid }

func (e *EntityFrostWalkerEvent) GetTargetBlock() Block { return e.targetBlock }

func (e *EntityFrostWalkerEvent) SetTargetBlock(targetBlock Block) { e.targetBlock = targetBlock }
