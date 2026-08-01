package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestWall() *Wall {
	idInfo, err := NewBlockIdentifier(1004, nil)
	if err != nil {
		panic(err)
	}
	return NewWall(idInfo, "Test Wall", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
}

func TestWallConnectionsRoundTrip(t *testing.T) {
	w := newTestWall()
	w.SetConnection(math.North, blockutils.WallConnectionTypeTall, true)
	w.SetConnection(math.South, blockutils.WallConnectionTypeShort, true)
	w.SetPost(true)
	// West/East deliberately left unconnected.

	data, err := w.encodeBlockOnlyState()
	if err != nil {
		t.Fatalf("encodeBlockOnlyState: %v", err)
	}

	decoded := newTestWall()
	if err := decoded.DecodeBlockOnlyState(data); err != nil {
		t.Fatalf("DecodeBlockOnlyState: %v", err)
	}

	if ty, ok := decoded.GetConnection(math.North); !ok || ty != blockutils.WallConnectionTypeTall {
		t.Errorf("North connection = (%v, %v), want (Tall, true)", ty, ok)
	}
	if ty, ok := decoded.GetConnection(math.South); !ok || ty != blockutils.WallConnectionTypeShort {
		t.Errorf("South connection = (%v, %v), want (Short, true)", ty, ok)
	}
	if _, ok := decoded.GetConnection(math.West); ok {
		t.Error("West should have no connection")
	}
	if _, ok := decoded.GetConnection(math.East); ok {
		t.Error("East should have no connection")
	}
	if !decoded.IsPost() {
		t.Error("Post should be true")
	}
}
