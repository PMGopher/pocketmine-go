package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

// chestWorld backs GetTile/GetBlockAt with simple coordinate maps, and satisfies tile.World too
// (via its own GetTileAt) so tile.Chest's pairing logic can look itself up.
type chestWorld struct {
	fakeWorld
	tiles      map[[3]int]Tile
	blockTiles map[[3]int]tile.Tile
	blocks     map[[3]int]Behavior
}

func (w *chestWorld) GetTile(pos Position) (Tile, bool) {
	t, ok := w.tiles[[3]int{pos.FloorX(), pos.FloorY(), pos.FloorZ()}]
	return t, ok
}

func (w *chestWorld) GetTileAt(x, y, z int) (tile.Tile, bool) {
	t, ok := w.blockTiles[[3]int{x, y, z}]
	return t, ok
}

func (w *chestWorld) RemoveTile(t tile.Tile) {}

func (w *chestWorld) GetBlockAt(x, y, z int) Behavior {
	if b, ok := w.blocks[[3]int{x, y, z}]; ok {
		return b
	}
	filler := newTestBlock(true) // transparent filler, e.g. air above a chest
	filler.SetPosition(w, x, y, z)
	return filler
}

func newTestChest(w World, x, y, z int) *Chest {
	c := NewChest(mustBlockIdentifier(1065), "Test Chest", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, x, y, z)
	return c
}

// placeChestWithTile constructs a Chest block + backing tile.Chest at the given position and
// registers both in the world's lookup maps, mirroring what World.SetBlock/AddTile would do.
func placeChestWithTile(w *chestWorld, x, y, z int, facing math.Facing) *Chest {
	c := newTestChest(w, x, y, z)
	c.Facing = facing
	t := tile.NewChest(w, math.NewVector3(float64(x), float64(y), float64(z)))
	w.tiles[[3]int{x, y, z}] = t
	w.blockTiles[[3]int{x, y, z}] = t
	w.blocks[[3]int{x, y, z}] = c
	return c
}

func TestChestPlaceFacesOppositePlayer(t *testing.T) {
	w := &chestWorld{tiles: map[[3]int]Tile{}, blockTiles: map[[3]int]tile.Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestChest(w, 1, 2, 3)
	tx := &fakeBlockTransaction{}

	player := &fakeSignPlayer{}
	c.Place(tx, fakeItem{}, c, c, math.Up, math.Vector3{}, player)

	if c.Facing != math.Opposite(player.GetHorizontalFacing()) {
		t.Errorf("Facing = %v, want opposite of player facing (%v)", c.Facing, math.Opposite(player.GetHorizontalFacing()))
	}
}

func TestChestOnPostPlacePairsWithAdjacentChest(t *testing.T) {
	w := &chestWorld{tiles: map[[3]int]Tile{}, blockTiles: map[[3]int]tile.Tile{}, blocks: map[[3]int]Behavior{}}
	// A North-facing chest pairs on its East/West sides (RotateY(North, ·) == East/West), so these
	// need to be X-adjacent (same Y/Z) - not Z-adjacent.
	placeChestWithTile(w, 1, 2, 3, math.North)
	c2 := placeChestWithTile(w, 2, 2, 3, math.North) // West of c2 is (1,2,3): RotateY(North, false) == West

	c2.OnPostPlace()

	t1, _ := w.GetTile(NewPosition(1, 2, 3, w))
	t2, _ := w.GetTile(NewPosition(2, 2, 3, w))
	tc1, tc2 := t1.(*tile.Chest), t2.(*tile.Chest)
	if !tc1.IsPaired() || !tc2.IsPaired() {
		t.Fatal("expected both chest tiles to be paired after OnPostPlace")
	}
}

func TestChestOnInteractBlockedByOpaqueLid(t *testing.T) {
	w := &chestWorld{tiles: map[[3]int]Tile{}, blockTiles: map[[3]int]tile.Tile{}, blocks: map[[3]int]Behavior{}}
	c := placeChestWithTile(w, 1, 2, 3, math.North)

	opaque := newTestBlock(false)
	opaque.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = opaque

	// OnInteract always returns true (matching PHP) regardless of whether it actually opens -
	// what we can verify is that it doesn't panic and completes with a present player.
	if !c.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Error("expected OnInteract to return true even when blocked")
	}
}

func TestChestGetFuelTime(t *testing.T) {
	w := &chestWorld{tiles: map[[3]int]Tile{}, blockTiles: map[[3]int]tile.Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestChest(w, 1, 2, 3)
	if c.GetFuelTime() != 300 {
		t.Errorf("GetFuelTime() = %d, want 300", c.GetFuelTime())
	}
}
