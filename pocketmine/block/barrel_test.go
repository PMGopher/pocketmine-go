package block

import (
	"testing"

	"pocketmine-go/pocketmine/block/tile"
	"pocketmine-go/pocketmine/math"
)

func newTestBarrel(w World) *Barrel {
	b := NewBarrel(mustBlockIdentifier(1067), "Test Barrel", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	b.SetPosition(w, 1, 2, 3)
	return b
}

// barrelPlayer is a fakeSignPlayer with a settable eye Y, for exercising Barrel.Place's
// look-direction branches.
type barrelPlayer struct {
	fakeSignPlayer
	pos    math.Vector3
	eyePos math.Vector3
}

func (p barrelPlayer) GetPosition() math.Vector3 { return p.pos }
func (p barrelPlayer) GetEyePos() math.Vector3   { return p.eyePos }

func TestBarrelPlaceFacesUpWhenPlayerLooksUpFromBelow(t *testing.T) {
	w := &fakeWorld{}
	b := newTestBarrel(w)
	tx := &fakeBlockTransaction{}
	// barrel at (1,2,3); player standing close, eye Y more than 2 above the barrel
	player := barrelPlayer{pos: math.Vector3{X: 1, Y: 0, Z: 3}, eyePos: math.Vector3{X: 1, Y: 4.5, Z: 3}}

	b.Place(tx, fakeItem{}, b, b, math.Up, math.Vector3{}, player)

	if b.Facing != math.Up {
		t.Errorf("Facing = %v, want Up", b.Facing)
	}
}

func TestBarrelPlaceFacesDownWhenPlayerLooksDownFromAbove(t *testing.T) {
	w := &fakeWorld{}
	b := newTestBarrel(w)
	tx := &fakeBlockTransaction{}
	// barrel at Y=2; player's eye level below that (but not more than 2 above, so the Up branch
	// doesn't fire either) triggers the Down branch.
	player := barrelPlayer{pos: math.Vector3{X: 1, Y: 2, Z: 3}, eyePos: math.Vector3{X: 1, Y: 1.0, Z: 3}}

	b.Place(tx, fakeItem{}, b, b, math.Up, math.Vector3{}, player)

	if b.Facing != math.Down {
		t.Errorf("Facing = %v, want Down", b.Facing)
	}
}

func TestBarrelPlaceFacesOppositePlayerWhenLevelOrFar(t *testing.T) {
	w := &fakeWorld{}
	b := newTestBarrel(w)
	tx := &fakeBlockTransaction{}
	// far away (>= 2 blocks on X/Z) - falls back to horizontal opposite regardless of eye height
	player := barrelPlayer{pos: math.Vector3{X: 10, Y: 2, Z: 10}, eyePos: math.Vector3{X: 10, Y: 10, Z: 10}}

	b.Place(tx, fakeItem{}, b, b, math.Up, math.Vector3{}, player)

	if b.Facing != math.Opposite(player.GetHorizontalFacing()) {
		t.Errorf("Facing = %v, want opposite of player horizontal facing", b.Facing)
	}
}

func TestBarrelOnInteractCompletesWithTile(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}}
	b := newTestBarrel(w)
	w.tiles[[3]int{1, 2, 3}] = tile.NewBarrel(w, math.NewVector3(1, 2, 3))

	if !b.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Error("expected OnInteract to return true")
	}
}

func TestBarrelGetFuelTime(t *testing.T) {
	w := &fakeWorld{}
	b := newTestBarrel(w)
	if b.GetFuelTime() != 300 {
		t.Errorf("GetFuelTime() = %d, want 300", b.GetFuelTime())
	}
}
