package tile

import (
	"testing"

	"pocketmine-go/pocketmine/math"
)

func TestBarrelDefaultName(t *testing.T) {
	w := &fakeWorld{}
	b := NewBarrel(w, math.Vector3{})
	if b.GetName() != "Barrel" {
		t.Errorf("GetName() = %q, want %q", b.GetName(), "Barrel")
	}
}

func TestBarrelNameRoundTripsThroughNBT(t *testing.T) {
	w := &fakeWorld{}
	b := NewBarrel(w, math.Vector3{})
	b.SetName("Storage")

	decoded := NewBarrel(w, math.Vector3{})
	if err := decoded.ReadSaveData(b.SaveNBT()); err != nil {
		t.Fatalf("ReadSaveData: %v", err)
	}
	if decoded.GetName() != "Storage" {
		t.Errorf("GetName() = %q, want %q", decoded.GetName(), "Storage")
	}
}

func TestBarrelCanOpenWithLock(t *testing.T) {
	w := &fakeWorld{}
	b := NewBarrel(w, math.Vector3{})
	b.Lock, b.HasLock = "key", true

	if b.CanOpenWith("wrong") {
		t.Error("expected CanOpenWith to fail with the wrong key")
	}
	if !b.CanOpenWith("key") {
		t.Error("expected CanOpenWith to succeed with the right key")
	}
}

func TestEnderChestViewerCount(t *testing.T) {
	w := &fakeWorld{}
	e := NewEnderChest(w, math.Vector3{})
	if e.GetViewerCount() != 0 {
		t.Errorf("GetViewerCount() = %d, want 0", e.GetViewerCount())
	}
	e.SetViewerCount(3)
	if e.GetViewerCount() != 3 {
		t.Errorf("GetViewerCount() = %d, want 3", e.GetViewerCount())
	}
}

func TestEnderChestSetViewerCountRejectsNegative(t *testing.T) {
	w := &fakeWorld{}
	e := NewEnderChest(w, math.Vector3{})
	defer func() {
		if recover() == nil {
			t.Error("expected SetViewerCount to panic for a negative value")
		}
	}()
	e.SetViewerCount(-1)
}
