package tile

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestChestDefaultName(t *testing.T) {
	w := &fakeWorld{}
	c := NewChest(w, math.Vector3{})
	if c.GetName() != "Chest" {
		t.Errorf("GetName() = %q, want %q", c.GetName(), "Chest")
	}
}

func TestChestCanOpenWithNoLock(t *testing.T) {
	w := &fakeWorld{}
	c := NewChest(w, math.Vector3{})
	if !c.CanOpenWith("anything") {
		t.Error("expected CanOpenWith to succeed when no lock is set")
	}
}

func TestChestCanOpenWithLock(t *testing.T) {
	w := &fakeWorld{}
	c := NewChest(w, math.Vector3{})
	c.Lock, c.HasLock = "secret", true

	if c.CanOpenWith("wrong") {
		t.Error("expected CanOpenWith to fail with the wrong key")
	}
	if !c.CanOpenWith("secret") {
		t.Error("expected CanOpenWith to succeed with the right key")
	}
}

func TestChestPairWithAdjacentChest(t *testing.T) {
	w := &fakeWorld{tiles: map[[3]int]Tile{}}
	a := NewChest(w, math.NewVector3(1, 2, 3))
	b := NewChest(w, math.NewVector3(1, 2, 4))
	w.tiles[[3]int{1, 2, 3}] = a
	w.tiles[[3]int{1, 2, 4}] = b

	if !a.PairWith(b) {
		t.Fatal("expected PairWith to succeed for adjacent chests")
	}
	if !a.IsPaired() || !b.IsPaired() {
		t.Fatal("expected both chests to report paired")
	}

	pair, ok := a.GetPair()
	if !ok || pair != b {
		t.Errorf("a.GetPair() = (%v, %v), want (b, true)", pair, ok)
	}
	pair, ok = b.GetPair()
	if !ok || pair != a {
		t.Errorf("b.GetPair() = (%v, %v), want (a, true)", pair, ok)
	}
}

func TestChestPairWithFailsIfAlreadyPaired(t *testing.T) {
	w := &fakeWorld{tiles: map[[3]int]Tile{}}
	a := NewChest(w, math.NewVector3(1, 2, 3))
	b := NewChest(w, math.NewVector3(1, 2, 4))
	c := NewChest(w, math.NewVector3(1, 2, 5))
	w.tiles[[3]int{1, 2, 3}] = a
	w.tiles[[3]int{1, 2, 4}] = b
	w.tiles[[3]int{1, 2, 5}] = c

	if !a.PairWith(b) {
		t.Fatal("expected the first PairWith to succeed")
	}
	if a.PairWith(c) {
		t.Error("expected PairWith to fail when a is already paired")
	}
}

func TestChestUnpair(t *testing.T) {
	w := &fakeWorld{tiles: map[[3]int]Tile{}}
	a := NewChest(w, math.NewVector3(1, 2, 3))
	b := NewChest(w, math.NewVector3(1, 2, 4))
	w.tiles[[3]int{1, 2, 3}] = a
	w.tiles[[3]int{1, 2, 4}] = b
	a.PairWith(b)

	if !a.Unpair() {
		t.Fatal("expected Unpair to succeed")
	}
	if a.IsPaired() || b.IsPaired() {
		t.Error("expected both chests to be unpaired")
	}
}

func TestChestSaveDataRoundTripWithPairing(t *testing.T) {
	w := &fakeWorld{tiles: map[[3]int]Tile{}}
	a := NewChest(w, math.NewVector3(1, 2, 3))
	b := NewChest(w, math.NewVector3(1, 2, 4))
	w.tiles[[3]int{1, 2, 3}] = a
	w.tiles[[3]int{1, 2, 4}] = b
	a.PairWith(b)
	a.SetName("Loot Chest")

	saved := a.SaveNBT()

	decoded := NewChest(w, math.NewVector3(1, 2, 3))
	if err := decoded.ReadSaveData(saved); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if !decoded.IsPaired() || decoded.PairX != 1 || decoded.PairZ != 4 {
		t.Errorf("decoded pairing = (%v, %d, %d), want (true, 1, 4)", decoded.IsPaired(), decoded.PairX, decoded.PairZ)
	}
	if decoded.GetName() != "Loot Chest" {
		t.Errorf("GetName() = %q, want %q", decoded.GetName(), "Loot Chest")
	}
}

func TestChestReadSaveDataRejectsNonAdjacentPair(t *testing.T) {
	w := &fakeWorld{}
	c := NewChest(w, math.NewVector3(1, 2, 3))

	tag := c.SaveNBT()
	tag.SetInt(ChestTagPairX, 10)
	tag.SetInt(ChestTagPairZ, 10)

	if err := c.ReadSaveData(tag); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if c.IsPaired() {
		t.Error("expected a non-adjacent pair position to be rejected")
	}
}
