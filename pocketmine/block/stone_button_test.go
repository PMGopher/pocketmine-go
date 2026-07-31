package block

import "testing"

func TestStoneButtonSatisfiesBehavior(t *testing.T) {
	idInfo, err := NewBlockIdentifier(int(1234), nil)
	if err != nil {
		t.Fatal(err)
	}
	breakInfo := BlockBreakInfoPickaxe(0.5, nil, nil)
	typeInfo := NewBlockTypeInfo(breakInfo, nil, nil)

	btn := NewStoneButton(idInfo, "Stone Button", typeInfo)

	if btn.ActivationTime != 20 {
		t.Errorf("ActivationTime = %d, want 20", btn.ActivationTime)
	}
	if btn.IsTransparent() != true {
		t.Error("StoneButton should be transparent (inherited from Transparent)")
	}
	if btn.CanBeFlowedInto() != true {
		t.Error("StoneButton should be flowable (inherited from Flowable)")
	}

	var _ Behavior = btn // compile-time proof StoneButton fully satisfies Behavior

	clone := btn.Clone().(*StoneButton)
	clone.SetPressed(true)
	if btn.IsPressed() {
		t.Error("cloning StoneButton leaked state back into the original")
	}
}
