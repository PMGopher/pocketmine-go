package block

import (
	stdmath "math"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/sound"
)

const ItemFrameRotations = 8

// ItemFrame is a port of pocketmine\block\ItemFrame, minus actually inserting a newly held item
// into the frame: PHP's `$this->framedItem = $item->pop();` relies on Item::pop() returning the
// popped clone, but this port's local Item interface can't declare a method returning Item and
// have it satisfied by the real (future) item package automatically - a self-referential return
// type means "block.Item" and "item.Item" would be different named types, and Go doesn't do
// covariant interface satisfaction across packages (unlike the block registry gaps elsewhere,
// this one can't be fixed by simply widening the local interface). FramedItem can still be set
// directly (SetFramedItem), and everything that doesn't need to construct a new item from a held
// one - rotating, ejecting, dropping, picking - is fully real.
type ItemFrame struct {
	Flowable
	FacingComponent

	HasMapValue    bool
	FramedItem     Item
	ItemRotation   int
	ItemDropChance float64
}

func NewItemFrame(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ItemFrame {
	i := &ItemFrame{
		Flowable:        Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		FacingComponent: NewFacingComponent(),
		ItemDropChance:  1.0,
	}
	i.Init(i)
	return i
}

func (i *ItemFrame) Clone() Behavior {
	c := *i
	c.rebind(&c)
	return &c
}

func (i *ItemFrame) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.Facing(&i.Facing)
	w.Bool(&i.HasMapValue)
}

// ReadStateFromWorld is a port of ItemFrame::readStateFromWorld.
func (i *ItemFrame) ReadStateFromWorld() Behavior {
	i.Flowable.ReadStateFromWorld()
	world, err := i.position.GetWorld()
	if err != nil {
		return i.self
	}
	t, ok := world.GetTile(i.position)
	if !ok {
		return i.self
	}
	tileFrame, ok := t.(*tile.ItemFrame)
	if !ok {
		return i.self
	}
	i.FramedItem = nil
	if tileItem, has := tileFrame.GetItem(); has {
		if bi, ok := tileItem.(Item); ok {
			i.FramedItem = bi
		}
	}
	i.ItemRotation = tileFrame.GetItemRotation() % ItemFrameRotations
	i.ItemDropChance = tileFrame.GetItemDropChance()
	return i.self
}

// GetFramedItem returns the framed item directly (no defensive clone - block.Item has no Clone()
// method, see type doc comment for why one can't be added generically).
func (i *ItemFrame) GetFramedItem() Item { return i.FramedItem }

// SetFramedItem is a port of ItemFrame::setFramedItem, minus cloning the incoming item (see type
// doc comment).
func (i *ItemFrame) SetFramedItem(item Item) {
	if item == nil || item.IsNull() {
		i.FramedItem = nil
		i.ItemRotation = 0
		return
	}
	i.FramedItem = item
}

func (i *ItemFrame) GetItemRotation() int { return i.ItemRotation }

func (i *ItemFrame) SetItemRotation(rotation int) { i.ItemRotation = rotation }

func (i *ItemFrame) GetItemDropChance() float64 { return i.ItemDropChance }

// SetItemDropChance is a port of ItemFrame::setItemDropChance - panics outside 0-1, matching the
// PHP original's InvalidArgumentException (same convention as Hopper.SetFacing).
func (i *ItemFrame) SetItemDropChance(chance float64) {
	if chance < 0.0 || chance > 1.0 || stdmath.IsNaN(chance) || stdmath.IsInf(chance, 0) {
		panic("Drop chance must be in range 0-1")
	}
	i.ItemDropChance = chance
}

func (i *ItemFrame) HasMap() bool { return i.HasMapValue }

func (i *ItemFrame) SetHasMap(hasMap bool) { i.HasMapValue = hasMap }

func itemFrameCanBeSupportedAt(blk Behavior, face math.Facing) bool {
	bg, ok := blk.(blockGeometry)
	if !ok {
		return false
	}
	return bg.GetAdjacentSupportType(face) != blockutils.SupportTypeNone
}

// OnInteract is a port of ItemFrame::onInteract, minus the item-insertion branch (see type doc
// comment). Rotating an already-framed item is fully real.
func (i *ItemFrame) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if i.FramedItem != nil {
		i.ItemRotation = (i.ItemRotation + 1) % ItemFrameRotations
		i.addSound(sound.ItemFrameRotateItemSound{})
		i.setSelf()
	}
	return true
}

// OnAttack is a port of ItemFrame::onAttack, minus actually dropping the ejected item into the
// world (World.DropItem isn't in the ported World interface - same gap as SweetBerryBush's doc
// comment). The framed-item-clearing state change and the return value are both real.
func (i *ItemFrame) OnAttack(item Item, face math.Facing, player Player) bool {
	if i.FramedItem == nil {
		return false
	}
	if utils.GetRandomFloat() <= i.ItemDropChance {
		i.addSound(sound.ItemFrameRemoveItemSound{})
	}
	i.SetFramedItem(nil)
	i.setSelf()
	return true
}

// OnNearbyBlockChange is a port of ItemFrame::onNearbyBlockChange.
func (i *ItemFrame) OnNearbyBlockChange() {
	if !itemFrameCanBeSupportedAt(i.self, math.Opposite(i.Facing)) {
		world, err := i.position.GetWorld()
		if err != nil {
			return
		}
		world.UseBreakOn(i.position.Vector3)
	}
}

// Place is a port of ItemFrame::place.
func (i *ItemFrame) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if !itemFrameCanBeSupportedAt(blockReplace, math.Opposite(face)) {
		return false
	}
	i.Facing = face
	return i.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// GetDropsForCompatibleTool is a port of ItemFrame::getDropsForCompatibleTool.
func (i *ItemFrame) GetDropsForCompatibleTool(item Item) []Item {
	drops := i.Flowable.GetDropsForCompatibleTool(item)
	if i.FramedItem != nil && utils.GetRandomFloat() <= i.ItemDropChance {
		drops = append(drops, i.FramedItem)
	}
	return drops
}

// GetPickedItem is a port of ItemFrame::getPickedItem.
func (i *ItemFrame) GetPickedItem(addUserData bool) Item {
	if i.FramedItem != nil {
		return i.FramedItem
	}
	return i.Flowable.GetPickedItem(addUserData)
}

func (i *ItemFrame) addSound(s sound.Sound) {
	world, err := i.position.GetWorld()
	if err != nil {
		return
	}
	world.AddSound(i.position.Vector3, s)
}

func (i *ItemFrame) setSelf() {
	world, err := i.position.GetWorld()
	if err != nil {
		return
	}
	_ = world.SetBlock(i.position, i.self)
}
