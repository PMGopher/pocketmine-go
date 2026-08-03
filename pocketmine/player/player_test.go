package player

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/generator"
)

func newTestWorld(t *testing.T) *world.World {
	t.Helper()
	tr := convert.NewBlockTranslator()
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), nil)
	return world.New(gen, tr, []block.Behavior{
		block.VanillaAir(),
		block.VanillaBedrock(),
		block.VanillaStone(),
		block.VanillaDirt(),
		block.VanillaGrass(),
	})
}

func newTestPlayer(t *testing.T, id int, pos math.Vector3) *Player {
	t.Helper()
	w := newTestWorld(t)
	return NewPlayer(id, "Steve", "uuid-"+string(rune(id)), "xuid-1", w, pos, GameModeSurvival)
}

func TestNewPlayerSetsIdentityFields(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(5, 70, 5))
	if got := p.GetName(); got != "Steve" {
		t.Errorf("GetName() = %q, want %q", got, "Steve")
	}
	if got := p.GetDisplayName(); got != "Steve" {
		t.Errorf("GetDisplayName() defaults to %q, want %q (the username)", got, "Steve")
	}
	if got := p.GetXuid(); got != "xuid-1" {
		t.Errorf("GetXuid() = %q, want %q", got, "xuid-1")
	}
	if got := p.GetID(); got != 1 {
		t.Errorf("GetID() = %d, want 1", got)
	}
}

func TestSetDisplayNameOverridesIt(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	p.SetDisplayName("Notch")
	if got := p.GetDisplayName(); got != "Notch" {
		t.Errorf("GetDisplayName() after SetDisplayName = %q, want %q", got, "Notch")
	}
}

func TestGamemodeChecksAgreeWithRealPHPsNonLiteralSemantics(t *testing.T) {
	cases := []struct {
		mode                                     GameMode
		survival, creative, adventure, spectator bool
	}{
		{GameModeSurvival, true, false, false, false},
		{GameModeCreative, false, true, false, false},
		{GameModeAdventure, true, false, true, false},
		{GameModeSpectator, false, true, true, true},
	}
	for _, c := range cases {
		p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
		p.SetGamemode(c.mode)
		if got := p.IsSurvival(); got != c.survival {
			t.Errorf("%v: IsSurvival() = %v, want %v", c.mode, got, c.survival)
		}
		if got := p.IsCreative(); got != c.creative {
			t.Errorf("%v: IsCreative() = %v, want %v", c.mode, got, c.creative)
		}
		if got := p.IsAdventure(); got != c.adventure {
			t.Errorf("%v: IsAdventure() = %v, want %v", c.mode, got, c.adventure)
		}
		if got := p.IsSpectator(); got != c.spectator {
			t.Errorf("%v: IsSpectator() = %v, want %v", c.mode, got, c.spectator)
		}
	}
}

func TestGetSpawnFallsBackToWorldSpawnWhenUnset(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	worldSpawn := math.NewVector3(100, 65, 100)
	p.GetWorld().SetSpawnLocation(worldSpawn)

	if got := p.GetSpawn(); got != worldSpawn {
		t.Errorf("GetSpawn() with no player-specific spawn set = %v, want the world spawn %v", got, worldSpawn)
	}

	own := math.NewVector3(5, 80, 5)
	p.SetSpawn(own)
	if got := p.GetSpawn(); got != own {
		t.Errorf("GetSpawn() after SetSpawn = %v, want %v", got, own)
	}
}

func TestGetHorizontalFacingMatchesTheRealAngleRanges(t *testing.T) {
	cases := []struct {
		yaw  float64
		want math.Facing
	}{
		{0, math.South},
		{44, math.South},
		{316, math.South},
		{45, math.West},
		{134, math.West},
		{135, math.North},
		{224, math.North},
		{225, math.East},
		{314, math.East},
		{-10, math.South}, // normalizes to 350
	}
	for _, c := range cases {
		p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
		p.SetRotation(c.yaw, 0)
		if got := p.GetHorizontalFacing(); got != c.want {
			t.Errorf("yaw=%v: GetHorizontalFacing() = %v, want %v", c.yaw, got, c.want)
		}
	}
}

func TestPlayerCanBeRegisteredAsARealWorldEntity(t *testing.T) {
	w := newTestWorld(t)
	p := NewPlayer(1, "Steve", "uuid-1", "xuid-1", w, math.NewVector3(5, 70, 5), GameModeSurvival)

	w.AddEntity(p)
	got, ok := w.GetEntity(1)
	if !ok {
		t.Fatal("GetEntity(1) not found after AddEntity")
	}
	if got.GetID() != 1 {
		t.Errorf("registered entity GetID() = %d, want 1", got.GetID())
	}

	w.RemoveEntity(p)
	if _, ok := w.GetEntity(1); ok {
		t.Error("GetEntity(1) still found after RemoveEntity")
	}
}

func TestGetInventoryReturnsARealUsableInventory(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	inv := p.GetInventory()
	if inv == nil {
		t.Fatal("GetInventory() = nil")
	}
	if got := inv.GetSize(); got != mainInventorySize {
		t.Errorf("GetInventory().GetSize() = %d, want %d", got, mainInventorySize)
	}
}
