package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

var chiseledBookshelfAllSlots = [6]blockutils.ChiseledBookshelfSlot{
	blockutils.ChiseledBookshelfSlotTopLeft,
	blockutils.ChiseledBookshelfSlotTopMiddle,
	blockutils.ChiseledBookshelfSlotTopRight,
	blockutils.ChiseledBookshelfSlotBottomLeft,
	blockutils.ChiseledBookshelfSlotBottomMiddle,
	blockutils.ChiseledBookshelfSlotBottomRight,
}

// ChiseledBookshelf is a port of pocketmine\block\ChiseledBookshelf, minus its actual inventory
// interaction: OnInteract's real work (checking/clearing/filling a slot) needs a real tile-backed
// inventory, which doesn't exist here - see tile.ChiseledBookshelf's doc comment for the same
// import-cycle constraint as every other container tile. The displayed-slot bitset,
// lastInteractedSlot state, and placement facing are all real.
type ChiseledBookshelf struct {
	Opaque
	HorizontalFacingComponent

	Slots              map[blockutils.ChiseledBookshelfSlot]bool
	LastInteractedSlot *blockutils.ChiseledBookshelfSlot
}

func NewChiseledBookshelf(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ChiseledBookshelf {
	c := &ChiseledBookshelf{
		Opaque:                    Opaque{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	c.Init(c)
	return c
}

func (c *ChiseledBookshelf) Clone() Behavior {
	cl := *c
	if c.Slots != nil {
		cl.Slots = make(map[blockutils.ChiseledBookshelfSlot]bool, len(c.Slots))
		for k, v := range c.Slots {
			cl.Slots[k] = v
		}
	}
	if c.LastInteractedSlot != nil {
		slot := *c.LastInteractedSlot
		cl.LastInteractedSlot = &slot
	}
	cl.rebind(&cl)
	return &cl
}

// DescribeBlockOnlyState is a port of ChiseledBookshelf::describeBlockOnlyState's
// enumSet($this->slots, ...) - RuntimeDataDescriber.EnumSet isn't ported (same convention as
// BrewingStand.DescribeBlockOnlyState's doc comment).
func (c *ChiseledBookshelf) DescribeBlockOnlyState(w runtime.DataDescriber) {
	c.DescribeHorizontalFacing(w)
	for _, slot := range chiseledBookshelfAllSlots {
		occupied := c.HasSlot(slot)
		w.Bool(&occupied)
		c.SetSlot(slot, occupied)
	}
}

// ReadStateFromWorld is a port of ChiseledBookshelf::readStateFromWorld.
func (c *ChiseledBookshelf) ReadStateFromWorld() Behavior {
	world, err := c.position.GetWorld()
	if err != nil {
		return c.self
	}
	t, ok := world.GetTile(c.position)
	if !ok {
		c.LastInteractedSlot = nil
		return c.self
	}
	tileShelf, ok := t.(*tile.ChiseledBookshelf)
	if !ok {
		c.LastInteractedSlot = nil
		return c.self
	}
	if slot, has := tileShelf.GetLastInteractedSlot(); has {
		c.LastInteractedSlot = &slot
	} else {
		c.LastInteractedSlot = nil
	}
	return c.self
}

func (c *ChiseledBookshelf) HasSlot(slot blockutils.ChiseledBookshelfSlot) bool { return c.Slots[slot] }

func (c *ChiseledBookshelf) SetSlot(slot blockutils.ChiseledBookshelfSlot, occupied bool) {
	if occupied {
		if c.Slots == nil {
			c.Slots = map[blockutils.ChiseledBookshelfSlot]bool{}
		}
		c.Slots[slot] = true
	} else {
		delete(c.Slots, slot)
	}
}

func (c *ChiseledBookshelf) GetSlots() map[blockutils.ChiseledBookshelfSlot]bool { return c.Slots }

func (c *ChiseledBookshelf) SetSlots(slots map[blockutils.ChiseledBookshelfSlot]bool) {
	c.Slots = map[blockutils.ChiseledBookshelfSlot]bool{}
	for slot, occupied := range slots {
		if occupied {
			c.SetSlot(slot, true)
		}
	}
}

func (c *ChiseledBookshelf) GetLastInteractedSlot() *blockutils.ChiseledBookshelfSlot {
	return c.LastInteractedSlot
}

func (c *ChiseledBookshelf) SetLastInteractedSlot(slot *blockutils.ChiseledBookshelfSlot) {
	c.LastInteractedSlot = slot
}

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (c *ChiseledBookshelf) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		c.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return c.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// ChiseledBookshelfInteractFunc is the inventory part of ChiseledBookshelf::onInteract, set by
// block/inventory (the tile inventory and book item types can't be seen here): a book in the slot
// is taken out (returned), otherwise a held book (writable/written, plain or enchanted) is put in
// (stored). changed is false if nothing happened.
var ChiseledBookshelfInteractFunc func(inv tile.Inventory, slot int, held Item) (returned Item, stored, changed bool)

// OnInteract is a port of ChiseledBookshelf::onInteract.
func (c *ChiseledBookshelf) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if face != c.Facing {
		return false
	}

	x := clickVector.X
	if math.FacingAxis(face) == math.AxisX {
		x = clickVector.Z
	}
	if math.IsPositive(math.RotateY(face, true)) {
		x = 1 - x
	}
	slot := blockutils.ChiseledBookshelfSlotFromBlockFaceCoordinates(x, clickVector.Y)
	t, ok := c.tileAt()
	if !ok {
		return false
	}
	tileShelf, ok := t.(*tile.ChiseledBookshelf)
	if !ok || ChiseledBookshelfInteractFunc == nil {
		return false
	}

	returned, stored, changed := ChiseledBookshelfInteractFunc(tileShelf.GetInventory(), int(slot), item)
	if !changed {
		return true
	}
	if returned != nil && returnedItems != nil {
		*returnedItems = append(*returnedItems, returned)
	}
	c.SetSlot(slot, stored)
	c.LastInteractedSlot = &slot

	if world, err := c.position.GetWorld(); err == nil {
		_ = world.SetBlock(c.position, c.self)
	}
	return true
}

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap).
func (c *ChiseledBookshelf) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (c *ChiseledBookshelf) IsAffectedBySilkTouch() bool { return true }

// WriteStateToWorld is a port of ChiseledBookshelf::writeStateToWorld.
func (c *ChiseledBookshelf) WriteStateToWorld() {
	c.Block.WriteStateToWorld()
	if t, ok := c.tileAt(); ok {
		if tileShelf, ok := t.(*tile.ChiseledBookshelf); ok {
			tileShelf.SetLastInteractedSlot(c.LastInteractedSlot)
		}
	}
}
