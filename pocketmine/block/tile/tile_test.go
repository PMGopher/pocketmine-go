package tile

import (
	"testing"

	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

type fakeWorld struct {
	removed []Tile
}

func (w *fakeWorld) RemoveTile(t Tile) { w.removed = append(w.removed, t) }

type fakeItem struct {
	blockNbt   *nbt.CompoundTag
	hasBlkNbt  bool
	customName string
	hasName    bool
}

func (f fakeItem) GetCustomBlockData() (*nbt.CompoundTag, bool) { return f.blockNbt, f.hasBlkNbt }
func (f fakeItem) HasCustomName() bool                          { return f.hasName }
func (f fakeItem) GetCustomName() string                        { return f.customName }

func TestNoteSaveNBTRoundTrip(t *testing.T) {
	w := &fakeWorld{}
	n := NewNote(w, math.Vector3{X: 1, Y: 2, Z: 3})
	n.SetPitch(12)

	saved := n.SaveNBT()
	if got, err := saved.GetString(TagID); err != nil || string(got) != "Music" {
		t.Errorf("saved id = %q, err %v, want \"Music\"", got, err)
	}
	if got, err := saved.GetInt(TagX); err != nil || int(got) != 1 {
		t.Errorf("saved x = %d, err %v, want 1", got, err)
	}

	decoded := NewNote(w, math.Vector3{})
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if decoded.GetPitch() != 12 {
		t.Errorf("GetPitch() = %d, want 12", decoded.GetPitch())
	}
}

func TestNoteSetPitchPanicsOutOfRange(t *testing.T) {
	w := &fakeWorld{}
	n := NewNote(w, math.Vector3{})

	defer func() {
		if recover() == nil {
			t.Error("expected SetPitch to panic for an out-of-range pitch")
		}
	}()
	n.SetPitch(NoteMaxPitch + 1)
}

func TestTileCloseRemovesFromWorld(t *testing.T) {
	w := &fakeWorld{}
	n := NewNote(w, math.Vector3{X: 5, Y: 6, Z: 7})

	if n.IsClosed() {
		t.Fatal("expected a fresh tile not to be closed")
	}
	n.Close()
	if !n.IsClosed() {
		t.Error("expected Close to mark the tile closed")
	}
	if len(w.removed) != 1 || w.removed[0] != Tile(n) {
		t.Errorf("expected RemoveTile to be called once with the tile itself, got %v", w.removed)
	}

	// Closing again should be a no-op (not double-remove).
	n.Close()
	if len(w.removed) != 1 {
		t.Errorf("expected Close to be idempotent, got %d RemoveTile calls", len(w.removed))
	}
}

func TestTileCopyDataFromItemAppliesCustomBlockData(t *testing.T) {
	w := &fakeWorld{}
	n := NewNote(w, math.Vector3{})

	blockNbt := nbt.NewCompoundTag()
	blockNbt.SetByte("note", nbt.ByteTag(9))
	item := fakeItem{blockNbt: blockNbt, hasBlkNbt: true}

	n.CopyDataFromItem(item)
	if n.GetPitch() != 9 {
		t.Errorf("GetPitch() = %d, want 9 (from copied item block data)", n.GetPitch())
	}
}

type namedTile struct {
	TileBase
	NameableComponent
}

func (n *namedTile) SaveID() string                        { return "Named" }
func (n *namedTile) ReadSaveData(t *nbt.CompoundTag) error { n.LoadName(t); return nil }
func (n *namedTile) WriteSaveData(t *nbt.CompoundTag)      { n.SaveName(t) }
func (n *namedTile) GetDefaultName() string                { return "Default" }

func TestNameableComponentFallsBackToDefaultName(t *testing.T) {
	w := &fakeWorld{}
	nt := &namedTile{TileBase: NewTileBase(w, math.Vector3{})}
	nt.Init(nt)

	if got := nt.GetName(nt); got != "Default" {
		t.Errorf("GetName() = %q, want %q", got, "Default")
	}
	if nt.HasName() {
		t.Error("expected HasName() to be false before any custom name is set")
	}

	nt.SetName("Custom")
	if got := nt.GetName(nt); got != "Custom" {
		t.Errorf("GetName() = %q, want %q", got, "Custom")
	}
	if !nt.HasName() {
		t.Error("expected HasName() to be true after SetName")
	}

	nt.SetName("")
	if nt.HasName() {
		t.Error("expected SetName(\"\") to clear the custom name")
	}
}
