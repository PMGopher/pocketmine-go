package block

import (
	"testing"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
)

func newTestCampfire(w World) *Campfire {
	c := NewCampfire(mustBlockIdentifier(1080), "Test Campfire", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	c.SetPosition(w, 1, 2, 3)
	return c
}

func newTestSoulCampfire(w World) *SoulCampfire {
	s := NewSoulCampfire(mustBlockIdentifier(1081), "Test Soul Campfire", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))
	s.SetPosition(w, 1, 2, 3)
	return s
}

func TestCampfireGetLightLevel(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCampfire(w)
	if c.GetLightLevel() != 0 {
		t.Errorf("GetLightLevel() = %d, want 0 while unlit", c.GetLightLevel())
	}
	c.Lit = true
	if c.GetLightLevel() != 15 {
		t.Errorf("GetLightLevel() = %d, want 15 while lit", c.GetLightLevel())
	}
}

func TestSoulCampfireGetLightLevel(t *testing.T) {
	w := &fakeWorld{}
	s := newTestSoulCampfire(w)
	if s.GetLightLevel() != 0 {
		t.Errorf("GetLightLevel() = %d, want 0 while unlit", s.GetLightLevel())
	}
	s.Lit = true
	if s.GetLightLevel() != 10 {
		t.Errorf("GetLightLevel() = %d, want 10 while lit", s.GetLightLevel())
	}
}

func TestCampfireEntityCollisionDamageDiffersFromSoulCampfire(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCampfire(w)
	s := newTestSoulCampfire(w)

	if c.GetEntityCollisionDamage() != 1 {
		t.Errorf("Campfire.GetEntityCollisionDamage() = %d, want 1", c.GetEntityCollisionDamage())
	}
	if s.GetEntityCollisionDamage() != 2 {
		t.Errorf("SoulCampfire.GetEntityCollisionDamage() = %d, want 2", s.GetEntityCollisionDamage())
	}
}

func TestCampfirePlaceSetsFacingDirectlyAndLightsIt(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCampfire(w)
	tx := &fakeBlockTransaction{}
	player := &fakeSignPlayer{}

	if !c.Place(tx, fakeItem{}, c, c, math.Up, math.Vector3{}, player) {
		t.Fatal("expected Place to succeed")
	}
	if c.Facing != player.GetHorizontalFacing() {
		t.Errorf("Facing = %v, want the player's own facing (%v), not its opposite", c.Facing, player.GetHorizontalFacing())
	}
	if !c.Lit {
		t.Error("expected a newly placed campfire to be lit")
	}
}

func TestCampfirePlaceFailsWhenBelowIsAnotherCampfire(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	below := newTestCampfire(w)
	below.SetPosition(w, 1, 1, 3)
	w.blocks[[3]int{1, 1, 3}] = below

	c := newTestCampfire(w)
	tx := &fakeBlockTransaction{}

	if c.Place(tx, fakeItem{}, c, c, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to fail when the block below is a campfire")
	}
}

// TestCampfirePlaceFailsWhenBelowIsASoulCampfire exercises the campfireMarker approach: a plain
// *Campfire's Place must still reject placement above a *SoulCampfire (and vice versa), matching
// PHP's `instanceof Campfire` (SoulCampfire IS-A Campfire), which a naive Go type assertion to
// *Campfire alone would have missed.
func TestCampfirePlaceFailsWhenBelowIsASoulCampfire(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	below := newTestSoulCampfire(w)
	below.SetPosition(w, 1, 1, 3)
	w.blocks[[3]int{1, 1, 3}] = below

	c := newTestCampfire(w)
	tx := &fakeBlockTransaction{}

	if c.Place(tx, fakeItem{}, c, c, math.Up, math.Vector3{}, nil) {
		t.Error("expected Place to fail when the block below is a soul campfire")
	}
}

func TestCampfireOnInteractFireChargeIgnitesAndPopsItem(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCampfire(w)
	item := fakeItem{typeID: itemTypeIDsFireCharge}

	if !c.OnInteract(item, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to handle a fire charge")
	}
	if !c.Lit {
		t.Error("expected the campfire to be lit")
	}
}

