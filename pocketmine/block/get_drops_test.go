package block

import "testing"

// fakeItem is a minimal Item for exercising BlockBreakInfo.IsToolCompatible / Block.GetDrops.
type fakeItem struct {
	typeID       int
	toolType     ToolType
	harvestLevel int
	null         bool
	customName   string
}

func (f fakeItem) GetTypeId() int                   { return f.typeID }
func (f fakeItem) GetBlockToolType() ToolType       { return f.toolType }
func (f fakeItem) GetBlockToolHarvestLevel() int    { return f.harvestLevel }
func (f fakeItem) GetMiningEfficiency(bool) float64 { return 1 }
func (f fakeItem) Pop()                             {}
func (f fakeItem) IsNull() bool                     { return f.null }
func (f fakeItem) GetCustomName() string            { return f.customName }

type fakeToolTier struct{ level int }

func (f fakeToolTier) GetHarvestLevel() int { return f.level }

// dropsTestBlock distinguishes compatible/incompatible-tool drops so GetDrops' dispatch through
// b.self (rather than Block's own methods) can be verified end-to-end.
type dropsTestBlock struct {
	Block
}

func newDropsTestBlock() *dropsTestBlock {
	idInfo, err := NewBlockIdentifier(1001, nil)
	if err != nil {
		panic(err)
	}
	breakInfo := BlockBreakInfoPickaxe(1.5, fakeToolTier{level: 1}, nil)
	d := &dropsTestBlock{Block: NewBlock(idInfo, "Drops Test Block", NewBlockTypeInfo(breakInfo, nil, nil))}
	d.Init(d)
	return d
}

func (d *dropsTestBlock) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *dropsTestBlock) GetDropsForCompatibleTool(item Item) []Item {
	return []Item{fakeItem{toolType: ToolTypePickaxe}}
}

func (d *dropsTestBlock) GetDropsForIncompatibleTool(item Item) []Item {
	return nil
}

func TestGetDropsDispatchesOnToolCompatibility(t *testing.T) {
	blk := newDropsTestBlock()

	pickaxe := fakeItem{toolType: ToolTypePickaxe, harvestLevel: 1}
	drops := blk.GetDrops(pickaxe)
	if len(drops) != 1 {
		t.Fatalf("compatible tool: got %d drops, want 1 (GetDrops did not reach the override via self-dispatch)", len(drops))
	}

	nothing := fakeItem{toolType: ToolTypeShovel, harvestLevel: 1}
	drops = blk.GetDrops(nothing)
	if drops != nil {
		t.Fatalf("incompatible tool: got %d drops, want 0", len(drops))
	}
}
