package object_test

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/object"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/generator"
)

func newTestWorld(t *testing.T) *world.World {
	t.Helper()
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), nil)
	return world.New(gen, convert.NewBlockTranslator(), []block.Behavior{
		block.VanillaAir(), block.VanillaBedrock(), block.VanillaStone(), block.VanillaDirt(), block.VanillaGrass(),
		block.VanillaSand(), block.VanillaTNT(),
	})
}

// groundY is the Y of the first air block above the terrain at x=0,z=0.
func groundY(t *testing.T, w *world.World) int {
	t.Helper()
	h, ok := w.GetOrLoadChunk(0, 0).GetHighestBlockAt(0, 0)
	if !ok {
		t.Fatal("test world has no terrain at 0,0")
	}
	return h + 1
}

// tickUntil ticks e (like World::tickEntities does) until done returns true or maxTicks pass.
func tickUntil(e world.Entity, maxTicks int, done func() bool) int {
	for tick := 1; tick <= maxTicks; tick++ {
		e.OnUpdate(int64(tick))
		if done() {
			return tick
		}
	}
	return -1
}

func TestPrimedTNTExplodesWhenFuseRunsOut(t *testing.T) {
	w := newTestWorld(t)
	y := groundY(t, w)
	tnt := object.NewPrimedTNT(entity.NewLocation(0.5, float64(y), 0.5, w, 0, 0), nil)
	if tnt.GetFuse() != 80 {
		t.Fatalf("default fuse = %d, want 80", tnt.GetFuse())
	}
	tnt.SetFuse(5)

	if got := tickUntil(tnt, 20, tnt.IsFlaggedForDespawn); got != 5 {
		t.Fatalf("TNT despawned after %d ticks, want 5", got)
	}
	// The explosion (radius 4) centred on the ground must have destroyed the grass block below.
	if b := w.GetBlockAt(0, y-1, 0); b.GetTypeId() != block.VanillaAir().GetTypeId() {
		t.Errorf("block under the TNT is %v after the explosion, want air", b.GetName())
	}
}

func TestPrimedTNTSavesFuse(t *testing.T) {
	w := newTestWorld(t)
	tnt := object.NewPrimedTNT(entity.NewLocation(0.5, 50, 0.5, w, 0, 0), nil)
	tnt.SetFuse(33)
	created, err := entity.GetEntityFactory().CreateFromData(w, tnt.SaveNBT())
	if err != nil {
		t.Fatal(err)
	}
	if loaded, ok := created.(*object.PrimedTNT); !ok || loaded.GetFuse() != 33 {
		t.Errorf("reloaded = %T (fuse %v), want *PrimedTNT with fuse 33", created, created)
	}
}

func TestFallingBlockLandsAndPlacesBlock(t *testing.T) {
	w := newTestWorld(t)
	y := groundY(t, w)
	fb := object.NewFallingBlock(entity.NewLocation(0.5, float64(y+5), 0.5, w, 0, 0), block.VanillaSand(), nil)

	if got := tickUntil(fb, 200, fb.IsFlaggedForDespawn); got < 0 {
		t.Fatal("the falling block never landed")
	}
	if b := w.GetBlockAt(0, y, 0); b.GetTypeId() != block.VanillaSand().GetTypeId() {
		t.Errorf("block at the landing spot is %v, want sand", b.GetName())
	}
}

func TestExperienceOrbSizes(t *testing.T) {
	// ExperienceOrb::splitIntoOrbSizes picks the largest orb sizes first.
	got := object.SplitIntoOrbSizes(40)
	want := []int{37, 3}
	if len(got) != len(want) {
		t.Fatalf("SplitIntoOrbSizes(40) = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SplitIntoOrbSizes(40) = %v, want %v", got, want)
		}
	}
	if object.GetMaxOrbSize(2500) != 2477 || object.GetMaxOrbSize(1) != 1 {
		t.Error("GetMaxOrbSize differs from ExperienceOrb::getMaxOrbSize")
	}
}

func TestItemEntityMerging(t *testing.T) {
	w := newTestWorld(t)
	apple := item.VanillaApple()
	apple.SetCount(10)
	a := object.NewItemEntity(entity.NewLocation(0.5, 50, 0.5, w, 0, 0), apple, nil)
	b := object.NewItemEntity(entity.NewLocation(0.6, 50, 0.5, w, 0, 0), apple.Clone(), nil)

	if !a.IsMergeable(b) {
		t.Fatal("two stacks of the same item are not mergeable")
	}
	if !a.TryMergeInto(b) {
		t.Fatal("TryMergeInto failed")
	}
	if got := b.GetItem().GetCount() + a.GetItem().GetCount()*boolToInt(!a.IsFlaggedForDespawn()); got != 20 {
		t.Errorf("total count after merging = %d, want 20", got)
	}

	stick := object.NewItemEntity(entity.NewLocation(0.5, 50, 0.5, w, 0, 0), item.VanillaStick(), nil)
	if a.IsMergeable(stick) {
		t.Error("different items are mergeable")
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
