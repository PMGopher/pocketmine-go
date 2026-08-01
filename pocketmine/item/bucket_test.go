package item

import (
	"testing"

	"pocketmine-go/pocketmine/block"
)

var (
	_ Item = (*FishingRod)(nil)
	_ Item = (*Bucket)(nil)
	_ Item = (*MilkBucket)(nil)
	_ Item = (*HoneyBottle)(nil)
	_ Item = (*LiquidBucket)(nil)
)

func TestFishingRodDurability(t *testing.T) {
	f := NewFishingRod(NewItemIdentifier(FISHING_ROD), "Fishing Rod")
	if f.GetMaxDurability() != 384 {
		t.Errorf("GetMaxDurability() = %d, want 384", f.GetMaxDurability())
	}
	if f.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", f.GetMaxStackSize())
	}
}

func TestBucketMaxStackSize(t *testing.T) {
	b := NewBucket(NewItemIdentifier(BUCKET), "Bucket")
	if b.GetMaxStackSize() != 16 {
		t.Errorf("GetMaxStackSize() = %d, want 16", b.GetMaxStackSize())
	}
}

func TestMilkBucketMaxStackSize(t *testing.T) {
	m := NewMilkBucket(NewItemIdentifier(MILK_BUCKET), "Milk Bucket")
	if m.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", m.GetMaxStackSize())
	}
}

func TestHoneyBottleFoodValues(t *testing.T) {
	h := NewHoneyBottle(NewItemIdentifier(HONEY_BOTTLE), "Honey Bottle")
	if h.GetMaxStackSize() != 16 {
		t.Errorf("GetMaxStackSize() = %d, want 16", h.GetMaxStackSize())
	}
	if h.RequiresHunger() {
		t.Error("expected HoneyBottle.RequiresHunger() to be false")
	}
	if h.GetFoodRestore() != 6 {
		t.Errorf("GetFoodRestore() = %d, want 6", h.GetFoodRestore())
	}
	if h.GetSaturationRestore() != 1.2 {
		t.Errorf("GetSaturationRestore() = %v, want 1.2", h.GetSaturationRestore())
	}
}

func newTestLava() *block.Lava {
	idInfo, err := block.NewBlockIdentifier(10, nil)
	if err != nil {
		panic(err)
	}
	typeInfo := block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil)
	return block.NewLava(idInfo, "Lava", typeInfo)
}

func newTestWater() *block.Water {
	idInfo, err := block.NewBlockIdentifier(9, nil)
	if err != nil {
		panic(err)
	}
	typeInfo := block.NewBlockTypeInfo(block.BlockBreakInfoInstant(block.ToolTypeNone, 0), nil, nil)
	return block.NewWater(idInfo, "Water", typeInfo)
}

func TestLiquidBucketGetFuelTimeForLava(t *testing.T) {
	l := NewLiquidBucket(NewItemIdentifier(LAVA_BUCKET), "Lava Bucket", newTestLava())
	if l.GetFuelTime() != 20000 {
		t.Errorf("GetFuelTime() = %d, want 20000", l.GetFuelTime())
	}
}

func TestLiquidBucketGetFuelTimeForWater(t *testing.T) {
	l := NewLiquidBucket(NewItemIdentifier(WATER_BUCKET), "Water Bucket", newTestWater())
	if l.GetFuelTime() != 0 {
		t.Errorf("GetFuelTime() = %d, want 0", l.GetFuelTime())
	}
}

func TestLiquidBucketCloneDeepCopiesLiquid(t *testing.T) {
	lava := newTestLava()
	l := NewLiquidBucket(NewItemIdentifier(LAVA_BUCKET), "Lava Bucket", lava)

	clone := l.Clone().(*LiquidBucket)
	if clone.GetLiquid() == l.GetLiquid() {
		t.Error("expected Clone() to deep-copy the wrapped liquid block")
	}
}
