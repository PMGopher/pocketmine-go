package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// BrewingStandSlot is a port of pocketmine\block\utils\BrewingStandSlot - the three visual bottle
// positions, each mapping to a real slot number in blockinventory.BrewingStandInventory's layout
// (BrewingStandInventory's SLOT_BOTTLE_* constants).
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

// OnInteract is a port of BrewingStand::onInteract.
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
		openTileWindow(player, tileStand.GetInventory())
	}
	return true
}

// OnScheduledUpdate is a port of BrewingStand::onScheduledUpdate: brews, then shows the bottles
// that are in the stand.
func (b *BrewingStand) OnScheduledUpdate() {
	world, err := b.position.GetWorld()
	if err != nil {
		return
	}
	t, _ := world.GetTile(b.position)
	brewing, ok := t.(*tile.BrewingStand)
	if !ok {
		return
	}
	if brewing.OnUpdate() {
		world.ScheduleDelayedBlockUpdate(b.position.Vector3, 1)
	}
	if tile.InventoryIsSlotEmptyFunc == nil {
		return
	}
	changed := false
	for _, slot := range []BrewingStandSlot{BrewingStandSlotEast, BrewingStandSlotNorthwest, BrewingStandSlotSouthwest} {
		occupied := !tile.InventoryIsSlotEmptyFunc(brewing.GetInventory(), slot.GetSlotNumber())
		if occupied != b.HasSlot(slot) {
			b.SetSlot(slot, occupied)
			changed = true
		}
	}
	if changed {
		_ = world.SetBlock(b.position, b.self)
	}
}