func TestCampfireOnInteractFlintAndSteelIgnitesAndDamages(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCampfire(w)
	axe := &fakeAxeItem{fakeItem: fakeItem{typeID: itemTypeIDsFlintAndSteel}}

	if !c.OnInteract(axe, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to handle flint and steel")
	}
	if !c.Lit {
		t.Error("expected the campfire to be lit")
	}
	if axe.damage != 1 {
		t.Errorf("durable damage = %d, want 1", axe.damage)
	}
}

func TestCampfireOnInteractShovelExtinguishesWhenLit(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCampfire(w)
	c.Lit = true
	shovel := &fakeAxeItem{}

	if !c.OnInteract(shovel, math.Up, math.Vector3{}, nil, nil) {
		t.Fatal("expected OnInteract to handle a shovel while lit")
	}
	if c.Lit {
		t.Error("expected the campfire to be extinguished")
	}
	if shovel.damage != 1 {
		t.Errorf("durable damage = %d, want 1", shovel.damage)
	}
}

func TestCampfireOnInteractIrrelevantItemDoesNothing(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCampfire(w)

	if c.OnInteract(fakeItem{}, math.Up, math.Vector3{}, nil, nil) {
		t.Error("expected an unrelated item not to be handled")
	}
}

// fakeTypedBlock is a minimal Behavior whose only interesting property is its type ID -
// OnNearbyBlockChange only cares about GetTypeId() == WATER.
type fakeTypedBlock struct{ Block }

func newFakeTypedBlock(typeID int) *fakeTypedBlock {
	idInfo, err := NewBlockIdentifier(typeID, nil)
	if err != nil {
		panic(err)
	}
	b := &fakeTypedBlock{Block: NewBlock(idInfo, "Fake Typed Block", NewBlockTypeInfo(BlockBreakInfoInstant(ToolTypeNone, 0), nil, nil))}
	b.Init(b)
	return b
}

func (b *fakeTypedBlock) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func TestCampfireOnNearbyBlockChangeExtinguishesUnderWater(t *testing.T) {
	w := &containerTileWorld{tiles: map[[3]int]Tile{}, blocks: map[[3]int]Behavior{}}
	c := newTestCampfire(w)
	c.Lit = true

	water := newFakeTypedBlock(WATER)
	water.SetPosition(w, 1, 3, 3)
	w.blocks[[3]int{1, 3, 3}] = water

	c.OnNearbyBlockChange()

	if c.Lit {
		t.Error("expected the campfire to be extinguished when water is placed above it")
	}
}

func TestCampfireOnEntityInsideIgnitesWhenEntityOnFire(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCampfire(w)
	fireEntity := &onFireTrackingEntity{reportsOnFire: true}

	handled := c.OnEntityInside(fireEntity)

	if handled {
		t.Error("expected OnEntityInside to return false when it ignites")
	}
	if !c.Lit {
		t.Error("expected the campfire to be lit")
	}
}

func TestCampfireOnEntityInsideDamagesLivingWhileLit(t *testing.T) {
	w := &fakeWorld{}
	c := newTestCampfire(w)
	c.Lit = true

	living := entity.NewLiving(math.NewVector3(0, 0, 0), math.OneAABB())
	startHealth := living.GetHealth()

	if !c.OnEntityInside(living) {
		t.Fatal("expected OnEntityInside to return true")
	}
	if living.GetHealth() != startHealth-1 { // Campfire.GetEntityCollisionDamage() == 1
		t.Errorf("GetHealth() = %v, want %v", living.GetHealth(), startHealth-1)
	}
}

// TestSoulCampfireOnEntityInsideDamagesLivingWithItsOwnAmount confirms the self-dispatch through
// campfireEntityDamageShaper reaches SoulCampfire's override (2 damage), not the inherited
// Campfire.GetEntityCollisionDamage (1) that OnEntityInside is defined on.
func TestSoulCampfireOnEntityInsideDamagesLivingWithItsOwnAmount(t *testing.T) {
	w := &fakeWorld{}
	s := newTestSoulCampfire(w)
	s.Lit = true

	living := entity.NewLiving(math.NewVector3(0, 0, 0), math.OneAABB())
	startHealth := living.GetHealth()

	s.OnEntityInside(living)

	if living.GetHealth() != startHealth-2 {
		t.Errorf("GetHealth() = %v, want %v (SoulCampfire.GetEntityCollisionDamage() == 2)", living.GetHealth(), startHealth-2)
	}
}
