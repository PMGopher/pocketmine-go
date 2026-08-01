package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// EnderChest is a port of pocketmine\block\EnderChest.
type EnderChest struct {
	Transparent
	HorizontalFacingComponent
}

func NewEnderChest(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *EnderChest {
	e := &EnderChest{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	e.Init(e)
	return e
}

func (e *EnderChest) Clone() Behavior {
	c := *e
	c.rebind(&c)
	return &c
}

func (e *EnderChest) GetLightLevel() int { return 7 }

func (e *EnderChest) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().ContractedCopy(0.025, 0, 0.025).TrimmedCopy(math.Up, 0.05)}
}

func (e *EnderChest) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (e *EnderChest) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		e.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return e.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract is a port of EnderChest::onInteract, minus actually opening the inventory window
// (player.SetCurrentWindow and player.GetEnderInventory aren't ported - see
// block.Chest.OnInteract's doc comment for the same gap on the window side). The lid-transparency
// check and the ViewerCount increment it gates are both fully real.
func (e *EnderChest) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player == nil {
		return true
	}
	world, err := e.position.GetWorld()
	if err != nil {
		return true
	}
	t, ok := world.GetTile(e.position)
	if !ok {
		return true
	}
	tileEnderChest, ok := t.(*tile.EnderChest)
	if !ok {
		return true
	}
	if !e.self.(blockGeometry).GetSide(math.Up, 1).IsTransparent() {
		return true
	}
	tileEnderChest.SetViewerCount(tileEnderChest.GetViewerCount() + 1)
	// player.SetCurrentWindow(NewEnderChestInventory(e.position, player.GetEnderInventory())) -
	// not ported, see doc comment above.
	return true
}

// GetDropsForCompatibleTool should return [VanillaBlocks.OBSIDIAN().AsItem().SetCount(8)] - needs
// the unported block registry and real Item construction (see Block.GetDropsForCompatibleTool's
// doc comment), so it's left as Block's default for now.

func (e *EnderChest) IsAffectedBySilkTouch() bool { return true }
