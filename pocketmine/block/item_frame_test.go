package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

func newTestItemFrame(w World) *ItemFrame {
	i := NewItemFrame(mustBlockIdentifier(1082), "Test Item Frame", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	i.SetPosition(w, 1, 2, 3)
	return i
}

// dualItem satisfies both block.Item and tile.Item, for exercising the block<->tile bridging in
// ItemFrame.ReadStateFromWorld.
type dualItem struct {
	fakeItem
}

func (dualItem) GetCustomBlockData() (*nbt.CompoundTag, bool) { return nil, false }
func (dualItem) GetNamedTag() *nbt.CompoundTag                { return nbt.NewCompoundTag() }
func (dualItem) HasCustomName() bool                          { return false }

// supportedNeighbor/unsupportedNeighbor are minimal Behaviors distinguished only by GetSupportType,
// for exercising itemFrameCanBeSupportedAt.
type supportlessBlock struct{ Block }

func newSupportlessBlock() *supportlessBlock {
	idInfo, err := NewBlockIdentifier(1083, nil)
	if err != nil {
		panic(err)
	}
	b := &supportlessBlock{Block: NewBlock(idInfo, "No Support", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))}
	b.Init(b)
	return b
}

func (b *supportlessBlock) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *supportlessBlock) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func TestItemFrameOnInteractRotatesFramedItem(t *testing.T) {
	w := &fakeWorld{}
	i := newTestItemFrame(w)
	i.FramedItem = fakeItem{}
	i.ItemRotation = 3

	if !i.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to return true")
	}
	if i.ItemRotation != 4 {
		t.Errorf("ItemRotation = %d, want 4", i.ItemRotation)
	}
	if w.lastSetBlock == nil {
		t.Error("expected the rotated item frame to be written back to the world")
	}
}

func TestItemFrameOnInteractRotationWrapsAround(t *testing.T) {
	w := &fakeWorld{}
	i := newTestItemFrame(w)
	i.FramedItem = fakeItem{}
	i.ItemRotation = 7

	i.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil)

	if i.ItemRotation != 0 {
		t.Errorf("ItemRotation = %d, want 0 (wrapped around)", i.ItemRotation)
	}
}

func TestItemFrameOnInteractDoesNothingWhenEmpty(t *testing.T) {
	w := &fakeWorld{}
	i := newTestItemFrame(w)

	if !i.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to return true even when empty")
	}
	if w.lastSetBlock != nil {
		t.Error("expected no state change when the frame is empty (insertion isn't ported)")
	}
}

func TestItemFrameOnAttackEjectsFramedItem(t *testing.T) {
	w := &fakeWorld{}
	i := newTestItemFrame(w)
	i.FramedItem = fakeItem{}
	i.ItemRotation = 5

	if !i.OnAttack(fakeItem{}, math.Up, nil) {
		t.Fatal("expected OnAttack to return true")
	}
	if i.FramedItem != nil {
		t.Error("expected FramedItem to be cleared")
	}
	if i.ItemRotation != 0 {
		t.Errorf("ItemRotation = %d, want 0 after clearing", i.ItemRotation)
	}
}

func TestItemFrameOnAttackReturnsFalseWhenEmpty(t *testing.T) {
	w := &fakeWorld{}
	i := newTestItemFrame(w)

	if i.OnAttack(fakeItem{}, math.Up, nil) {
		t.Error("expected OnAttack to return false when the frame is already empty")
	}
}

func TestItemFrameSetItemDropChancePanicsOutOfRange(t *testing.T) {
	w := &fakeWorld{}
	i := newTestItemFrame(w)
	defer func() {
		if recover() == nil {
			t.Error("expected SetItemDropChance(1.5) to panic")
		}
	}()
	i.SetItemDropChance(1.5)
}

func TestItemFrameHasMapGetSet(t *testing.T) {
	w := &fakeWorld{}
	i := newTestItemFrame(w)
	if i.HasMap() {
		t.Error("expected HasMap() to default to false")
	}
	i.SetHasMap(true)
	if !i.HasMap() {
		t.Error("expected HasMap() to be true after SetHasMap(true)")
	}
}

