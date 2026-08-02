package block

import "testing"

func newTestPumpkinStem(w World) *PumpkinStem {
	idInfo, err := NewBlockIdentifier(1011, nil)
	if err != nil {
		panic(err)
	}
	p := NewPumpkinStem(idInfo, "Test Pumpkin Stem", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	p.SetPosition(w, 1, 2, 3)
	return p
}

func TestPumpkinStemGetPlantTypeID(t *testing.T) {
	w := &stemWorld{}
	p := newTestPumpkinStem(w)
	if p.GetPlantTypeID() != PUMPKIN {
		t.Errorf("GetPlantTypeID() = %d, want PUMPKIN (%d)", p.GetPlantTypeID(), PUMPKIN)
	}
}

func TestPumpkinStemGetPlantReturnsAPumpkinBlock(t *testing.T) {
	w := &stemWorld{}
	p := newTestPumpkinStem(w)
	if got := p.GetPlant(); got.GetTypeId() != PUMPKIN {
		t.Errorf("GetPlant().GetTypeId() = %d, want PUMPKIN (%d)", got.GetTypeId(), PUMPKIN)
	}
}
