package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// ShulkerBox is a port of pocketmine\block\ShulkerBox.
//
// WriteStateToWorld (pushing Facing back into the tile on placement) isn't ported - there's no
// WriteStateToWorld hook on Behavior yet, same documented gap as Note/Bed/BaseBanner/MobHead.
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

// OnInteract is a port of ShulkerBox::onInteract, minus actually opening the inventory window
// (player.SetCurrentWindow isn't ported - see block.Chest.OnInteract's doc comment for the same
// gap). The obstruction check (something solid blocking the box's opening side) and the
// CanOpenWith lock check are both fully real.
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
	// player.SetCurrentWindow(tileShulker.GetInventory()) - not ported, see doc comment above.
	return true
}

func (s *ShulkerBox) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

// GetDropsForCompatibleTool/GetPickedItem should copy the tile's cleaned NBT/name onto the
// dropped/picked item (addDataFromTile) - needs real Item construction and SetNamedTag/
// SetCustomName wiring from the unported item package integration (see
// Block.GetDropsForCompatibleTool's doc comment), so both are left as Block's defaults for now.