func TestItemFramePlaceSucceedsWhenSupported(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	i := newTestItemFrame(w)
	tx := &fakeBlockTransaction{}

	blockReplace := newTestBlock(true)
	blockReplace.SetPosition(w, 1, 2, 3)
	// The default filler at (1,3,3) is a full-support testBlock, so the side opposite Down (Up) is
	// supported.

	if !i.Place(tx, fakeItem{}, blockReplace, blockReplace, math.Down, math.Vector3{}, nil) {
		t.Error("expected Place to succeed when the opposite side is supported")
	}
	if i.Facing != math.Down {
		t.Errorf("Facing = %v, want Down (the clicked face)", i.Facing)
	}
}

func TestItemFramePlaceFailsWhenUnsupported(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	i := newTestItemFrame(w)
	tx := &fakeBlockTransaction{}

	blockReplace := newTestBlock(true)
	blockReplace.SetPosition(w, 1, 2, 3)
	w.blocks[[3]int{1, 2, 3}] = blockReplace

	// itemFrameCanBeSupportedAt checks the side opposite the clicked face (Down -> Up), so the
	// block above blockReplace is what needs to report no support.
	unsupported := newSupportlessBlock()
	unsupported.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = unsupported

	if i.Place(tx, fakeItem{}, blockReplace, blockReplace, math.Down, math.Vector3{}, nil) {
		t.Error("expected Place to fail when the block on the opposite side has no support")
	}
}

func TestItemFrameOnNearbyBlockChangeBreaksWhenUnsupported(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	i := newTestItemFrame(w)
	i.Facing = math.Down
	w.blocks[[3]int{1, 2, 3}] = i

	// itemFrameCanBeSupportedAt checks the side opposite Facing (Up), so the block directly above
	// the frame is what needs to report no support.
	unsupported := newSupportlessBlock()
	unsupported.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = unsupported

	i.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("breakCalls = %v, want exactly one UseBreakOn call", w.breakCalls)
	}
}

func TestItemFrameReadStateFromWorldPullsFromTile(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	i := newTestItemFrame(w)

	tileFrame := tile.NewItemFrame(w, math.NewVector3(1, 2, 3))
	tileFrame.SetItem(dualItem{fakeItem{typeID: 42}})
	tileFrame.SetItemRotation(3)
	tileFrame.SetItemDropChance(0.5)
	w.tiles[[3]int{1, 2, 3}] = tileFrame

	i.ReadStateFromWorld()

	if i.FramedItem == nil {
		t.Fatal("expected FramedItem to be pulled from the tile")
	}
	if i.FramedItem.GetTypeId() != 42 {
		t.Errorf("FramedItem.GetTypeId() = %d, want 42", i.FramedItem.GetTypeId())
	}
	if i.ItemRotation != 3 {
		t.Errorf("ItemRotation = %d, want 3", i.ItemRotation)
	}
	if i.ItemDropChance != 0.5 {
		t.Errorf("ItemDropChance = %v, want 0.5", i.ItemDropChance)
	}
}

func TestItemFrameGetDropsForCompatibleToolIncludesFramedItemWhenChanceGuaranteed(t *testing.T) {
	w := &fakeWorld{}
	i := newTestItemFrame(w)
	i.FramedItem = fakeItem{typeID: 7}
	i.ItemDropChance = 1.0 // guarantee the random check passes

	drops := i.GetDropsForCompatibleTool(fakeItem{})
	found := false
	for _, d := range drops {
		if d.GetTypeId() == 7 {
			found = true
		}
	}
	if !found {
		t.Error("expected the framed item to be included in drops when drop chance is 1.0")
	}
}

func TestItemFrameGetPickedItemReturnsFramedItem(t *testing.T) {
	w := &fakeWorld{}
	i := newTestItemFrame(w)
	i.FramedItem = fakeItem{typeID: 9}

	if got := i.GetPickedItem(false); got == nil || got.GetTypeId() != 9 {
		t.Errorf("GetPickedItem() = %v, want the framed item (typeID 9)", got)
	}
}
