package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

func newTestLectern(w World) *Lectern {
	l := NewLectern(mustBlockIdentifier(1084), "Test Lectern", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	l.SetPosition(w, 1, 2, 3)
	return l
}

// fakeWritableBook satisfies block.WritableBookBase (Item + PageExists).
type fakeWritableBook struct {
	fakeItem
	pages int
}

func (f fakeWritableBook) PageExists(pageID int) bool { return pageID >= 0 && pageID < f.pages }

// dualWritableBook additionally satisfies tile.Item, for exercising the block<->tile bridging in
// Lectern.ReadStateFromWorld.
type dualWritableBook struct {
	fakeWritableBook
}

func (dualWritableBook) GetCustomBlockData() (*nbt.CompoundTag, bool) { return nil, false }
func (dualWritableBook) GetNamedTag() *nbt.CompoundTag                { return nbt.NewCompoundTag() }
func (dualWritableBook) HasCustomName() bool                          { return false }

func TestLecternPlaceFacesOppositePlayer(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)
	tx := &fakeBlockTransaction{}
	player := &fakeSignPlayer{}

	l.Place(tx, fakeItem{}, l, l, math.Up, math.Vector3{}, player)

	if l.Facing != math.Opposite(player.GetHorizontalFacing()) {
		t.Errorf("Facing = %v, want opposite of player facing", l.Facing)
	}
}

func TestLecternOnInteractPlacesBookWhenEmpty(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)
	book := fakeWritableBook{pages: 3}

	if !l.OnInteract(book, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to return true")
	}
	if l.Book == nil {
		t.Fatal("expected a book to be placed")
	}
	if w.lastSetBlock == nil {
		t.Error("expected the lectern to be written back to the world")
	}
}

func TestLecternOnInteractDoesNothingWhenBookAlreadyPresent(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)
	existing := fakeWritableBook{pages: 1}
	l.Book = existing

	newBook := fakeWritableBook{pages: 5}
	if !l.OnInteract(newBook, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to return true")
	}
	if l.Book != WritableBookBase(existing) {
		t.Error("expected the existing book not to be replaced")
	}
	if w.lastSetBlock != nil {
		t.Error("expected no state write when a book is already present")
	}
}

func TestLecternOnInteractIgnoresNonBookItem(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)

	if !l.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to return true even for a non-book item")
	}
	if l.Book != nil {
		t.Error("expected no book to be placed for a non-WritableBookBase item")
	}
}

func TestLecternOnAttackEjectsBook(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)
	l.Book = fakeWritableBook{pages: 2}

	if l.OnAttack(fakeItem{}, math.Up, nil) {
		t.Error("expected OnAttack to always return false")
	}
	if l.Book != nil {
		t.Error("expected the book to be cleared")
	}
}

func TestLecternOnAttackReturnsFalseWhenEmpty(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)

	if l.OnAttack(fakeItem{}, math.Up, nil) {
		t.Error("expected OnAttack to return false when there's no book either")
	}
}

func TestLecternOnPageTurnSamePageReturnsTrueRegardlessOfBook(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)

	if !l.OnPageTurn(0) {
		t.Error("expected turning to the already-viewed page to return true")
	}
}

func TestLecternOnPageTurnFailsWhenNoBook(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)

	if l.OnPageTurn(1) {
		t.Error("expected OnPageTurn to fail with no book present")
	}
}

func TestLecternOnPageTurnFailsWhenPageOutOfRange(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)
	l.Book = fakeWritableBook{pages: 2}

	if l.OnPageTurn(5) {
		t.Error("expected OnPageTurn to fail for an out-of-range page")
	}
}

func TestLecternOnPageTurnSucceedsAndSchedulesUpdateOnlyOnce(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)
	l.Book = fakeWritableBook{pages: 3}

	if !l.OnPageTurn(1) {
		t.Fatal("expected the first page turn to succeed")
	}
	if l.ViewedPage != 1 {
		t.Errorf("ViewedPage = %d, want 1", l.ViewedPage)
	}
	if !l.ProducingSignal {
		t.Error("expected ProducingSignal to become true")
	}
	if w.scheduleDelay != 1 {
		t.Errorf("scheduleDelay = %d, want 1", w.scheduleDelay)
	}

	w.scheduleDelay = 0 // reset to prove the second turn doesn't reschedule
	if !l.OnPageTurn(2) {
		t.Fatal("expected the second page turn to succeed")
	}
	if w.scheduleDelay != 0 {
		t.Error("expected no second delayed-update schedule while already producing a signal")
	}
}

func TestLecternOnScheduledUpdateClearsProducingSignal(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)
	l.ProducingSignal = true

	l.OnScheduledUpdate()

	if l.ProducingSignal {
		t.Error("expected ProducingSignal to be cleared")
	}
	if w.lastSetBlock == nil {
		t.Error("expected the lectern to be written back to the world")
	}
}

func TestLecternReadStateFromWorldPullsFromTile(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	l := newTestLectern(w)

	tileLectern := tile.NewLectern(w, math.NewVector3(1, 2, 3))
	tileLectern.SetBook(dualWritableBook{fakeWritableBook{fakeItem: fakeItem{typeID: 55}, pages: 4}})
	tileLectern.SetViewedPage(2)
	w.tiles[[3]int{1, 2, 3}] = tileLectern

	l.ReadStateFromWorld()

	if l.Book == nil {
		t.Fatal("expected a book to be pulled from the tile")
	}
	if l.Book.GetTypeId() != 55 {
		t.Errorf("Book.GetTypeId() = %d, want 55", l.Book.GetTypeId())
	}
	if l.ViewedPage != 2 {
		t.Errorf("ViewedPage = %d, want 2", l.ViewedPage)
	}
}

func TestLecternGetDropsIncludesBook(t *testing.T) {
	w := &fakeWorld{}
	l := newTestLectern(w)
	l.Book = fakeWritableBook{fakeItem: fakeItem{typeID: 8}, pages: 1}

	drops := l.GetDrops(fakeItem{})
	found := false
	for _, d := range drops {
		if d.GetTypeId() == 8 {
			found = true
		}
	}
	if !found {
		t.Error("expected the book to be included in drops")
	}
}
