package world

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world/generator"
)

func newTestWorldManager(t *testing.T) *WorldManager {
	t.Helper()
	return NewWorldManager(t.TempDir(), convert.NewBlockTranslator(), []block.Behavior{
		block.VanillaAir(),
		block.VanillaBedrock(),
		block.VanillaStone(),
		block.VanillaDirt(),
		block.VanillaGrass(),
		block.VanillaWater(),
	})
}

func TestGenerateWorldCreatesALoadedWorldWithAnAssignedIDAndFolderName(t *testing.T) {
	m := newTestWorldManager(t)

	w, err := m.GenerateWorld("myworld", generator.NewNormal(1), 1, "normal", "")
	if err != nil {
		t.Fatalf("GenerateWorld: %v", err)
	}

	if w.GetID() == 0 {
		t.Error("GetID() = 0, want a nonzero assigned ID")
	}
	if got := w.GetFolderName(); got != "myworld" {
		t.Errorf("GetFolderName() = %q, want %q", got, "myworld")
	}
	if !m.IsWorldLoaded("myworld") {
		t.Error("IsWorldLoaded(\"myworld\") = false after GenerateWorld")
	}
	if !m.IsWorldGenerated("myworld") {
		t.Error("IsWorldGenerated(\"myworld\") = false after GenerateWorld")
	}
	if got, ok := m.GetWorldByName("myworld"); !ok || got != w {
		t.Error("GetWorldByName(\"myworld\") did not return the generated world")
	}
	if got, ok := m.GetWorld(w.GetID()); !ok || got != w {
		t.Error("GetWorld(w.GetID()) did not return the generated world")
	}
}

