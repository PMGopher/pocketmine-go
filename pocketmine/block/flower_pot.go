package block

import (
	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

// blockItemDropper is World::dropItem for blocks.
type blockItemDropper interface {
	DropBlockItem(source math.Vector3, it Item)
}

// dropItem is $world->dropItem($source, $item) for this package.
func dropItem(world World, source math.Vector3, it Item) {
	if it == nil {
		return
	}
	if dropper, ok := world.(blockItemDropper); ok {
		dropper.DropBlockItem(source, it)
	}
}

// FlowerPot is a port of pocketmine\block\FlowerPot.
//
// onInteract drops the removed plant instead of adding it to the player's inventory first
// (block.Player has no inventory).
type FlowerPot struct {
	Flowable

	plant Behavior
}

func NewFlowerPot(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *FlowerPot {
	f := &FlowerPot{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	f.Init(f)
	return f
}

func (f *FlowerPot) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

// GetPlant is a port of FlowerPot::getPlant.
func (f *FlowerPot) GetPlant() Behavior { return f.plant }

// SetPlant is a port of FlowerPot::setPlant.
func (f *FlowerPot) SetPlant(plant Behavior) {
	if plant == nil {
		f.plant = nil
		return
	}
	if _, isAir := plant.(*Air); isAir {
		f.plant = nil
		return
	}
	f.plant = plant.Clone()
}

// CanAddPlant is a port of FlowerPot::canAddPlant.
func (f *FlowerPot) CanAddPlant(blk Behavior) bool {
	if f.plant != nil {
		return false
	}
	return f.isValidPlant(blk)
}

func (f *FlowerPot) isValidPlant(blk Behavior) bool {
	return blk != nil && blk.(blockGeometry).HasTypeTag(BlockTypeTagsPottablePlants)
}

func (f *FlowerPot) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().ContractedCopy(3.0/16, 0, 3.0/16).TrimmedCopy(math.Up, 5.0/8)}
}

func (f *FlowerPot) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Down).HasCenterSupport()
}

func (f *FlowerPot) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return f.canBeSupportedAt(blockReplace) && f.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (f *FlowerPot) OnNearbyBlockChange() {
	if !f.canBeSupportedAt(f.self) {
		if world, err := f.position.GetWorld(); err == nil {
			world.UseBreakOn(f.position.AsVector3())
		}
	}
}

// OnInteract is a port of FlowerPot::onInteract.
func (f *FlowerPot) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	world, err := f.position.GetWorld()
	if err != nil {
		return false
	}
	var plant Behavior
	if ib, ok := item.(ItemBlockLike); ok {
		plant = ib.GetBlock()
	}
	if f.plant != nil {
		if f.isValidPlant(plant) {
			//for some reason, vanilla doesn't remove the contents of the pot if the held item is plantable
			//and will also cause a new plant to be placed if clicking on the side
			return false
		}

		if removed := asItemOrNil(f.plant); removed != nil {
			if dropper, ok := world.(blockItemDropper); ok {
				dropper.DropBlockItem(f.position.Add(0.5, 0.5, 0.5), removed)
			}
		}

		f.SetPlant(nil)
		_ = world.SetBlock(f.position, f)
		return true
	} else if f.isValidPlant(plant) {
		f.SetPlant(plant)
		item.Pop()
		_ = world.SetBlock(f.position, f)
		return true
	}
	return false
}

// GetDropsForCompatibleTool is a port of FlowerPot::getDropsForCompatibleTool.
func (f *FlowerPot) GetDropsForCompatibleTool(item Item) []Item {
	items := f.Flowable.GetDropsForCompatibleTool(item)
	if f.plant != nil {
		if plantItem := asItemOrNil(f.plant); plantItem != nil {
			items = append(items, plantItem)
		}
	}
	return items
}

// GetPickedItem is a port of FlowerPot::getPickedItem.
func (f *FlowerPot) GetPickedItem(addUserData bool) Item {
	if f.plant != nil {
		return asItemOrNil(f.plant)
	}
	return f.Flowable.GetPickedItem(addUserData)
}

// ReadStateFromWorld is a port of FlowerPot::readStateFromWorld.
func (f *FlowerPot) ReadStateFromWorld() Behavior {
	f.Block.ReadStateFromWorld()
	f.SetPlant(nil)
	if t, ok := f.tileAt(); ok {
		if potTile, ok := t.(*tile.FlowerPot); ok {
			if plant, ok := potTile.GetPlant().(Behavior); ok {
				f.SetPlant(plant)
			}
		}
	}
	return f.self
}

// WriteStateToWorld is a port of FlowerPot::writeStateToWorld.
func (f *FlowerPot) WriteStateToWorld() {
	f.Block.WriteStateToWorld()
	if t, ok := f.tileAt(); ok {
		if potTile, ok := t.(*tile.FlowerPot); ok {
			if f.plant == nil {
				potTile.SetPlant(nil)
			} else {
				potTile.SetPlant(f.plant.Clone())
			}
		}
	}
}
