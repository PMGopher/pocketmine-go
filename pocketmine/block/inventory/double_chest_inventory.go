package blockinventory

import (
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/world/sound"
)

// DoubleChestInventory is a port of pocketmine\block\inventory\DoubleChestInventory: a virtual
// 54-slot inventory delegating every operation to two real 27-slot ChestInventory halves.
//
// Unlike ChestInventory (which wraps AnimatedInventory/SimpleInventory directly), this composes
// its own inventory.BaseInventory since its storage isn't a flat array - GetItem/SetItem/etc. all
// route to whichever half a given slot index falls into. The open/close animate-on-transition
// logic is duplicated from AnimatedInventory rather than shared, mirroring the PHP original's own
// duplication-by-mixin (both ChestInventory and DoubleChestInventory `use
// AnimatedBlockInventoryTrait` independently).
type DoubleChestInventory struct {
	inventory.BaseInventory
	BlockInventoryTrait

	Left  *ChestInventory
	Right *ChestInventory
}

func NewDoubleChestInventory(left, right *ChestInventory) *DoubleChestInventory {
	d := &DoubleChestInventory{
		BlockInventoryTrait: BlockInventoryTrait{Holder: left.GetHolder()},
		Left:                left,
		Right:               right,
	}
	d.BaseInventory.Init(d)
	return d
}

func (d *DoubleChestInventory) GetSize() int { return d.Left.GetSize() + d.Right.GetSize() }

func (d *DoubleChestInventory) GetItem(index int) item.Item {
	if index < d.Left.GetSize() {
		return d.Left.GetItem(index)
	}
	return d.Right.GetItem(index - d.Left.GetSize())
}

func (d *DoubleChestInventory) InternalSetItem(index int, it item.Item) {
	if index < d.Left.GetSize() {
		d.Left.SetItem(index, it)
	} else {
		d.Right.SetItem(index-d.Left.GetSize(), it)
	}
}

func (d *DoubleChestInventory) GetContents(includeEmpty bool) map[int]item.Item {
	result := d.Left.GetContents(includeEmpty)
	leftSize := d.Left.GetSize()
	for i, it := range d.Right.GetContents(includeEmpty) {
		result[i+leftSize] = it
	}
	return result
}

func (d *DoubleChestInventory) InternalSetContents(items map[int]item.Item) {
	leftSize := d.Left.GetSize()
	leftContents := map[int]item.Item{}
	rightContents := map[int]item.Item{}
	for i, it := range items {
		if i < leftSize {
			leftContents[i] = it
		} else {
			rightContents[i-leftSize] = it
		}
	}
	d.Left.SetContents(leftContents)
	d.Right.SetContents(rightContents)
}

func (d *DoubleChestInventory) GetMatchingItemCount(slot int, test item.Item, checkTags bool) int {
	leftSize := d.Left.GetSize()
	if slot < leftSize {
		return d.Left.GetMatchingItemCount(slot, test, checkTags)
	}
	return d.Right.GetMatchingItemCount(slot-leftSize, test, checkTags)
}

func (d *DoubleChestInventory) IsSlotEmpty(index int) bool {
	leftSize := d.Left.GetSize()
	if index < leftSize {
		return d.Left.IsSlotEmpty(index)
	}
	return d.Right.IsSlotEmpty(index - leftSize)
}

func (d *DoubleChestInventory) GetViewerCount() int { return len(d.GetViewers()) }

// OnOpen is a port of AnimatedBlockInventoryTrait::onOpen as used by DoubleChestInventory.
func (d *DoubleChestInventory) OnOpen(who inventory.Player) {
	d.BaseInventory.OnOpen(who)

	if d.Holder.IsValid() && d.GetViewerCount() == 1 {
		d.animateBlock(true)
		if world, err := d.Holder.GetWorld(); err == nil {
			world.AddSound(d.Holder.Add(0.5, 0.5, 0.5), sound.ChestOpenSound{})
		}
	}
}

// OnClose is a port of AnimatedBlockInventoryTrait::onClose as used by DoubleChestInventory.
func (d *DoubleChestInventory) OnClose(who inventory.Player) {
	if d.Holder.IsValid() && d.GetViewerCount() == 1 {
		d.animateBlock(false)
		if world, err := d.Holder.GetWorld(); err == nil {
			world.AddSound(d.Holder.Add(0.5, 0.5, 0.5), sound.ChestCloseSound{})
		}
	}

	d.BaseInventory.OnClose(who)
}

func (d *DoubleChestInventory) animateBlock(isOpen bool) {
	d.Left.animateBlock(isOpen)
	d.Right.animateBlock(isOpen)
}

func (d *DoubleChestInventory) GetLeftSide() *ChestInventory { return d.Left }

func (d *DoubleChestInventory) GetRightSide() *ChestInventory { return d.Right }
