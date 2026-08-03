package player

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
)

func TestGetDirectionVectorPointsSouthAtYawZeroPitchZero(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	p.SetRotation(0, 0)
	got := p.GetDirectionVector()
	want := math.NewVector3(0, 0, 1)
	if stdmathAbs(got.X-want.X) > 1e-9 || stdmathAbs(got.Y-want.Y) > 1e-9 || stdmathAbs(got.Z-want.Z) > 1e-9 {
		t.Errorf("GetDirectionVector() at yaw=0,pitch=0 = %v, want %v", got, want)
	}
}

func stdmathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func TestCanInteractIsFalseBeyondMaxDistance(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	if p.CanInteract(math.NewVector3(100, 70, 0), 7) {
		t.Error("CanInteract() at 100 blocks away with maxDistance=7 = true, want false")
	}
}

func TestAttackBlockOnAFarAwayBlockReturnsFalse(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	if p.AttackBlock(math.NewVector3(1000, 70, 1000), math.Up, fakeHandItem{}) {
		t.Error("AttackBlock() on a block 1000 blocks away = true, want false")
	}
}

func TestAttackBlockOnABreakableBlockStartsABlockBreakHandler(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	p.SetOnGround(true)
	pos := math.NewVector3(0, 70, 1)
	if err := p.GetWorld().SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, p.GetWorld()), block.VanillaStone()); err != nil {
		t.Fatal(err)
	}

	if !p.AttackBlock(pos, math.Up, fakeHandItem{}) {
		t.Fatal("AttackBlock() on a nearby breakable block = false, want true")
	}
	if p.GetBlockBreakHandler() == nil {
		t.Error("GetBlockBreakHandler() = nil after AttackBlock() on a breakable block in survival")
	}
}

func TestAttackBlockInCreativeDoesNotStartABlockBreakHandler(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	p.SetGamemode(GameModeCreative)
	pos := math.NewVector3(0, 70, 1)
	if err := p.GetWorld().SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, p.GetWorld()), block.VanillaStone()); err != nil {
		t.Fatal(err)
	}

	p.AttackBlock(pos, math.Up, fakeHandItem{})
	if p.GetBlockBreakHandler() != nil {
		t.Error("GetBlockBreakHandler() != nil after AttackBlock() in creative mode")
	}
}

func TestStopBreakBlockClearsTheHandlerForTheMatchingPosition(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	p.SetOnGround(true)
	pos := math.NewVector3(0, 70, 1)
	if err := p.GetWorld().SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, p.GetWorld()), block.VanillaStone()); err != nil {
		t.Fatal(err)
	}
	p.AttackBlock(pos, math.Up, fakeHandItem{})
	if p.GetBlockBreakHandler() == nil {
		t.Fatal("precondition failed: no block break handler started")
	}

	p.StopBreakBlock(math.NewVector3(999, 999, 999)) // different position - should not clear it
	if p.GetBlockBreakHandler() == nil {
		t.Error("StopBreakBlock cleared the handler despite targeting a different position")
	}

	p.StopBreakBlock(pos)
	if p.GetBlockBreakHandler() != nil {
		t.Error("GetBlockBreakHandler() != nil after StopBreakBlock on the matching position")
	}
}

func TestUpdateBreakingBlockClearsTheHandlerOnceBreakCompletes(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	p.SetOnGround(true)
	pos := math.NewVector3(0, 70, 1)
	if err := p.GetWorld().SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, p.GetWorld()), block.VanillaStone()); err != nil {
		t.Fatal(err)
	}
	p.AttackBlock(pos, math.Up, fakeHandItem{})

	ticks := 0
	for p.GetBlockBreakHandler() != nil {
		p.UpdateBreakingBlock(fakeHandItem{})
		ticks++
		if ticks > 1000 {
			t.Fatal("block break handler never completed - looks like an infinite loop")
		}
	}
}

func TestBreakBlockActuallyReplacesTheBlockWithAirWhenInReach(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	pos := math.NewVector3(0, 70, 1)
	if err := p.GetWorld().SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, p.GetWorld()), block.VanillaStone()); err != nil {
		t.Fatal(err)
	}

	if !p.BreakBlock(pos) {
		t.Fatal("BreakBlock() = false for a block within reach")
	}
	if got := p.GetWorld().GetBlockAt(0, 70, 1).GetTypeId(); got != block.AIR {
		t.Errorf("block type after BreakBlock() = %d, want AIR (%d)", got, block.AIR)
	}
}

func TestBreakBlockOutOfReachReturnsFalseAndLeavesTheBlockIntact(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0, 70, 0))
	pos := math.NewVector3(1000, 70, 1000)
	if err := p.GetWorld().SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, p.GetWorld()), block.VanillaStone()); err != nil {
		t.Fatal(err)
	}

	if p.BreakBlock(pos) {
		t.Fatal("BreakBlock() = true for a block far out of reach")
	}
	if got := p.GetWorld().GetBlockAt(1000, 70, 1000).GetTypeId(); got != block.STONE {
		t.Errorf("block type after a failed BreakBlock() = %d, want it to remain STONE (%d)", got, block.STONE)
	}
}
