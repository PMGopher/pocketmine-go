package particle

import "testing"

func TestNewDragonEggTeleportParticleAcceptsBoundaryValues(t *testing.T) {
	p, err := NewDragonEggTeleportParticle(-255, 0, 255)
	if err != nil {
		t.Fatalf("NewDragonEggTeleportParticle(-255,0,255) returned an error: %v", err)
	}
	if p.XDiff != -255 || p.YDiff != 0 || p.ZDiff != 255 {
		t.Errorf("got %+v, want XDiff=-255 YDiff=0 ZDiff=255", p)
	}
}

func TestNewDragonEggTeleportParticleRejectsOutOfRangeValues(t *testing.T) {
	cases := [][3]int{
		{256, 0, 0},
		{0, 256, 0},
		{0, 0, 256},
		{-256, 0, 0},
	}
	for _, c := range cases {
		if _, err := NewDragonEggTeleportParticle(c[0], c[1], c[2]); err == nil {
			t.Errorf("NewDragonEggTeleportParticle%v = nil error, want an error", c)
		}
	}
}

func TestDefaultPotionSplashColorMatchesTheRealVanillaValue(t *testing.T) {
	if got := DefaultPotionSplashColor.ToARGB(); got&0xFFFFFF != 0x385dc6 {
		t.Errorf("DefaultPotionSplashColor.ToARGB() & 0xFFFFFF = %#x, want %#x", got&0xFFFFFF, 0x385dc6)
	}
}
