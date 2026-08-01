package blockutils

import "testing"

func TestChiseledBookshelfSlotFromBlockFaceCoordinates(t *testing.T) {
	cases := []struct {
		x, y float64
		want ChiseledBookshelfSlot
	}{
		{0, 0.9, ChiseledBookshelfSlotTopLeft},
		{0.5, 0.9, ChiseledBookshelfSlotTopMiddle},
		{0.9, 0.9, ChiseledBookshelfSlotTopRight},
		{0, 0.1, ChiseledBookshelfSlotBottomLeft},
		{0.5, 0.1, ChiseledBookshelfSlotBottomMiddle},
		{0.9, 0.1, ChiseledBookshelfSlotBottomRight},
		// exact grid boundaries (6/16 and 11/16)
		{6.0 / 16, 0.9, ChiseledBookshelfSlotTopMiddle},
		{11.0 / 16, 0.9, ChiseledBookshelfSlotTopRight},
	}
	for _, c := range cases {
		if got := ChiseledBookshelfSlotFromBlockFaceCoordinates(c.x, c.y); got != c.want {
			t.Errorf("FromBlockFaceCoordinates(%v, %v) = %v, want %v", c.x, c.y, got, c.want)
		}
	}
}

func TestChiseledBookshelfSlotFromBlockFaceCoordinatesPanicsOutOfRange(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected an out-of-range x to panic")
		}
	}()
	ChiseledBookshelfSlotFromBlockFaceCoordinates(1.5, 0.5)
}
