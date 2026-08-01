package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestChemistryTable(w World) *ChemistryTable {
	c := NewChemistryTable(mustBlockIdentifier(1074), "Test Chemistry Table", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func TestChemistryTablePlaceFacesOppositePlayer(t *testing.T) {
	w := &fakeWorld{}
	c := newTestChemistryTable(w)
	tx := &fakeBlockTransaction{}
	player := &fakeSignPlayer{}

	c.Place(tx, fakeItem{}, c, c, math.Up, math.Vector3{}, player)

	if c.Facing != math.Opposite(player.GetHorizontalFacing()) {
		t.Errorf("Facing = %v, want opposite of player facing (%v)", c.Facing, math.Opposite(player.GetHorizontalFacing()))
	}
}

func TestChemistryTableOnInteractReturnsFalse(t *testing.T) {
	w := &fakeWorld{}
	c := newTestChemistryTable(w)

	if c.OnInteract(fakeItem{}, math.Up, math.Vector3{}, &fakeSignPlayer{}, nil) {
		t.Error("expected OnInteract to return false (unimplemented in PHP original too)")
	}
}
