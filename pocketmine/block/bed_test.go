package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

type bedTileWorld struct {
	fakeWorld
	tile *tile.Bed
}

func (w *bedTileWorld) GetTile(pos Position) (Tile, bool) {
	if w.tile == nil {
		return nil, false
	}
	return w.tile, true
}

type recordingTransaction struct {
	blocks []Behavior
}

func (r *recordingTransaction) AddBlock(pos Position, blk Behavior) { r.blocks = append(r.blocks, blk) }

// bedPlaceWorld is solid everywhere (for support) except at one designated position, which is
// air (so the other bed half has room to be placed).
type bedPlaceWorld struct {
	fakeWorld
	airX, airY, airZ int
}

func (w *bedPlaceWorld) GetBlockAt(x, y, z int) Behavior {
	if x == w.airX && y == w.airY && z == w.airZ {
		air := NewAir(mustBlockIdentifier(0), "Air", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
		air.SetPosition(w, x, y, z)
		return air
	}
	filler := newTestBlock(false)
	filler.SetPosition(w, x, y, z)
	return filler
}

func newTestBed(w World, x, y, z int) *Bed {
	b := NewBed(mustBlockIdentifier(1055), "Test Bed", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.SetPosition(w, x, y, z)
	return b
}

func TestBedReadStateFromWorldPullsColorFromTile(t *testing.T) {
	w := &bedTileWorld{}
	bedTile := tile.NewBed(nil, math.Vector3{})
	bedTile.SetColor(blockutils.DyeColorBlue)
	w.tile = bedTile

	b := newTestBed(w, 1, 2, 3)
	b.ReadStateFromWorld()

	if b.GetColor() != blockutils.DyeColorBlue {
		t.Errorf("GetColor() = %v, want Blue", b.GetColor())
	}
}

func TestBedPlaceCreatesBothHalves(t *testing.T) {
	// No player -> Facing defaults to North, so the head half lands one block North (Z-1).
	w := &bedPlaceWorld{airX: 1, airY: 2, airZ: 2}
	foot := newTestBed(w, 1, 2, 3)
	tx := &recordingTransaction{}

	if !foot.Place(tx, fakeItem{}, foot, foot, math.Up, math.Vector3{}, nil) {
		t.Fatal("expected Place to succeed with solid support and a replaceable next block")
	}
	if len(tx.blocks) != 2 {
		t.Fatalf("expected 2 blocks added to the transaction, got %d", len(tx.blocks))
	}
	footState, ok := tx.blocks[0].(*Bed)
	if !ok || footState.Head {
		t.Errorf("expected the first block to be the non-head half")
	}
	headState, ok := tx.blocks[1].(*Bed)
	if !ok || !headState.Head {
		t.Errorf("expected the second block to be the head half")
	}
}

func TestBedGetOtherHalfRequiresMatchingFacingAndDifferentHead(t *testing.T) {
	w := &vineWorld{blocks: map[[3]int]Behavior{}}

	foot := newTestBed(w, 1, 2, 3)
	foot.Facing = math.North
	foot.Head = false

	head := newTestBed(w, 1, 2, 4) // North of (1,2,3) is (1,2,2); use South side instead
	head.Facing = math.North
	head.Head = true

	w.blocks[[3]int{1, 2, 2}] = head // North offset

	other, ok := foot.GetOtherHalf()
	if !ok || other != head {
		t.Errorf("GetOtherHalf() = (%v, %v), want (head, true)", other, ok)
	}
}

func TestBedOnEntityLandHalvesFallDistance(t *testing.T) {
	w := &fakeWorld{}
	b := newTestBed(w, 1, 2, 3)

	e := &onFireTrackingEntity{}
	e2 := &fallTrackingEntity{fakeItemLikeEntity: e.fakeItemLikeEntity, fallDistance: 10, motionY: -8}
	bounce, handled := b.OnEntityLand(e2)
	if !handled {
		t.Fatal("expected OnEntityLand to report handled")
	}
	if e2.fallDistance != 5 {
		t.Errorf("fallDistance = %v, want 5 (halved)", e2.fallDistance)
	}
	if want := -8 * -3.0 / 4; bounce != want {
		t.Errorf("bounce = %v, want %v", bounce, want)
	}
}

type fallTrackingEntity struct {
	fakeItemLikeEntity
	fallDistance float64
	motionY      float64
}

func (f *fallTrackingEntity) GetFallDistance() float64             { return f.fallDistance }
func (f *fallTrackingEntity) SetFallDistance(fallDistance float64) { f.fallDistance = fallDistance }
func (f *fallTrackingEntity) GetMotion() math.Vector3              { return math.Vector3{Y: f.motionY} }
