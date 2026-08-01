package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

// containerTileWorld wraps chestWorld's coordinate-map approach but for the general "has a
// container tile at a fixed position" shape shared by Furnace/Barrel/ShulkerBox/BrewingStand.
type containerTileWorld struct {
	fakeWorld
	tiles  map[[3]int]Tile
	blocks map[[3]int]Behavior
}

func (w *containerTileWorld) GetTile(pos Position) (Tile, bool) {
	t, ok := w.tiles[[3]int{pos.FloorX(), pos.FloorY(), pos.FloorZ()}]
	return t, ok
}

func (w *containerTileWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := newTestBlock(true) // transparent filler by default
	filler.SetPosition(w, x, y, z)
	return filler
}

// RemoveTile/GetTileAt satisfy tile.World, needed to construct real tile.* values in tests.
func (w *containerTileWorld) RemoveTile(t tile.Tile) {}
func (w *containerTileWorld) GetTileAt(x, y, z int) (tile.Tile, bool) {
	t, ok := w.tiles[[3]int{x, y, z}]
	if !ok {
		return nil, false
	}
	tt, ok := t.(tile.Tile)
	return tt, ok
}

func newTestFurnace(w World) *Furnace {
	f := NewFurnace(mustBlockIdentifier(1066), "Test Furnace", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), tile.FurnaceTypeFurnace)
	f.SetPosition(w, 1, 2, 3)
	return f
}

func TestFurnaceGetLightLevel(t *testing.T) {
	w := &fakeWorld{}
	f := newTestFurnace(w)
	if f.GetLightLevel() != 0 {
		t.Errorf("GetLightLevel() = %d, want 0 while unlit", f.GetLightLevel())
	}
	f.SetLit(true)
	if f.GetLightLevel() != 13 {
		t.Errorf("GetLightLevel() = %d, want 13 while lit", f.GetLightLevel())
	}
}

func TestFurnaceGetFurnaceType(t *testing.T) {
	w := &fakeWorld{}
	f := NewFurnace(mustBlockIdentifier(1066), "Test Blast Furnace", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil), tile.FurnaceTypeBlastFurnace)
	f.SetPosition(w, 1, 2, 3)
	if f.GetFurnaceType() != tile.FurnaceTypeBlastFurnace {
		t.Errorf("GetFurnaceType() = %v, want BlastFurnace", f.GetFurnaceType())
	}
}

func TestFurnacePlaceFacesOppositePlayer(t *testing.T) {
	w := &fakeWorld{}
	f := newTestFurnace(w)
	tx := &fakeBlockTransaction{}
	player := &fakeSignPlayer{}

	f.Place(tx, fakeItem{}, f, f, math.Up, math.Vector3{}, player)

	if f.Facing != math.Opposite(player.GetHorizontalFacing()) {
		t.Errorf("Facing = %v, want opposite of player facing", f.Facing)
	}
}

func TestFurnaceOnInteractCompletesWithTile(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}}
	f := newTestFurnace(w)
	w.tiles[[3]int{1, 2, 3}] = tile.NewFurnace(w, math.NewVector3(1, 2, 3))

	if !f.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Error("expected OnInteract to return true")
	}
}
