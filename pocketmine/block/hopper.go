package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Hopper is a port of pocketmine\block\Hopper. Redstone-powered sucking/pushing logic is marked
// //TODO even in the PHP original, so there's nothing to port for OnScheduledUpdate.
type Hopper struct {
	Transparent
	PoweredByRedstoneComponent

	Facing math.Facing
}

func NewHopper(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Hopper {
	h := &Hopper{
		Transparent: Transparent{NewBlock(idInfo, name, typeInfo)},
		Facing:      math.Down,
	}
	h.Init(h)
	return h
}

func (h *Hopper) Clone() Behavior {
	c := *h
	c.rebind(&c)
	return &c
}

func (h *Hopper) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.FacingExcept(&h.Facing, math.Up)
	w.Bool(&h.Powered)
}

func (h *Hopper) GetFacing() math.Facing { return h.Facing }

// SetFacing panics if facing is Up, mirroring the PHP original's InvalidArgumentException (a
// programmer error at the call site) - hoppers can't face upward.
func (h *Hopper) SetFacing(facing math.Facing) {
	if facing == math.Up {
		panic("Hopper may not face upward")
	}
	h.Facing = facing
}

func (h *Hopper) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	result := []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 6.0/16)}
	for _, f := range math.HorizontalFacing {
		result = append(result, math.OneAABB().TrimmedCopy(f, 14.0/16))
	}
	return result
}

func (h *Hopper) GetSupportType(facing math.Facing) blockutils.SupportType {
	switch facing {
	case math.Up:
		return blockutils.SupportTypeFull
	case math.Down:
		if h.Facing == math.Down {
			return blockutils.SupportTypeCenter
		}
		return blockutils.SupportTypeNone
	default:
		return blockutils.SupportTypeNone
	}
}

// Place is a port of Hopper::place.
func (h *Hopper) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if face == math.Down {
		h.Facing = math.Down
	} else {
		h.Facing = math.Opposite(face)
	}
	return h.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract is a port of Hopper::onInteract, minus actually opening the inventory window
// (player.SetCurrentWindow isn't ported - see block.Chest.OnInteract's doc comment for the same
// gap). Notably (matching the PHP original) this returns false rather than true when there's no
// player, unlike most other container blocks' OnInteract.
func (h *Hopper) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player == nil {
		return false
	}
	world, err := h.position.GetWorld()
	if err != nil {
		return true
	}
	if t, ok := world.GetTile(h.position); ok {
		if _, ok := t.(*tile.Hopper); ok {
			// player.SetCurrentWindow(tileHopper.GetInventory()) - not ported, see doc comment above.
		}
	}
	return true
}

func (h *Hopper) OnScheduledUpdate() {}
