package handler

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/network/mcpe"
)

func flags(set ...int) protocol.InputFlags {
	f := protocol.NewInputFlags(packet.InputFlagCount)
	for _, i := range set {
		f.Set(i)
	}
	return f
}

func TestResolveOnOffInputFlags(t *testing.T) {
	start, stop := packet.InputFlagStartSprinting, packet.InputFlagStopSprinting
	if v := resolveOnOffInputFlags(flags(start), start, stop); v == nil || !*v {
		t.Error("start flag alone should resolve to true")
	}
	if v := resolveOnOffInputFlags(flags(stop), start, stop); v == nil || *v {
		t.Error("stop flag alone should resolve to false")
	}
	if resolveOnOffInputFlags(flags(), start, stop) != nil || resolveOnOffInputFlags(flags(start, stop), start, stop) != nil {
		t.Error("neither or both flags should resolve to nil")
	}
}

func TestInputFlagsEqual(t *testing.T) {
	if !inputFlagsEqual(flags(packet.InputFlagJumping), flags(packet.InputFlagJumping)) {
		t.Error("equal flag sets compared unequal")
	}
	if inputFlagsEqual(flags(packet.InputFlagJumping), flags(packet.InputFlagSneaking)) {
		t.Error("different flag sets compared equal")
	}
}

func TestValidateFacing(t *testing.T) {
	for face := int32(0); face <= 5; face++ {
		if err := validateFacing(face); err != nil {
			t.Errorf("face %d rejected: %v", face, err)
		}
	}
	if validateFacing(-1) == nil || validateFacing(6) == nil {
		t.Error("an invalid face was accepted")
	}
}

func TestTranslateItemStackContainerID(t *testing.T) {
	tests := []struct {
		container      byte
		slot           int
		window, wantSl int
	}{
		{protocol.ContainerArmor, 2, mcpe.ContainerIDArmor, 2},
		{protocol.ContainerHotBar, 3, mcpe.ContainerIDInventory, 3},
		{protocol.ContainerCombinedHotBarAndInventory, 20, mcpe.ContainerIDInventory, 20},
		{protocol.ContainerOffhand, 1, mcpe.ContainerIDOffhand, 0},
		{protocol.ContainerCursor, 0, mcpe.ContainerIDUI, 0},
		{protocol.ContainerCraftingInput, 28, mcpe.ContainerIDUI, 28},
		{protocol.ContainerLevelEntity, 5, 7, 5},
	}
	for _, tt := range tests {
		window, slot, err := TranslateItemStackContainerID(tt.container, 7, tt.slot)
		if err != nil || window != tt.window || slot != tt.wantSl {
			t.Errorf("container %d slot %d: got (%d, %d, %v), want (%d, %d)", tt.container, tt.slot, window, slot, err, tt.window, tt.wantSl)
		}
	}
	if _, _, err := TranslateItemStackContainerID(protocol.ContainerCraftingOutputPreview, 7, 0); err == nil {
		t.Error("preview containers should be rejected")
	}
}

func TestJSONDepth(t *testing.T) {
	if d := jsonDepth(true); d != 0 {
		t.Errorf("scalar depth = %d", d)
	}
	if d := jsonDepth([]any{"a", 1.0}); d != 1 {
		t.Errorf("flat array depth = %d", d)
	}
	if d := jsonDepth([]any{[]any{1.0}}); d != 2 {
		t.Errorf("nested array depth = %d", d)
	}
}
