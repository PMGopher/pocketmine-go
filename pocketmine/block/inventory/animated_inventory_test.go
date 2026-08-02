package blockinventory

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

var _ inventory.Inventory = (*AnimatedInventory)(nil)

// fakeWorld implements block.World with no-ops except AddSound, which records calls for
// assertions.
type fakeWorld struct {
	sounds []sound.Sound
}

func (w *fakeWorld) GetBlockAt(x, y, z int) block.Behavior                 { return nil }
func (w *fakeWorld) SetBlock(pos block.Position, blk block.Behavior) error { return nil }
func (w *fakeWorld) GetTile(pos block.Position) (block.Tile, bool)         { return nil, false }
func (w *fakeWorld) AddTile(tile block.Tile)                               {}
func (w *fakeWorld) GetOrLoadChunkAtPosition(pos block.Position) (block.Chunk, bool) {
	return nil, false
}
func (w *fakeWorld) AddSound(pos math.Vector3, s sound.Sound)               { w.sounds = append(w.sounds, s) }
func (w *fakeWorld) ScheduleDelayedBlockUpdate(pos math.Vector3, delay int) {}
func (w *fakeWorld) GetFullLightAt(x, y, z int) int                         { return 15 }
func (w *fakeWorld) GetBlockLightAt(x, y, z int) int                        { return 15 }
func (w *fakeWorld) GetRealBlockSkyLightAt(x, y, z int) int                 { return 15 }
func (w *fakeWorld) GetSunAnglePercentage() float64                         { return 1 }
func (w *fakeWorld) GetNearbyEntities(bb math.AxisAlignedBB) []block.Entity { return nil }
func (w *fakeWorld) GetHighestAdjacentFullLightAt(x, y, z int) int          { return 15 }
func (w *fakeWorld) GetHighestAdjacentBlockLightAt(x, y, z int) int         { return 15 }
func (w *fakeWorld) GetPotentialLightAt(x, y, z int) int                    { return 15 }
func (w *fakeWorld) UseBreakOn(pos math.Vector3) bool                       { return true }

type fakePlayer struct{ name string }

func TestBlockInventoryTraitGetHolder(t *testing.T) {
	w := &fakeWorld{}
	holder := block.NewPosition(1, 2, 3, w)

	trait := BlockInventoryTrait{Holder: holder}
	if trait.GetHolder() != holder {
		t.Errorf("GetHolder() = %v, want %v", trait.GetHolder(), holder)
	}
}

func TestAnimatedInventoryOnOpenAnimatesAndPlaysSoundOnFirstViewer(t *testing.T) {
	w := &fakeWorld{}
	holder := block.NewPosition(1, 2, 3, w)
	animateCalls := []bool{}

	inv := NewAnimatedInventory(27, holder, sound.ChestOpenSound{}, sound.ChestCloseSound{}, func(isOpen bool) {
		animateCalls = append(animateCalls, isOpen)
	})

	p1 := &fakePlayer{"a"}
	inv.OnOpen(p1)

	if len(animateCalls) != 1 || animateCalls[0] != true {
		t.Fatalf("animateCalls = %v, want [true]", animateCalls)
	}
	if len(w.sounds) != 1 {
		t.Fatalf("len(sounds) = %d, want 1", len(w.sounds))
	}
	if _, ok := w.sounds[0].(sound.ChestOpenSound); !ok {
		t.Errorf("sounds[0] = %T, want ChestOpenSound", w.sounds[0])
	}
}

func TestAnimatedInventoryOnOpenDoesNotReanimateForSecondViewer(t *testing.T) {
	w := &fakeWorld{}
	holder := block.NewPosition(1, 2, 3, w)
	animateCalls := 0

	inv := NewAnimatedInventory(27, holder, sound.ChestOpenSound{}, sound.ChestCloseSound{}, func(isOpen bool) {
		animateCalls++
	})

	inv.OnOpen(&fakePlayer{"a"})
	inv.OnOpen(&fakePlayer{"b"})

	if animateCalls != 1 {
		t.Errorf("animateCalls = %d, want 1 (only the first viewer triggers animation)", animateCalls)
	}
}

func TestAnimatedInventoryOnCloseAnimatesOnlyWhenLastViewerLeaves(t *testing.T) {
	w := &fakeWorld{}
	holder := block.NewPosition(1, 2, 3, w)
	var closeAnimated bool

	inv := NewAnimatedInventory(27, holder, sound.ChestOpenSound{}, sound.ChestCloseSound{}, func(isOpen bool) {
		if !isOpen {
			closeAnimated = true
		}
	})

	p1, p2 := &fakePlayer{"a"}, &fakePlayer{"b"}
	inv.OnOpen(p1)
	inv.OnOpen(p2)

	inv.OnClose(p1)
	if closeAnimated {
		t.Fatal("expected no close animation while a second viewer remains")
	}

	inv.OnClose(p2)
	if !closeAnimated {
		t.Error("expected the close animation to fire once the last viewer closes")
	}
}

func TestAnimatedInventoryGetViewerCount(t *testing.T) {
	w := &fakeWorld{}
	holder := block.NewPosition(1, 2, 3, w)
	inv := NewAnimatedInventory(27, holder, sound.ChestOpenSound{}, sound.ChestCloseSound{}, nil)

	inv.OnOpen(&fakePlayer{"a"})
	inv.OnOpen(&fakePlayer{"b"})
	if inv.GetViewerCount() != 2 {
		t.Errorf("GetViewerCount() = %d, want 2", inv.GetViewerCount())
	}
}
