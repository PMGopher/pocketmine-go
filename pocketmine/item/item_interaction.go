package item

import (
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

// Player is what the item interaction methods receive for pocketmine\player\Player. Items that
// need more (an inventory, the world, ...) type-assert to the interface they need, like PHP code
// narrowing with instanceof.
type Player = block.Player

// Entity is what the item interaction methods receive for pocketmine\entity\Entity.
type Entity = block.Entity

// interactions is the part of the Item interface for Player/Entity interactions (see Item).
type interactions interface {
	// GetBlock is a port of Item::getBlock: the block corresponding to this item (air for
	// non-block items).
	GetBlock() block.Behavior
	// GetBlockForFace is Item::getBlock($clickedFace): the block to place when clicking that
	// face (nil means unknown).
	GetBlockForFace(clickedFace *math.Facing) block.Behavior
	// CanBePlaced is a port of Item::canBePlaced.
	CanBePlaced() bool
	// GetPlacementTransaction is a port of Item::getPlacementTransaction: the blocks placing this
	// item changes, or nil if it can't be placed there. player may be nil.
	GetPlacementTransaction(blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, player Player) *block.BlockTransactionImpl
	// OnInteractBlock is a port of Item::onInteractBlock: called when a player uses this item on
	// a block.
	OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult
	// OnClickAir is a port of Item::onClickAir: called when a player uses the item on air, for
	// example throwing a projectile. directionVector is the direction the player is aiming.
	OnClickAir(player Player, directionVector math.Vector3, returnedItems *[]Item) ItemUseResult
	// OnReleaseUsing is a port of Item::onReleaseUsing: called when a player is using this item
	// and releases it. Used to handle bow shoot actions.
	OnReleaseUsing(player Player, returnedItems *[]Item) ItemUseResult
	// OnDestroyBlock is a port of Item::onDestroyBlock: called when this item is used to destroy
	// a block. Usually used to update durability. Returns whether the item changed.
	OnDestroyBlock(blk block.Behavior, returnedItems *[]Item) bool
	// OnAttackEntity is a port of Item::onAttackEntity: called when this item is used to attack
	// an entity. Usually used to update durability. Returns whether the item changed.
	OnAttackEntity(victim Entity, returnedItems *[]Item) bool
	// OnInteractEntity is a port of Item::onInteractEntity: called when a player uses the item to
	// interact with entity, for example by using a name tag. Returns whether the item changed.
	OnInteractEntity(player Player, entity Entity, clickVector math.Vector3) bool
}

// GetBlock is Item::getBlock's default: air.
func (b *ItemBase) GetBlock() block.Behavior { return block.VanillaAir() }

// GetBlockForFace is Item::getBlock($clickedFace)'s default: the face doesn't matter.
func (b *ItemBase) GetBlockForFace(clickedFace *math.Facing) block.Behavior {
	return b.self.GetBlock()
}

// CanBePlaced is a port of Item::canBePlaced.
func (b *ItemBase) CanBePlaced() bool { return b.self.GetBlock().CanBePlaced() }

// positionable is the part of block.Block tryPlacementTransaction needs.
type positionable interface {
	SetPosition(world block.World, x, y, z int)
}

// TryPlacementTransaction is a port of Item::tryPlacementTransaction.
func TryPlacementTransaction(self Item, blockPlace, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, player Player) *block.BlockTransactionImpl {
	position := blockReplace.GetPosition()
	world, err := position.GetWorld()
	if err != nil {
		return nil
	}
	if p, ok := blockPlace.(positionable); ok {
		p.SetPosition(world, position.FloorX(), position.FloorY(), position.FloorZ())
	}
	clicked := blockClicked.GetPosition()
	isClickedBlock := position.FloorX() == clicked.FloorX() && position.FloorY() == clicked.FloorY() && position.FloorZ() == clicked.FloorZ()
	if !blockPlace.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock) {
		return nil
	}
	transaction := block.NewBlockTransaction(world)
	if blockPlace.Place(transaction, self, blockReplace, blockClicked, face, clickVector, player) {
		return transaction
	}
	return nil
}

// GetPlacementTransaction is a port of Item::getPlacementTransaction.
func (b *ItemBase) GetPlacementTransaction(blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, player Player) *block.BlockTransactionImpl {
	return TryPlacementTransaction(b.self, b.self.GetBlockForFace(&face), blockReplace, blockClicked, face, clickVector, player)
}

func (b *ItemBase) OnInteractBlock(player Player, blockReplace, blockClicked block.Behavior, face math.Facing, clickVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	return ItemUseResultNone
}

func (b *ItemBase) OnClickAir(player Player, directionVector math.Vector3, returnedItems *[]Item) ItemUseResult {
	return ItemUseResultNone
}

func (b *ItemBase) OnReleaseUsing(player Player, returnedItems *[]Item) ItemUseResult {
	return ItemUseResultNone
}

func (b *ItemBase) OnDestroyBlock(blk block.Behavior, returnedItems *[]Item) bool { return false }

func (b *ItemBase) OnAttackEntity(victim Entity, returnedItems *[]Item) bool { return false }

func (b *ItemBase) OnInteractEntity(player Player, entity Entity, clickVector math.Vector3) bool {
	return false
}