func TestGenerateWorldRejectsEmptyOrDuplicateNames(t *testing.T) {
	m := newTestWorldManager(t)

	if _, err := m.GenerateWorld("", generator.NewNormal(1), 1, "normal", ""); err == nil {
		t.Error("GenerateWorld(\"\") = nil error, want an error")
	}
	if _, err := m.GenerateWorld("  ", generator.NewNormal(1), 1, "normal", ""); err == nil {
		t.Error("GenerateWorld(\"  \") = nil error, want an error")
	}

	if _, err := m.GenerateWorld("dup", generator.NewNormal(1), 1, "normal", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := m.GenerateWorld("dup", generator.NewNormal(1), 1, "normal", ""); err == nil {
		t.Error("GenerateWorld with an already-generated name = nil error, want an error")
	}
}

func TestLoadWorldReturnsTheSameInstanceIfAlreadyLoaded(t *testing.T) {
	m := newTestWorldManager(t)
	generated, err := m.GenerateWorld("myworld", generator.NewNormal(1), 1, "normal", "")
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := m.LoadWorld("myworld")
	if err != nil {
		t.Fatal(err)
	}
	if loaded != generated {
		t.Error("LoadWorld on an already-loaded world returned a different instance")
	}
}

func TestLoadWorldRejectsAnUngeneratedName(t *testing.T) {
	m := newTestWorldManager(t)
	if _, err := m.LoadWorld("never-generated"); err == nil {
		t.Error("LoadWorld on a never-generated name = nil error, want an error")
	}
}

func TestUnloadWorldThenLoadWorldReconstructsAnEquivalentWorldFromDisk(t *testing.T) {
	m := newTestWorldManager(t)
	original, err := m.GenerateWorld("myworld", generator.NewNormal(7), 7, "normal", "")
	if err != nil {
		t.Fatal(err)
	}

	// Force some real terrain to be generated and saved, so reloading has something concrete to
	// verify against.
	before := original.GetBlockAt(5, 0, 5).GetTypeId()
	originalID := original.GetID()

	ok, err := m.UnloadWorld(original, false)
	if err != nil || !ok {
		t.Fatalf("UnloadWorld: ok=%v err=%v", ok, err)
	}
	if m.IsWorldLoaded("myworld") {
		t.Error("IsWorldLoaded(\"myworld\") = true after UnloadWorld")
	}
	if _, ok := m.GetWorld(originalID); ok {
		t.Error("GetWorld(originalID) still found the unloaded world")
	}

	reloaded, err := m.LoadWorld("myworld")
	if err != nil {
		t.Fatalf("LoadWorld after unload: %v", err)
	}
	if reloaded == original {
		t.Error("LoadWorld after unload returned the same (closed) instance instead of a fresh one")
	}
	if got := reloaded.GetFolderName(); got != "myworld" {
		t.Errorf("reloaded GetFolderName() = %q, want %q", got, "myworld")
	}
	if got := reloaded.GetBlockAt(5, 0, 5).GetTypeId(); got != before {
		t.Errorf("reloaded terrain at (5,0,5) = %d, want %d (same seed/generator round-tripped through level.dat+LevelDB)", got, before)
	}
}

func TestSetDefaultWorldOnlyAcceptsALoadedWorldOrNil(t *testing.T) {
	m := newTestWorldManager(t)
	w, err := m.GenerateWorld("myworld", generator.NewNormal(1), 1, "normal", "")
	if err != nil {
		t.Fatal(err)
	}

	m.SetDefaultWorld(w)
	if m.GetDefaultWorld() != w {
		t.Fatal("SetDefaultWorld did not accept a loaded world")
	}

	// A world instance that was never registered with this manager (e.g. one built via New
	// directly, bypassing GenerateWorld/LoadWorld) must not become the default.
	unmanaged := New(generator.NewNormal(1), m.translator, m.knownBlocks)
	m.SetDefaultWorld(unmanaged)
	if m.GetDefaultWorld() != w {
		t.Error("SetDefaultWorld accepted an unmanaged/unloaded world")
	}

	m.SetDefaultWorld(nil)
	if m.GetDefaultWorld() != nil {
		t.Error("SetDefaultWorld(nil) did not clear the default world")
	}
}

func TestUnloadWorldRefusesToUnloadTheDefaultWorldWithoutForce(t *testing.T) {
	m := newTestWorldManager(t)
	w, err := m.GenerateWorld("myworld", generator.NewNormal(1), 1, "normal", "")
	if err != nil {
		t.Fatal(err)
	}
	m.SetDefaultWorld(w)

	if _, err := m.UnloadWorld(w, false); err == nil {
		t.Error("UnloadWorld(defaultWorld, false) = nil error, want an error")
	}
	if !m.IsWorldLoaded("myworld") {
		t.Error("the default world was unloaded despite forceUnload=false")
	}

	ok, err := m.UnloadWorld(w, true)
	if err != nil || !ok {
		t.Fatalf("UnloadWorld(defaultWorld, true): ok=%v err=%v", ok, err)
	}
	if m.GetDefaultWorld() != nil {
		t.Error("GetDefaultWorld() still set after force-unloading it")
	}
}

func TestUnloadWorldRefusesDuringItsOwnTick(t *testing.T) {
	m := newTestWorldManager(t)
	w, err := m.GenerateWorld("myworld", generator.NewNormal(1), 1, "normal", "")
	if err != nil {
		t.Fatal(err)
	}

	w.doingTick = true // simulates being called re-entrantly from inside DoTick
	if _, err := m.UnloadWorld(w, true); err == nil {
		t.Error("UnloadWorld while IsDoingTick()=true = nil error, want an error")
	}
	w.doingTick = false

	if _, err := m.UnloadWorld(w, true); err != nil {
		t.Errorf("UnloadWorld after tick finished: %v", err)
	}
}

func TestTickAdvancesEveryManagedWorldAndAutosavesOnSchedule(t *testing.T) {
	m := newTestWorldManager(t)
	w, err := m.GenerateWorld("myworld", generator.NewNormal(1), 1, "normal", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := m.SetAutoSaveInterval(2); err != nil {
		t.Fatal(err)
	}

	m.Tick(1)
	if got := w.GetTime(); got != 1 {
		t.Errorf("GetTime() after one Tick() = %d, want 1", got)
	}

	wd := m.worldData[w.GetID()]
	if got := wd.GetTime(); got != 0 {
		t.Errorf("level.dat Time before autosave interval elapsed = %d, want 0 (unsaved)", got)
	}

	m.Tick(2)
	if got := w.GetTime(); got != 2 {
		t.Errorf("GetTime() after two Tick() calls = %d, want 2", got)
	}
	if got := wd.GetTime(); got != 2 {
		t.Errorf("level.dat Time after the autosave interval elapsed = %d, want 2", got)
	}
}

func TestFindEntityAcrossWorlds(t *testing.T) {
	m := newTestWorldManager(t)
	a, err := m.GenerateWorld("a", generator.NewNormal(1), 1, "normal", "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := m.GenerateWorld("b", generator.NewNormal(1), 1, "normal", "")
	if err != nil {
		t.Fatal(err)
	}

	bb, err := math.NewAxisAlignedBB(0, 0, 0, 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	ent := newFakeEntity(42, bb)
	b.AddEntity(ent)

	if _, ok := m.FindEntity(1); ok {
		t.Error("FindEntity found an entity that was never added")
	}
	found, ok := m.FindEntity(42)
	if !ok {
		t.Fatal("FindEntity(42) = not found, want the entity added to world b")
	}
	if found != block.Entity(ent) {
		t.Error("FindEntity(42) returned a different entity than the one added")
	}
	_ = a
}
