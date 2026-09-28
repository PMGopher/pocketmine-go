package player

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	_ "pocketmine-go/pocketmine/entity/object"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
)

// setBlockAt puts blk at (x, y, z) in p's world.
func setBlockAt(t *testing.T, p *Player, x, y, z int, blk block.Behavior) {
	t.Helper()
	if err := p.GetWorld().SetBlock(block.NewPosition(float64(x), float64(y), float64(z), p.GetWorld()), blk); err != nil {
		t.Fatal(err)
	}
}

func TestSeedItemsPlaceTheirCrops(t *testing.T) {
	for name, want := range map[string]string{"wheat_seeds": "wheat", "carrot": "carrots", "potato": "potatoes", "redstone_dust": "redstone_wire", "string": "tripwire", "melon_seeds": "melon_stem"} {
		if got := item.VanillaItem(name).GetBlock().GetTypeId(); got != block.VanillaBlock(want).GetTypeId() {
			t.Errorf("%s places type %d, want %s", name, got, want)
		}
	}
}

func TestBucketFillAndEmpty(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	setBlockAt(t, p, 0, 69, 2, block.VanillaStone())
	setBlockAt(t, p, 0, 70, 2, block.VanillaBlock("water"))

	p.GetInventory().SetItemInHand(item.VanillaBucket())
	if !p.InteractBlock(math.NewVector3(0, 70, 2), math.Up, math.NewVector3(0.5, 1, 0.5)) {
		t.Fatal("using a bucket on water failed")
	}
	if got := p.GetWorld().GetBlockAt(0, 70, 2).GetTypeId(); got != block.AIR {
		t.Errorf("water source still there after filling (type %d)", got)
	}
	held := p.GetInventory().GetItemInHand()
	if held.GetTypeId() != item.VanillaItem("water_bucket").GetTypeId() {
		t.Fatalf("held after filling = %v, want a water bucket", held)
	}

	// Emptying it on top of the stone puts flowing water back.
	if !p.InteractBlock(math.NewVector3(0, 69, 2), math.Up, math.NewVector3(0.5, 1, 0.5)) {
		t.Fatal("emptying the water bucket failed")
	}
	if got := p.GetWorld().GetBlockAt(0, 70, 2).GetTypeId(); got != block.VanillaBlock("water").GetTypeId() {
		t.Errorf("block after emptying = type %d, want water", got)
	}
	if held := p.GetInventory().GetItemInHand(); held.GetTypeId() != item.VanillaBucket().GetTypeId() || p.IsCreative() {
		t.Errorf("held after emptying = %v, want an empty bucket", held)
	}
}

func TestFlintAndSteelLightsFire(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	setBlockAt(t, p, 0, 69, 2, block.VanillaStone())
	setBlockAt(t, p, 0, 70, 2, block.VanillaAir())

	p.GetInventory().SetItemInHand(item.VanillaItem("flint_and_steel"))
	if !p.InteractBlock(math.NewVector3(0, 69, 2), math.Up, math.NewVector3(0.5, 1, 0.5)) {
		t.Fatal("using flint and steel failed")
	}
	if got := p.GetWorld().GetBlockAt(0, 70, 2).GetTypeId(); got != block.VanillaBlock("fire").GetTypeId() {
		t.Errorf("block above = type %d, want fire", got)
	}
	if d := p.GetInventory().GetItemInHand().(*item.FlintSteel).GetDamage(); d != 1 {
		t.Errorf("flint and steel damage = %d, want 1", d)
	}
}

func TestSpawnEggSpawnsItsMob(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	setBlockAt(t, p, 0, 69, 2, block.VanillaStone())
	setBlockAt(t, p, 0, 70, 2, block.VanillaAir())

	egg := item.VanillaZombieSpawnEgg()
	egg.SetCustomName("Bob")
	p.GetInventory().SetItemInHand(egg)
	if !p.InteractBlock(math.NewVector3(0, 69, 2), math.Up, math.NewVector3(0.5, 1, 0.5)) {
		t.Fatal("using the spawn egg failed")
	}
	var zombie *entity.Zombie
	for _, e := range p.GetWorld().GetEntities() {
		if z, ok := e.(*entity.Zombie); ok {
			zombie = z
		}
	}
	if zombie == nil {
		t.Fatal("no zombie was spawned")
	}
	if pos := zombie.GetPosition(); pos.X != 0.5 || pos.Y != 70 || pos.Z != 2.5 {
		t.Errorf("zombie at %v, want (0.5, 70, 2.5)", pos)
	}
	if zombie.GetNameTag() != "Bob" {
		t.Errorf("name tag = %q, want the egg's custom name", zombie.GetNameTag())
	}
	if !p.GetInventory().GetItemInHand().IsNull() {
		t.Error("the egg wasn't used up")
	}
}

func TestGlassBottleFillsWithWater(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	setBlockAt(t, p, 0, 69, 2, block.VanillaStone())
	setBlockAt(t, p, 0, 70, 2, block.VanillaBlock("water"))

	p.GetInventory().SetItemInHand(item.VanillaGlassBottle())
	p.InteractBlock(math.NewVector3(0, 70, 2), math.Up, math.NewVector3(0.5, 1, 0.5))
	potion, ok := p.GetInventory().GetItemInHand().(*item.Potion)
	if !ok || potion.PotionTypeValue != item.PotionTypeWater {
		t.Errorf("held = %v, want a water bottle", p.GetInventory().GetItemInHand())
	}
}
