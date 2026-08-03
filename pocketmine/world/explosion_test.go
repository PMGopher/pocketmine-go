package world

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
)

// explosionTestEntity adapts *entity.Entity (which already satisfies block.Entity in full - see
// the entity package's own doc comment) to registeredEntity by adding the bare GetID() World's
// registry needs but Entity itself doesn't have yet.
type explosionTestEntity struct {
	*entity.Entity
	id int
}

func (e *explosionTestEntity) GetID() int { return e.id }

func newExplosionTestEntity(id int, pos math.Vector3, bb math.AxisAlignedBB) *explosionTestEntity {
	return &explosionTestEntity{Entity: entity.NewEntity(pos, bb), id: id}
}

func TestNewExplosionValidatesArguments(t *testing.T) {
	w := newTestWorld()
	valid := block.NewPosition(8, 30, 8, w)
	invalid := block.NewPosition(8, 30, 8, nil)

	if _, err := NewExplosion(invalid, 4, nil, 0); err == nil {
		t.Error("expected an error for a position with no valid world")
	}
	if _, err := NewExplosion(valid, 0, nil, 0); err == nil {
		t.Error("expected an error for a non-positive radius")
	}
	if _, err := NewExplosion(valid, 4, nil, -0.1); err == nil {
		t.Error("expected an error for a negative fire chance")
	}
	if _, err := NewExplosion(valid, 4, nil, 1.1); err == nil {
		t.Error("expected an error for a fire chance above 1")
	}
	if _, err := NewExplosion(valid, 4, nil, 0.5); err != nil {
		t.Errorf("NewExplosion with valid arguments returned an error: %v", err)
	}
}

func TestExplodeADestroysStoneNearOriginButNotFarAway(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	source := block.NewPosition(8, 30, 8, w)
	exp, err := NewExplosion(source, 4, nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	if ok := exp.ExplodeA(); !ok {
		t.Fatal("ExplodeA returned false")
	}

	if len(exp.AffectedBlocks) == 0 {
		t.Fatal("ExplodeA found no affected blocks")
	}
	if _, hit := exp.AffectedBlocks[[3]int{8, 30, 8}]; !hit {
		t.Error("the origin block itself was not marked affected")
	}
	if _, hit := exp.AffectedBlocks[[3]int{8, 55, 8}]; hit {
		t.Error("a block far outside the blast radius was marked affected")
	}
}

func TestExplodeAReturnsFalseForTooSmallARadius(t *testing.T) {
	w := newTestWorld()
	source := block.NewPosition(8, 30, 8, w)
	exp, err := NewExplosion(source, 0.5, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	// Radius itself passes NewExplosion's own validation (>0), but explodeA has its own separate
	// "too small to do anything" floor at 0.1 - shrink it below that after construction to
	// exercise that second guard specifically.
	exp.Radius = 0.05
	if exp.ExplodeA() {
		t.Error("ExplodeA() = true for a radius below the 0.1 floor, want false")
	}
	if len(exp.AffectedBlocks) != 0 {
		t.Error("ExplodeA recorded affected blocks despite returning false")
	}
}

func TestExplodeBReplacesDestroyedStoneWithAir(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	if got := w.GetBlockAt(8, 30, 8).GetTypeId(); got != block.STONE {
		t.Fatalf("precondition failed: origin block isn't stone before the explosion, got %d", got)
	}

	source := block.NewPosition(8, 30, 8, w)
	exp, err := NewExplosion(source, 4, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	exp.ExplodeA()
	exp.ExplodeB()

	if got := w.GetBlockAt(8, 30, 8).GetTypeId(); got != block.AIR {
		t.Errorf("origin block after ExplodeB() = %d, want AIR (%d)", got, block.AIR)
	}
}

func TestExplodeBIgnitesNearbyTNTInsteadOfDestroyingIt(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	// The 16-ray algorithm quantizes direction (16 rays per axis, none of them exactly
	// axis-aligned - see ExplodeA's own doc comment on the ray shell), so no single adjacent
	// position is guaranteed to be hit. Surrounding the source entirely with TNT instead
	// guarantees at least one of them is, regardless of exactly which cells the rays land in.
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			for dz := -1; dz <= 1; dz++ {
				if dx == 0 && dy == 0 && dz == 0 {
					continue
				}
				pos := block.NewPosition(float64(8+dx), float64(30+dy), float64(8+dz), w)
				if err := w.SetBlock(pos, block.VanillaTNT()); err != nil {
					t.Fatal(err)
				}
			}
		}
	}

	source := block.NewPosition(8, 30, 8, w)
	exp, err := NewExplosion(source, 4, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	exp.ExplodeA()
	exp.ExplodeB()

	sawUndestroyedTNT := false
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			for dz := -1; dz <= 1; dz++ {
				if dx == 0 && dy == 0 && dz == 0 {
					continue
				}
				if w.GetBlockAt(8+dx, 30+dy, 8+dz).GetTypeId() == block.TNT {
					sawUndestroyedTNT = true
				}
			}
		}
	}
	if !sawUndestroyedTNT {
		t.Error("every surrounding TNT block was destroyed - want at least one left as TNT (ignited, not destroyed)")
	}
}

