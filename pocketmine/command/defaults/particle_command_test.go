package defaults

import (
	"strings"
	"testing"

	"pocketmine-go/pocketmine/world/particle"
)

func TestGetParticle(t *testing.T) {
	str := func(s string) *string { return &s }
	if _, ok := getParticle("crit", nil).(particle.CriticalParticle); !ok {
		t.Error("crit")
	}
	if p := getParticle("reddust", nil).(particle.RedstoneParticle); p.Lifetime != 1 {
		t.Errorf("reddust lifetime = %d, want the default 1", p.Lifetime)
	}
	if p := getParticle("blockdust", str("300_2_3")).(particle.DustParticle); p.Color.GetR() != 300&0xff || p.Color.GetA() != 255 {
		t.Errorf("blockdust colour = %v", p.Color)
	}
	if p, ok := getParticle("terrain", str("stone")).(particle.TerrainParticle); !ok || p.BlockStateID == 0 {
		t.Errorf("terrain stone = %v", p)
	}
	if getParticle("terrain", str("diamond")) != nil {
		t.Error("an item with no block made a terrain particle")
	}
	if p, ok := getParticle("itembreak", str("diamond")).(particle.ItemBreakParticle); !ok || p.NetworkID == 0 {
		t.Errorf("itembreak diamond = %v", p)
	}
	if getParticle("nope", nil) != nil {
		t.Error("unknown particle")
	}
}

func TestParticleCommand(t *testing.T) {
	s, sender := newTestServer(t)
	s.DispatchCommand(sender, "particle flame 0 64 0 1 1 1 5", false)
	if msg := lastMessage(sender); !strings.Contains(msg, "flame") || !strings.Contains(msg, "5") {
		t.Errorf("success message = %q", msg)
	}
	s.DispatchCommand(sender, "particle nope 0 64 0 1 1 1", false)
	if msg := lastMessage(sender); !strings.Contains(msg, "nope") {
		t.Errorf("not found message = %q", msg)
	}
}
