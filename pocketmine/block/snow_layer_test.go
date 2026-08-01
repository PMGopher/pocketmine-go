package block

import (
	"testing"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

func newTestSnowLayer(w World) *SnowLayer {
	idInfo, err := NewBlockIdentifier(1017, nil)
	if err != nil {
		panic(err)
	}
	s := NewSnowLayer(idInfo, "Test Snow Layer", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, 1, 2, 3)
	return s
}

func TestSnowLayerStacksOnExistingLayer(t *testing.T) {
	w := &candleWorld{}
	existing := newTestSnowLayer(w)
	existing.Layers = 3

	next := newTestSnowLayer(w)
	tx := &fakeBlockTransaction{}
	if !next.Place(tx, fakeItem{}, existing, existing, math.Up, math.Vector3{}, nil) {
		t.Fatal("expected Place to succeed when stacking onto an existing snow layer")
	}
	if next.Layers != 4 {
		t.Errorf("Layers = %d, want 4", next.Layers)
	}
}

func TestSnowLayerCannotStackPastMax(t *testing.T) {
	w := &candleWorld{}
	existing := newTestSnowLayer(w)
	existing.Layers = snowLayerMaxLayers

	next := newTestSnowLayer(w)
	tx := &fakeBlockTransaction{}
	if next.Place(tx, fakeItem{}, existing, existing, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to fail when the existing layer is already at max")
	}
}

func TestSnowLayerGetSupportTypeFullOnlyAtMaxLayers(t *testing.T) {
	w := &candleWorld{}
	s := newTestSnowLayer(w)

	s.Layers = snowLayerMaxLayers - 1
	if s.GetSupportType(math.Up) != blockutils.SupportTypeNone {
		t.Error("expected SupportTypeNone below max layers")
	}

	s.Layers = snowLayerMaxLayers
	if s.GetSupportType(math.Up) != blockutils.SupportTypeFull {
		t.Error("expected SupportTypeFull at max layers")
	}
}
