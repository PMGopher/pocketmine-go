package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Chest is a port of pocketmine\block\Chest.
type Chest struct {
	Transparent
	HorizontalFacingComponent
}

func NewChest(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Chest {
	c := &Chest{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	c.Init(c)
	return c
}

func (c *Chest) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Chest) DescribeBlockOnlyState(w runtime.DataDescriber) { c.DescribeHorizontalFacing(w) }

func (c *Chest) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().ContractedCopy(0.025, 0, 0.025).TrimmedCopy(math.Up, 0.05)}
}

func (c *Chest) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (c *Chest) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		c.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return c.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnPostPlace is a port of Chest::onPostPlace, minus firing ChestPairEvent - the event package
// isn't wired into the block package anywhere yet, so pairing proceeds unconditionally instead of
// being cancellable (matching the PHP original's behavior in the common case where nothing
// cancels the event).
func (c *Chest) OnPostPlace() {
	world, err := c.position.GetWorld()
	if err != nil {
		return
	}
	t, ok := world.GetTile(c.position)
	if !ok {
		return
	}
	tileChest, ok := t.(*tile.Chest)
	if !ok {
		return
	}

	for _, clockwise := range [2]bool{false, true} {
		side := math.RotateY(c.Facing, clockwise)
		neighbor := c.self.(blockGeometry).GetSide(side, 1)
		other, ok := neighbor.(*Chest)
		if !ok || !other.HasSameTypeId(c.self) || other.Facing != c.Facing {
			continue
		}
		pairTileRaw, ok := world.GetTile(other.position)
		if !ok {
			continue
		}
		pairTile, ok := pairTileRaw.(*tile.Chest)
		if !ok || pairTile.IsPaired() {
			continue
		}
		pairTile.PairWith(tileChest)
		break
	}
}

// OnInteract is a port of Chest::onInteract, minus actually opening the inventory window
// (player.SetCurrentWindow isn't ported - needs a real Player with window-management state). The
// support/lock checks that decide WHETHER it could be opened are fully real.
func (c *Chest) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player == nil {
		return true
	}
	world, err := c.position.GetWorld()
	if err != nil {
		return true
	}
	t, ok := world.GetTile(c.position)
	if !ok {
		return true
	}
	tileChest, ok := t.(*tile.Chest)
	if !ok {
		return true
	}

	if !c.self.(blockGeometry).GetSide(math.Up, 1).IsTransparent() {
		return true
	}
	if pair, ok := tileChest.GetPair(); ok {
		pairPos := pair.GetPosition()
		pairBlock := world.GetBlockAt(pairPos.FloorX(), pairPos.FloorY(), pairPos.FloorZ())
		if !pairBlock.(blockGeometry).GetSide(math.Up, 1).IsTransparent() {
			return true
		}
	}
	if !tileChest.CanOpenWith(item.GetCustomName()) {
		return true
	}

	// player.SetCurrentWindow(tileChest.GetInventory()) - not ported, see doc comment above.
	return true
}

func (c *Chest) GetFuelTime() int { return 300 }
