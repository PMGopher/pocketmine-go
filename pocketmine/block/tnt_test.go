package block

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func newTestTNT(w World) *TNTBlock {
	idInfo, err := NewBlockIdentifier(1008, nil)
	if err != nil {
		panic(err)
	}
	t := NewTNT(idInfo, "Test TNT", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	t.SetPosition(w, 1, 2, 3)
	return t
}

func TestTNTFlintAndSteelHandlesInteractAndDamagesDurable(t *testing.T) {
	w := &fakeWorld{}
	tnt := newTestTNT(w)

	axe := &fakeAxeItem{fakeItem: fakeItem{typeID: itemTypeIDsFlintAndSteel}}
	if !tnt.OnInteract(axe, 0, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to handle flint and steel")
	}
	if axe.damage != 1 {
		t.Errorf("durable damage = %d, want 1", axe.damage)
	}
}

func TestTNTFireChargePopsItemAndHandlesInteract(t *testing.T) {
	w := &fakeWorld{}
	tnt := newTestTNT(w)

	fireCharge := fakeItem{typeID: itemTypeIDsFireCharge}
	if !tnt.OnInteract(fireCharge, 0, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to handle a fire charge")
	}
}

func TestTNTIrrelevantItemDoesNotIgnite(t *testing.T) {
	w := &fakeWorld{}
	tnt := newTestTNT(w)

	if tnt.OnInteract(fakeItem{}, 0, math.Vector3{}, nil, nil) {
		t.Error("expected an unrelated item not to be handled")
	}
}

func TestTNTUnstableIgnitesOnBreak(t *testing.T) {
	w := &fakeWorld{}
	tnt := newTestTNT(w)
	tnt.Unstable = true

	if !tnt.OnBreak(fakeItem{}, nil, nil) {
		t.Fatal("expected an unstable TNT's OnBreak to report the ignite as handled")
	}
}

func TestTNTStableOnBreakFallsThroughToDefault(t *testing.T) {
	w := &fakeWorld{}
	tnt := newTestTNT(w)

	// Block's default OnBreak returns true (break proceeds normally) - a stable TNT should just
	// fall through to that, without the unstable-ignite special case.
	if !tnt.OnBreak(fakeItem{}, nil, nil) {
		t.Error("expected a stable TNT's OnBreak to fall through to the default")
	}
}
