package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestBuddingAmethyst(w World) *BuddingAmethyst {
	b := NewBuddingAmethyst(mustBlockIdentifier(1099), "Test Budding Amethyst", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.SetPosition(w, 1, 2, 3)
	return b
}

func TestBuddingAmethystTryGrowBudSpawnsSmallBudInAir(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	b := newTestBuddingAmethyst(w)
	air := VanillaAir().(*Air)
	air.SetPosition(w, 2, 2, 3)
	w.blocks[[3]int{2, 2, 3}] = air

	b.tryGrowBud(math.East)

	grown, ok := w.lastSetBlock.(*AmethystCluster)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *AmethystCluster, got %T", w.lastSetBlock)
	}
	if grown.Stage != AmethystClusterStageSmallBud {
		t.Errorf("Stage = %d, want AmethystClusterStageSmallBud", grown.Stage)
	}
	if grown.Facing != math.East {
		t.Errorf("Facing = %v, want East", grown.Facing)
	}
}

func TestBuddingAmethystTryGrowBudAdvancesExistingBudStage(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	b := newTestBuddingAmethyst(w)
	existing := NewAmethystCluster(mustBlockIdentifier(AMETHYST_CLUSTER), "Amethyst Cluster", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	existing.Stage = AmethystClusterStageSmallBud
	existing.SetFacing(math.East)
	existing.SetPosition(w, 2, 2, 3)
	w.blocks[[3]int{2, 2, 3}] = existing

	b.tryGrowBud(math.East)

	grown, ok := w.lastSetBlock.(*AmethystCluster)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *AmethystCluster, got %T", w.lastSetBlock)
	}
	if grown.Stage != AmethystClusterStageMediumBud {
		t.Errorf("Stage = %d, want AmethystClusterStageMediumBud", grown.Stage)
	}
}

func TestBuddingAmethystTryGrowBudDoesNothingWhenBudFacingDiffers(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	b := newTestBuddingAmethyst(w)
	existing := NewAmethystCluster(mustBlockIdentifier(AMETHYST_CLUSTER), "Amethyst Cluster", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	existing.Stage = AmethystClusterStageSmallBud
	existing.SetFacing(math.North) // doesn't match the face we grow towards below
	existing.SetPosition(w, 2, 2, 3)
	w.blocks[[3]int{2, 2, 3}] = existing

	b.tryGrowBud(math.East)

	if w.lastSetBlock != nil {
		t.Errorf("expected no growth, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestBuddingAmethystTryGrowBudDoesNothingAtMaxStage(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	b := newTestBuddingAmethyst(w)
	existing := NewAmethystCluster(mustBlockIdentifier(AMETHYST_CLUSTER), "Amethyst Cluster", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	existing.Stage = AmethystClusterStageCluster
	existing.SetFacing(math.East)
	existing.SetPosition(w, 2, 2, 3)
	w.blocks[[3]int{2, 2, 3}] = existing

	b.tryGrowBud(math.East)

	if w.lastSetBlock != nil {
		t.Errorf("expected no growth at max stage, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestBuddingAmethystTryGrowBudDoesNothingForUnrelatedBlock(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	b := newTestBuddingAmethyst(w)
	// Default filler at (2,2,3) is a solid, non-air, non-cluster testBlock.

	b.tryGrowBud(math.East)

	if w.lastSetBlock != nil {
		t.Errorf("expected no growth against an unrelated block, got SetBlock(%T)", w.lastSetBlock)
	}
}
