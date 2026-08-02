package block

import (
	"testing"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
)

func newTestFarmland(w World) *Farmland {
	f := NewFarmland(mustBlockIdentifier(1091), "Test Farmland", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	f.SetPosition(w, 1, 2, 3)
	return f
}

func TestFarmlandOnNearbyBlockChangeBecomesDirtWhenCoveredBySolid(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	f := newTestFarmland(w)

	solid := newTestBlock(false) // testBlock's IsSolid defaults to true
	solid.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = solid

	f.OnNearbyBlockChange()

	dirt, ok := w.lastSetBlock.(*Dirt)
	if !ok {
		t.Fatalf("expected SetBlock to be called with a *Dirt, got %T", w.lastSetBlock)
	}
	_ = dirt
}

func TestFarmlandOnNearbyBlockChangeStaysFarmlandWhenUncovered(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	f := newTestFarmland(w)

	// testBlock's default IsSolid() is true (matching Block's own default), so the filler needs
	// to be explicitly non-solid to exercise the "stays farmland" branch.
	air := VanillaAir().(*Air)
	air.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = air

	f.OnNearbyBlockChange()

	if w.lastSetBlock != nil {
		t.Errorf("expected no block change, got SetBlock(%T)", w.lastSetBlock)
	}
}

func TestFarmlandOnRandomTickBecomesDirtWhenFullyDried(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	f := newTestFarmland(w)
	f.Wetness = 0 // already dry, and canHydrate() finds no water in the default filler world

	f.OnRandomTick()

	if _, ok := w.lastSetBlock.(*Dirt); !ok {
		t.Fatalf("expected SetBlock to be called with a *Dirt, got %T", w.lastSetBlock)
	}
}

func TestFarmlandOnEntityLandTramplesIntoDirtWithGuaranteedRoll(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	f := newTestFarmland(w)

	living := entity.NewLiving(math.NewVector3(0, 0, 0), math.OneAABB())
	living.SetFallDistance(100) // guarantees GetRandomFloat() < fallDistance-0.5

	damage, handled := f.OnEntityLand(living)
	if handled {
		t.Error("expected OnEntityLand to defer to default fall damage (handled=false)")
	}
	if damage != 0 {
		t.Errorf("damage = %v, want 0", damage)
	}
	if _, ok := w.lastSetBlock.(*Dirt); !ok {
		t.Fatalf("expected SetBlock to be called with a *Dirt, got %T", w.lastSetBlock)
	}
}

func TestFarmlandOnEntityLandIgnoresNonLivingEntities(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	f := newTestFarmland(w)

	e := entity.NewEntity(math.NewVector3(0, 0, 0), math.OneAABB())
	e.SetFallDistance(100)

	f.OnEntityLand(e)

	if w.lastSetBlock != nil {
		t.Errorf("expected no trampling for a non-Living entity, got SetBlock(%T)", w.lastSetBlock)
	}
}
