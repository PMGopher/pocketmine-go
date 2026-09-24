package handler

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
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

func TestValidFacing(t *testing.T) {
	for face := int32(0); face <= 5; face++ {
		if !validFacing(face) {
			t.Errorf("face %d rejected", face)
		}
	}
	if validFacing(-1) || validFacing(6) {
		t.Error("an invalid face was accepted")
	}
}
