package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// BrewingStandSlot is a port of pocketmine\block\utils\BrewingStandSlot - the three visual bottle
// positions, each mapping to a real slot number in blockinventory.BrewingStandInventory's layout
// (confirmed against BrewingStandInventory.php's SLOT_BOTTLE_* constants, even though that
// inventory type itself isn't ported yet).
type BrewingStandSlot int

const (
	BrewingStandSlotEast BrewingStandSlot = iota
	BrewingStandSlotNorthwest
	BrewingStandSlotSouthwest
)

func (s BrewingStandSlot) GetSlotNumber() int {
	switch s {
	case BrewingStandSlotEast:
		return 1
	case BrewingStandSlotNorthwest:
		return 2
	default:
		return 3
	}
}

// BrewingStand is a port of pocketmine\block\BrewingStand.
type BrewingStand struct {
	Transparent

	Slots map[BrewingStandSlot]bool
}

func NewBrewingStand(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *BrewingStand {
	b := &BrewingStand{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}
	b.Init(b)
	return b
}

func (b *BrewingStand) Clone() Behavior {
	c := *b
	if b.Slots != nil {
		c.Slots = make(map[BrewingStandSlot]bool, len(b.Slots))
		for k, v := range b.Slots {
			c.Slots[k] = v
		}
	}
	c.rebind(&c)
	return &c
}

// DescribeBlockOnlyState is a port of BrewingStand::describeBlockOnlyState's enumSet($this->slots,
// BrewingStandSlot::cases()) - RuntimeDataDescriber.EnumSet isn't ported (only three cases exist,
// so this just describes each slot's occupied flag directly, the same convention used throughout
// this port for small fixed enum sets).
func (b *BrewingStand) DescribeBlockOnlyState(w runtime.DataDescriber) {
	for _, slot := range [3]BrewingStandSlot{BrewingStandSlotEast, BrewingStandSlotNorthwest, BrewingStandSlotSouthwest} {
		occupied := b.HasSlot(slot)
		w.Bool(&occupied)
		b.SetSlot(slot, occupied)
	}
}

func (b *BrewingStand) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{
		math.OneAABB().TrimmedCopy(math.Up, 7.0/8),
		math.OneAABB().SquashedCopy(math.AxisX, 7.0/16).SquashedCopy(math.AxisZ, 7.0/16).TrimmedCopy(math.Up, 1.0/8),
	}
}

func (b *BrewingStand) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (b *BrewingStand) HasSlot(slot BrewingStandSlot) bool { return b.Slots[slot] }

func (b *BrewingStand) SetSlot(slot BrewingStandSlot, occupied bool) {
	if occupied {
		if b.Slots == nil {
			b.Slots = map[BrewingStandSlot]bool{}
		}
		b.Slots[slot] = true
	} else {
		delete(b.Slots, slot)
	}
}

func (b *BrewingStand) GetSlots() map[BrewingStandSlot]bool { return b.Slots }

func (b *BrewingStand) SetSlots(slots map[BrewingStandSlot]bool) {
	b.Slots = map[BrewingStandSlot]bool{}
	for slot, occupied := range slots {
		if occupied {
			b.Slots[slot] = true
		}
	}
}

// OnInteract is a port of BrewingStand::onInteract, minus actually opening the inventory window
// (player.SetCurrentWindow isn't ported - see block.Chest.OnInteract's doc comment for the same
// gap). The CanOpenWith lock check that would gate it is fully real.
func (b *BrewingStand) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if player == nil {
		return true
	}
	world, err := b.position.GetWorld()
	if err != nil {
		return true
	}
	t, ok := world.GetTile(b.position)
	if !ok {
		return true
	}
	tileStand, ok := t.(*tile.BrewingStand)
	if !ok {
		return true
	}
	if tileStand.CanOpenWith(item.GetCustomName()) {
		// player.SetCurrentWindow(tileStand.GetInventory()) - not ported, see doc comment above.
	}
	return true
}

// OnScheduledUpdate should drive tile.BrewingStand.OnUpdate (not ported - see its doc comment)
// and then sync this block's visual Slots from the tile's real inventory occupancy - since the
// tile has no real inventory here (see ContainerComponent's doc comment), there's nothing to sync
// from, so this is a no-op.
func (b *BrewingStand) OnScheduledUpdate() {}
