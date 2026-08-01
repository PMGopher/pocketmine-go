package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world/sound"
)

func newTestJukebox(w World) *Jukebox {
	j := NewJukebox(mustBlockIdentifier(1090), "Test Jukebox", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	j.SetPosition(w, 1, 2, 3)
	return j
}

// fakeRecord satisfies block.Record (Item + GetRecordType).
type fakeRecord struct {
	fakeItem
	recordType blockutils.RecordType
}

func (f fakeRecord) GetRecordType() blockutils.RecordType { return f.recordType }

// dualRecord additionally satisfies tile.Item, for exercising the block<->tile bridging in
// Jukebox.ReadStateFromWorld.
type dualRecord struct {
	fakeRecord
}

func (dualRecord) GetCustomBlockData() (*nbt.CompoundTag, bool) { return nil, false }
func (dualRecord) GetNamedTag() *nbt.CompoundTag                { return nbt.NewCompoundTag() }
func (dualRecord) HasCustomName() bool                          { return false }

func TestJukeboxGetFuelTime(t *testing.T) {
	w := &fakeWorld{}
	j := newTestJukebox(w)
	if j.GetFuelTime() != 300 {
		t.Errorf("GetFuelTime() = %d, want 300", j.GetFuelTime())
	}
}

func TestJukeboxInsertRecordStartsSound(t *testing.T) {
	w := &fakeWorld{}
	j := newTestJukebox(w)

	j.InsertRecord(fakeRecord{recordType: blockutils.RecordTypeDiskCat})

	if j.RecordItem == nil {
		t.Fatal("expected a record to be inserted")
	}
	if len(w.sounds) != 1 {
		t.Fatalf("expected 1 sound, got %d", len(w.sounds))
	}
	if s, ok := w.sounds[0].(sound.RecordSound); !ok || s.RecordType != blockutils.RecordTypeDiskCat {
		t.Errorf("sound = %#v, want a RecordSound with RecordTypeDiskCat", w.sounds[0])
	}
}

func TestJukeboxInsertRecordDoesNothingWhenAlreadyOccupied(t *testing.T) {
	w := &fakeWorld{}
	j := newTestJukebox(w)
	j.RecordItem = fakeRecord{recordType: blockutils.RecordTypeDiskCat}

	j.InsertRecord(fakeRecord{recordType: blockutils.RecordTypeDiskBlocks})

	if j.RecordItem.GetRecordType() != blockutils.RecordTypeDiskCat {
		t.Error("expected the existing record not to be replaced")
	}
}

func TestJukeboxEjectRecordClearsAndStopsSound(t *testing.T) {
	w := &fakeWorld{}
	j := newTestJukebox(w)
	j.RecordItem = fakeRecord{recordType: blockutils.RecordTypeDiskCat}

	j.EjectRecord()

	if j.RecordItem != nil {
		t.Error("expected RecordItem to be cleared")
	}
	if len(w.sounds) != 1 {
		t.Fatalf("expected 1 sound, got %d", len(w.sounds))
	}
}

func TestJukeboxOnInteractEjectsWhenOccupied(t *testing.T) {
	w := &fakeWorld{}
	j := newTestJukebox(w)
	j.RecordItem = fakeRecord{recordType: blockutils.RecordTypeDiskCat}

	if !j.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Fatal("expected OnInteract to return true")
	}
	if j.RecordItem != nil {
		t.Error("expected the record to be ejected")
	}
	if w.lastSetBlock == nil {
		t.Error("expected the jukebox to be written back to the world")
	}
}

func TestJukeboxOnBreakStopsSound(t *testing.T) {
	w := &fakeWorld{}
	j := newTestJukebox(w)

	if !j.OnBreak(fakeItem{}, nil, nil) {
		t.Fatal("expected OnBreak to fall through to the default (true)")
	}
	if len(w.sounds) != 1 {
		t.Fatalf("expected 1 stop sound, got %d", len(w.sounds))
	}
}

func TestJukeboxGetDropsForCompatibleToolIncludesRecord(t *testing.T) {
	w := &fakeWorld{}
	j := newTestJukebox(w)
	j.RecordItem = fakeRecord{fakeItem: fakeItem{typeID: 5}, recordType: blockutils.RecordTypeDiskCat}

	drops := j.GetDropsForCompatibleTool(fakeItem{})
	found := false
	for _, d := range drops {
		if d.GetTypeId() == 5 {
			found = true
		}
	}
	if !found {
		t.Error("expected the record to be included in drops")
	}
}

func TestJukeboxReadStateFromWorldPullsRecordFromTile(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	j := newTestJukebox(w)

	tileJukebox := tile.NewJukebox(w, math.NewVector3(1, 2, 3))
	tileJukebox.SetRecord(dualRecord{fakeRecord{fakeItem: fakeItem{typeID: 42}, recordType: blockutils.RecordTypeDiskChirp}})
	w.tiles[[3]int{1, 2, 3}] = tileJukebox

	j.ReadStateFromWorld()

	if j.RecordItem == nil {
		t.Fatal("expected a record to be pulled from the tile")
	}
	if j.RecordItem.GetTypeId() != 42 {
		t.Errorf("RecordItem.GetTypeId() = %d, want 42", j.RecordItem.GetTypeId())
	}
	if j.RecordItem.GetRecordType() != blockutils.RecordTypeDiskChirp {
		t.Errorf("RecordItem.GetRecordType() = %v, want DiskChirp", j.RecordItem.GetRecordType())
	}
}
