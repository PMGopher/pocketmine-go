package block

import (
	"math/rand"

	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
	"pocketmine-go/pocketmine/utils"
)

// TreeType is pocketmine\world\generator\object\TreeType's case (object.TreeType's value): this
// package can't import world/generator/object (it imports this package).
type TreeType int

// TreeType cases, in object.TreeType's order.
const (
	TreeTypeOak TreeType = iota
	TreeTypeSpruce
	TreeTypeBirch
	TreeTypeJungle
	TreeTypeAcacia
	TreeTypeDarkOak
	TreeTypeCrimson
	TreeTypeWarped
	TreeTypeAzalea
)

// Growth hooks set by world/generator/object's init().
var (
	// TreeTransactionFunc is TreeFactory::get($random, $type)?->getBlockTransaction(...): nil if
	// the tree type has no tree or it can't grow there.
	TreeTransactionFunc func(treeType TreeType, world World, x, y, z int, random *utils.Random) *BlockTransactionImpl
	// GrowGrassFunc is TallGrass::growGrass.
	GrowGrassFunc func(world World, x, y, z int, random *utils.Random, count, radius int)
)

// treeTransaction calls TreeTransactionFunc if it's set.
func treeTransaction(treeType TreeType, world World, x, y, z int, random *utils.Random) *BlockTransactionImpl {
	if TreeTransactionFunc == nil {
		return nil
	}
	return TreeTransactionFunc(treeType, world, x, y, z, random)
}

// growStructure is the shared body of Sapling, Azalea and NetherFungus::grow: the tree grown at
// the block's position, after StructureGrowEvent.
func growStructure(source Behavior, treeType TreeType, player Player) bool {
	pos := source.GetPosition()
	world, err := pos.GetWorld()
	if err != nil {
		return false
	}
	random := utils.NewRandom(rand.Int())
	tx := treeTransaction(treeType, world, pos.FloorX(), pos.FloorY(), pos.FloorZ(), random)
	if tx == nil {
		return false
	}
	var eventPlayer blockevent.Player
	if player != nil {
		eventPlayer = player
	}
	ev := blockevent.NewStructureGrowEvent(source, tx, eventPlayer)
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}
	return tx.Apply()
}

// hasFiniteResources is `$player->hasFiniteResources()` (survival/adventure).
func hasFiniteResources(player Player) bool {
	finite, ok := player.(interface{ HasFiniteResources() bool })
	return ok && finite.HasFiniteResources()
}

// isFertilizer is `$item instanceof Fertilizer` (bone meal is the only Fertilizer).
func isFertilizer(item Item) bool { return item.GetTypeId() == itemTypeIDsBoneMeal }

// isHoe is `$item instanceof Hoe`: hoes are the only items with the hoe tool type.
func isHoe(item Item) bool { return item.GetBlockToolType() == ToolTypeHoe }

// isShovel is `$item instanceof Shovel`: shovels are the only items with the shovel tool type.
func isShovel(item Item) bool { return item.GetBlockToolType() == ToolTypeShovel }

// applyDamage is `$item->applyDamage($amount)` for a Durable item (no-op otherwise).
func applyDamage(item Item, amount int) {
	if durable, ok := item.(Durable); ok {
		durable.ApplyDamage(amount)
	}
}

// eventPlayer converts a (possibly nil) block Player to the event package's Player.
func eventPlayer(player Player) blockevent.Player {
	if player == nil {
		return nil
	}
	return player
}
