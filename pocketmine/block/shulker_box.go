package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// ShulkerBox is a port of pocketmine\block\ShulkerBox.
type ShulkerBox struct {
	Opaque
	FacingComponent
}

func NewShulkerBox(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ShulkerBox {
	s := &ShulkerBox{
		Opaque:          Opaque{NewBlock(idInfo, name, typeInfo)},
		FacingComponent: NewFacingComponent(),
	}
	s.Init(s)
	return s
}

func (s *ShulkerBox) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

// DescribeBlockOnlyState is a NOOP, matching the PHP original - facing isn't read/written as
// block state here because the tile persists it instead.
func (s *ShulkerBox) DescribeBlockOnlyState(w runtime.DataDescriber) {}

// ReadStateFromWorld is a port of ShulkerBox::readStateFromWorld.
func (s *ShulkerBox) ReadStateFromWorld() Behavior {
	s.Block.ReadStateFromWorld()
	world, err := s.position.GetWorld()
	if err != nil {
		return s.self
	}
	t, ok := world.GetTile(s.position)
	if !ok {
		return s.self
	}
	if tileShulker, ok := t.(*tile.ShulkerBox); ok {
		s.Facing = math.Facing(tileShulker.GetFacing())
	}
	return s.self
}

func (s *ShulkerBox) GetMaxStackSize() int { return 1 }

// Place is a port of ShulkerBox::place - faces directly towards the clicked face, unlike most
// other AnyFacing blocks.
func (s *ShulkerBox) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	s.Facing = face
	return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract is a port of ShulkerBox::onInteract.
func (s *ShulkerBox) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player == nil {
		return true
	}
	world, err := s.position.GetWorld()
	if err != nil {
		return true
	}
	t, ok := world.GetTile(s.position)
	if !ok {
		return true
	}
	tileShulker, ok := t.(*tile.ShulkerBox)
	if !ok {
		return true
	}
	if s.self.(blockGeometry).GetSide(s.Facing, 1).IsSolid() {
		return true
	}
	if !tileShulker.CanOpenWith(item.GetCustomName()) {
		return true
	}
	openTileWindow(player, tileShulker.GetInventory())
	return true
}

func (s *ShulkerBox) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

// addDataFromTile is a port of ShulkerBox::addDataFromTile: the box keeps its contents and name.
func (s *ShulkerBox) addDataFromTile(shulker *tile.ShulkerBox, it Item) {
	if tag := shulker.GetCleanedNBT(); tag != nil {
		if named, ok := it.(interface{ SetNamedTag(tag *nbt.CompoundTag) }); ok {
			named.SetNamedTag(tag)
		}
	}
	if shulker.HasName() {
		if named, ok := it.(interface{ SetCustomName(name string) }); ok {
			named.SetCustomName(shulker.GetName())
		}
	}
}

// GetDropsForCompatibleTool is a port of ShulkerBox::getDropsForCompatibleTool.
func (s *ShulkerBox) GetDropsForCompatibleTool(item Item) []Item {
	drop := asItemOrNil(s.self)
	if drop == nil {
		return nil
	}
	if t, ok := s.tileAt(); ok {
		if shulker, ok := t.(*tile.ShulkerBox); ok {
			s.addDataFromTile(shulker, drop)
		}
	}
	return []Item{drop}
}

// GetPickedItem is a port of ShulkerBox::getPickedItem.
func (s *ShulkerBox) GetPickedItem(addUserData bool) Item {
	result := s.Block.GetPickedItem(addUserData)
	if addUserData && result != nil {
		if t, ok := s.tileAt(); ok {
			if shulker, ok := t.(*tile.ShulkerBox); ok {
				s.addDataFromTile(shulker, result)
			}
		}
	}
	return result
}

// WriteStateToWorld is a port of ShulkerBox::writeStateToWorld.
func (s *ShulkerBox) WriteStateToWorld() {
	s.Block.WriteStateToWorld()
	if t, ok := s.tileAt(); ok {
		if shulker, ok := t.(*tile.ShulkerBox); ok {
			shulker.SetFacing(int(s.Facing))
		}
	}
}
