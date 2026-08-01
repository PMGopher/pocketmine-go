package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

type noteTileWorld struct {
	fakeWorld
	tile *tile.Note
}

func (w *noteTileWorld) GetTile(pos Position) (Tile, bool) {
	if w.tile == nil {
		return nil, false
	}
	return w.tile, true
}

func TestNoteReadStateFromWorldPullsPitchFromTile(t *testing.T) {
	w := &noteTileWorld{}
	noteTile := tile.NewNote(nil, math.Vector3{})
	noteTile.SetPitch(17)
	w.tile = noteTile

	n := NewNote(mustBlockIdentifier(1043), "Test Note", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	n.SetPosition(w, 1, 2, 3)

	n.ReadStateFromWorld()

	if n.GetPitch() != 17 {
		t.Errorf("GetPitch() = %d, want 17", n.GetPitch())
	}
}

func TestNoteReadStateFromWorldDefaultsWithoutTile(t *testing.T) {
	w := &noteTileWorld{}
	n := NewNote(mustBlockIdentifier(1044), "Test Note", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	n.Pitch = 5
	n.SetPosition(w, 1, 2, 3)

	n.ReadStateFromWorld()

	if n.GetPitch() != NoteMinPitch {
		t.Errorf("GetPitch() = %d, want %d (no tile present)", n.GetPitch(), NoteMinPitch)
	}
}
