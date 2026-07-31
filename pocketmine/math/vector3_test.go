package math

import "testing"

func TestVector3Basics(t *testing.T) {
	a := NewVector3(1, 2, 3)
	b := NewVector3(4, 5, 6)

	if got := a.AddVector(b); got != (Vector3{5, 7, 9}) {
		t.Fatalf("AddVector() = %v, want {5 7 9}", got)
	}
	if got := a.Dot(b); got != 32 {
		t.Fatalf("Dot() = %v, want 32", got)
	}
	if got := a.Cross(b); got != (Vector3{2*6 - 3*5, 3*4 - 1*6, 1*5 - 2*4}) {
		t.Fatalf("Cross() = %v", got)
	}
	if !a.Equals(NewVector3(1, 2, 3)) {
		t.Fatalf("Equals() should be true for identical components")
	}
}

func TestVector3GetSide(t *testing.T) {
	origin := Vector3Zero()
	if got := origin.Up(1); got != (Vector3{0, 1, 0}) {
		t.Fatalf("Up(1) = %v, want {0 1 0}", got)
	}
	if got := origin.East(2); got != (Vector3{2, 0, 0}) {
		t.Fatalf("East(2) = %v, want {2 0 0}", got)
	}
}

func TestVector3GetIntermediateWithXValue(t *testing.T) {
	a := NewVector3(0, 0, 0)
	b := NewVector3(10, 10, 10)
	v, ok := a.GetIntermediateWithXValue(b, 5)
	if !ok {
		t.Fatalf("expected an intermediate point to exist")
	}
	if v != (Vector3{5, 5, 5}) {
		t.Fatalf("GetIntermediateWithXValue() = %v, want {5 5 5}", v)
	}
}

func TestFacingOppositeAndRotate(t *testing.T) {
	if Opposite(North) != South {
		t.Fatalf("Opposite(North) should be South")
	}
	if RotateY(North, true) != East {
		t.Fatalf("RotateY(North, clockwise) should be East")
	}
}

func TestAxisAlignedBBIntersectsAndVolume(t *testing.T) {
	a, err := NewAxisAlignedBB(0, 0, 0, 1, 1, 1)
	if err != nil {
		t.Fatalf("NewAxisAlignedBB() error = %v", err)
	}
	b, _ := NewAxisAlignedBB(0.5, 0.5, 0.5, 1.5, 1.5, 1.5)
	if !a.IntersectsWith(b, 0.00001) {
		t.Fatalf("expected overlapping boxes to intersect")
	}
	if got := a.GetVolume(); got != 1 {
		t.Fatalf("GetVolume() = %v, want 1", got)
	}

	expanded := a.ExpandedCopy(1, 1, 1)
	if expanded.MinX != -1 || expanded.MaxX != 2 {
		t.Fatalf("ExpandedCopy() = %v, want MinX=-1 MaxX=2", expanded)
	}
	if a.MinX != 0 {
		t.Fatalf("ExpandedCopy() must not mutate the receiver, got MinX=%v", a.MinX)
	}
}

func TestVoxelRayTraceBetweenPoints(t *testing.T) {
	seq, err := BetweenPoints(NewVector3(0.5, 0.5, 0.5), NewVector3(3.5, 0.5, 0.5))
	if err != nil {
		t.Fatalf("BetweenPoints() error = %v", err)
	}
	var got []Vector3
	for v := range seq {
		got = append(got, v)
	}
	want := []Vector3{{0, 0, 0}, {1, 0, 0}, {2, 0, 0}, {3, 0, 0}}
	if len(got) != len(want) {
		t.Fatalf("BetweenPoints() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("BetweenPoints()[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}
