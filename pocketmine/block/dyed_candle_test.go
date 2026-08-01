package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestDyedCandle(w World, color blockutils.DyeColor) *DyedCandle {
	idInfo, err := NewBlockIdentifier(1005, nil)
	if err != nil {
		panic(err)
	}
	d := NewDyedCandle(idInfo, "Test Dyed Candle", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	d.SetColor(color)
	d.SetPosition(w, 1, 2, 3)
	return d
}

func TestDyedCandleStacksOnlyWithSameColor(t *testing.T) {
	w := &candleWorld{}
	red := newTestDyedCandle(w, blockutils.DyeColorRed)

	sameColor := newTestDyedCandle(w, blockutils.DyeColorRed)
	if !sameColor.CanBePlacedAt(red, math.Vector3{}, math.Up, true) {
		t.Error("expected same-colour dyed candles to be stackable")
	}

	differentColor := newTestDyedCandle(w, blockutils.DyeColorBlue)
	if differentColor.CanBePlacedAt(red, math.Vector3{}, math.Up, true) {
		t.Error("expected different-coloured dyed candles not to be stackable")
	}
}

func TestDyedCandleDoesNotStackWithPlainCandle(t *testing.T) {
	w := &candleWorld{}
	plain := newTestCandle(w)
	dyed := newTestDyedCandle(w, blockutils.DyeColorRed)

	if dyed.CanBePlacedAt(plain, math.Vector3{}, math.Up, true) {
		t.Error("expected a dyed candle not to stack onto a plain (undyed) candle")
	}
}
