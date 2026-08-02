package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

type bannerTileWorld struct {
	fakeWorld
	tile *tile.Banner
}

func (w *bannerTileWorld) GetTile(pos Position) (Tile, bool) {
	if w.tile == nil {
		return nil, false
	}
	return w.tile, true
}

func newTestFloorBanner(w World) *FloorBanner {
	f := NewFloorBanner(mustBlockIdentifier(1045), "Test Floor Banner", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	f.SetPosition(w, 1, 2, 3)
	return f
}

func newTestWallBanner(w World) *WallBanner {
	wb := NewWallBanner(mustBlockIdentifier(1046), "Test Wall Banner", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	wb.SetPosition(w, 1, 2, 3)
	return wb
}

func TestBaseBannerReadStateFromWorldPullsColorAndPatternsFromTile(t *testing.T) {
	w := &bannerTileWorld{}
	bannerTile := tile.NewBanner(nil, math.Vector3{})
	bannerTile.SetBaseColor(blockutils.DyeColorBlue)
	bannerTile.SetPatterns([]blockutils.BannerPatternLayer{
		blockutils.NewBannerPatternLayer(blockutils.BannerPatternTypeSkull, blockutils.DyeColorWhite),
	})
	w.tile = bannerTile

	f := newTestFloorBanner(w)
	f.ReadStateFromWorld()

	if f.GetColor() != blockutils.DyeColorBlue {
		t.Errorf("GetColor() = %v, want Blue", f.GetColor())
	}
	if len(f.GetPatterns()) != 1 || f.GetPatterns()[0].GetType() != blockutils.BannerPatternTypeSkull {
		t.Errorf("GetPatterns() = %v, want [Skull/White]", f.GetPatterns())
	}
}

func TestBaseBannerReadStateFromWorldSkipsOminousTile(t *testing.T) {
	w := &bannerTileWorld{}
	bannerTile := tile.NewBanner(nil, math.Vector3{})
	bannerTile.SetType(tile.BannerTypeOminous)
	bannerTile.SetBaseColor(blockutils.DyeColorRed)
	w.tile = bannerTile

	f := newTestFloorBanner(w)
	f.SetColor(blockutils.DyeColorWhite)
	f.ReadStateFromWorld()

	if f.GetColor() != blockutils.DyeColorWhite {
		t.Errorf("GetColor() = %v, want unchanged White (ominous tile should be skipped)", f.GetColor())
	}
}

func TestFloorBannerReadStateFromWorldReturnsOminousVersionForOminousTile(t *testing.T) {
	w := &bannerTileWorld{}
	bannerTile := tile.NewBanner(nil, math.Vector3{})
	bannerTile.SetType(tile.BannerTypeOminous)
	w.tile = bannerTile

	f := newTestFloorBanner(w)
	f.SetRotation(7)
	got := f.ReadStateFromWorld()

	ominous, ok := got.(*OminousFloorBanner)
	if !ok {
		t.Fatalf("expected ReadStateFromWorld to return a *OminousFloorBanner, got %T", got)
	}
	if ominous.GetRotation() != 7 {
		t.Errorf("GetRotation() = %d, want 7 (carried over from the original banner)", ominous.GetRotation())
	}
}

func TestWallBannerReadStateFromWorldReturnsOminousVersionForOminousTile(t *testing.T) {
	w := &bannerTileWorld{}
	bannerTile := tile.NewBanner(nil, math.Vector3{})
	bannerTile.SetType(tile.BannerTypeOminous)
	w.tile = bannerTile

	wb := newTestWallBanner(w)
	wb.SetFacing(math.East)
	got := wb.ReadStateFromWorld()

	ominous, ok := got.(*OminousWallBanner)
	if !ok {
		t.Fatalf("expected ReadStateFromWorld to return a *OminousWallBanner, got %T", got)
	}
	if ominous.GetFacing() != math.East {
		t.Errorf("GetFacing() = %v, want East (carried over from the original banner)", ominous.GetFacing())
	}
}

func TestFloorBannerPlaceOnlyOnUpFace(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}
	f := newTestFloorBanner(w)

	if f.Place(tx, fakeItem{}, f, f, math.East, math.Vector3{}, nil) {
		t.Error("expected Place to reject a non-Up face")
	}
	if !f.Place(tx, fakeItem{}, f, f, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to succeed on Up with center support below")
	}
}

func TestWallBannerPlaceRejectsVerticalFace(t *testing.T) {
	w := &candleWorld{}
	tx := &fakeBlockTransaction{}
	wb := newTestWallBanner(w)

	if wb.Place(tx, fakeItem{}, wb, wb, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to reject a vertical face")
	}
	if !wb.Place(tx, fakeItem{}, wb, wb, math.North, math.Vector3{}, nil) {
		t.Error("expected Place to succeed on a horizontal face with solid support")
	}
	if wb.Facing != math.North {
		t.Errorf("Facing = %v, want North", wb.Facing)
	}
}

func TestBannerOnNearbyBlockChangeBreaksWithoutSolidSupport(t *testing.T) {
	w := &noSupportWorld{}
	f := newTestFloorBanner(w)

	f.OnNearbyBlockChange()

	if len(w.breakCalls) != 1 {
		t.Fatalf("expected UseBreakOn to be called once, got %d", len(w.breakCalls))
	}
}
