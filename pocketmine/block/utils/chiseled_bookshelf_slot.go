package blockutils

import "fmt"

// ChiseledBookshelfSlot is a port of pocketmine\block\utils\ChiseledBookshelfSlot.
type ChiseledBookshelfSlot int

const (
	ChiseledBookshelfSlotTopLeft ChiseledBookshelfSlot = iota
	ChiseledBookshelfSlotTopMiddle
	ChiseledBookshelfSlotTopRight
	ChiseledBookshelfSlotBottomLeft
	ChiseledBookshelfSlotBottomMiddle
	ChiseledBookshelfSlotBottomRight
)

const chiseledBookshelfSlotsPerShelf = 3

// ChiseledBookshelfSlotFromBlockFaceCoordinates is a port of
// ChiseledBookshelfSlot::fromBlockFaceCoordinates. Panics on an out-of-range coordinate, matching
// the PHP original's InvalidArgumentException (same convention as e.g. Hopper.SetFacing).
func ChiseledBookshelfSlotFromBlockFaceCoordinates(x, y float64) ChiseledBookshelfSlot {
	if x < 0 || x > 1 {
		panic(fmt.Sprintf("X must be between 0 and 1, got %v", x))
	}
	if y < 0 || y > 1 {
		panic(fmt.Sprintf("Y must be between 0 and 1, got %v", y))
	}

	slot := 0
	if y < 0.5 {
		slot = chiseledBookshelfSlotsPerShelf
	}
	switch {
	case x < 6.0/16:
		slot += 0
	case x < 11.0/16:
		slot += 1
	default:
		slot += 2
	}
	return ChiseledBookshelfSlot(slot)
}