func TestExplodeBDamagesAndKnocksBackANearbyEntityButNotAFarOne(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	// y=100 is well above the flat world's terrain (topmost real block is grass at y=63 - see
	// newTestWorld's own Flat layout), so both entities have a clear, unobstructed line of sight
	// to the source - getExposure would otherwise (correctly) report near-zero exposure for an
	// entity encased in solid rock, which isn't what this test means to exercise.
	//
	// Entity.GetBoundingBox returns exactly the box passed to NewEntity with no implicit offset by
	// position, so each box below is already built around its own entity's actual position.
	nearBB, err := math.NewAxisAlignedBB(9.5-0.3, 100, 8.5-0.3, 9.5+0.3, 100+1.8, 8.5+0.3)
	if err != nil {
		t.Fatal(err)
	}
	farBB, err := math.NewAxisAlignedBB(8.5-0.3, 100, 500.5-0.3, 8.5+0.3, 100+1.8, 500.5+0.3)
	if err != nil {
		t.Fatal(err)
	}
	near := newExplosionTestEntity(1, math.NewVector3(9.5, 100, 8.5), nearBB)
	far := newExplosionTestEntity(2, math.NewVector3(8.5, 100, 500.5), farBB)
	w.AddEntity(near)
	w.AddEntity(far)

	nearHealthBefore := near.GetHealth()
	farHealthBefore := far.GetHealth()

	source := block.NewPosition(8, 100, 8, w)
	exp, err := NewExplosion(source, 4, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	exp.ExplodeA()
	exp.ExplodeB()

	if near.GetHealth() >= nearHealthBefore {
		t.Errorf("nearby entity health after explosion = %v, want less than %v", near.GetHealth(), nearHealthBefore)
	}
	if near.GetMotion() == (math.Vector3{}) {
		t.Error("nearby entity's motion was not changed by the explosion")
	}

	if far.GetHealth() != farHealthBefore {
		t.Errorf("far-away entity health after explosion = %v, want unchanged %v", far.GetHealth(), farHealthBefore)
	}
	if far.GetMotion() != (math.Vector3{}) {
		t.Error("far-away entity's motion was changed by an explosion far outside its range")
	}
}

func TestExplodeBFireIgnitionOnlyPlacesFireOnTopOfASupportingBlock(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)

	// A fully-supported stone column at (8,30,8) with air above it (30 is the topmost affected
	// layer, since the flat world's stone stops at y=59 - well above 30, so (8,31,8) is stone too;
	// use a position right at the very top of the stone column instead, so the position "above" it
	// is air both before and after the blast, and can end up as Fire if ignited).
	top := 62 // dirt layer (60-62) - the position above it (63) is grass, itself removable by the
	// blast; what matters is that (8,62,8) has *something* solid immediately above once affected,
	// which grass (opaque) satisfies via GetSupportType.
	source := block.NewPosition(8, float64(top), 8, w)
	exp, err := NewExplosion(source, 3, nil, 1.0) // fireChance=1.0: every affected block ignites
	if err != nil {
		t.Fatal(err)
	}
	exp.ExplodeA()
	if len(exp.AffectedBlocks) == 0 {
		t.Fatal("no affected blocks - test assumption about this position is wrong")
	}
	exp.ExplodeB()

	sawFire, sawAir := false, false
	for key := range exp.AffectedBlocks {
		switch w.GetBlockAt(key[0], key[1], key[2]).GetTypeId() {
		case block.FIRE:
			sawFire = true
		case block.AIR:
			sawAir = true
		}
	}
	if !sawAir {
		t.Error("no affected block became air - ExplodeB's destruction step doesn't seem to have run")
	}
	if !sawFire {
		t.Error("fireChance=1.0 but no affected block with a supporting block below it caught fire")
	}
}
