package entity_test

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/entity/effect"
	entityevent "pocketmine-go/pocketmine/event/entity"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/world"
	"pocketmine-go/pocketmine/world/generator"
)

func newTestWorld(t *testing.T) *world.World {
	t.Helper()
	gen := generator.NewFlat(0, generator.VanillaFlatLayers(), generator.VanillaFlatBiomeID, int32(block.VanillaAir().GetStateId()), nil)
	return world.New(gen, convert.NewBlockTranslator(), []block.Behavior{
		block.VanillaAir(), block.VanillaBedrock(), block.VanillaStone(), block.VanillaDirt(), block.VanillaGrass(),
	})
}

func newTestSkin(t *testing.T) *entity.Skin {
	t.Helper()
	skin, err := entity.NewSkin("Standard_Custom", make([]byte, 64*64*4), nil, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	return skin
}

func newTestHuman(t *testing.T, w *world.World) *entity.Human {
	t.Helper()
	h := entity.NewHuman(entity.NewLocation(0.5, 100, 0.5, w, 0, 0), newTestSkin(t), nil)
	h.NoDamageTicks = 0
	return h
}

func attack(h *entity.Human, cause int, damage float64) entityevent.DamageSource {
	ev := entityevent.NewEntityDamageEvent(h, cause, damage, nil)
	h.Attack(ev)
	return ev
}

func TestNewHumanDefaults(t *testing.T) {
	w := newTestWorld(t)
	h := newTestHuman(t, w)

	if h.GetHealth() != 20 || h.GetMaxHealth() != 20 {
		t.Errorf("health = %v/%d, want 20/20", h.GetHealth(), h.GetMaxHealth())
	}
	if got := h.GetHungerManager().GetFood(); got != 20 {
		t.Errorf("food = %v, want 20", got)
	}
	if got := h.GetHungerManager().GetSaturation(); got != 20 {
		t.Errorf("saturation = %v, want 20 (AttributeFactory default)", got)
	}
	if h.GetInventory().GetSize() != 36 {
		t.Errorf("inventory size = %d, want 36", h.GetInventory().GetSize())
	}
	if _, ok := w.GetEntity(h.GetID()); !ok {
		t.Error("human is not registered in its world after construction")
	}
	if h.GetEyeHeight() != 1.62 {
		t.Errorf("GetEyeHeight() = %v, want 1.62", h.GetEyeHeight())
	}
}

func TestAttackAppliesDamageAndCooldown(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))

	attack(h, entityevent.CauseMagic, 4)
	if h.GetHealth() != 16 {
		t.Fatalf("health = %v after 4 damage, want 16", h.GetHealth())
	}

	// Living::applyDamageModifiers: during the attack cooldown a weaker hit is cancelled and a
	// stronger one only deals the difference.
	if ev := attack(h, entityevent.CauseMagic, 3); !ev.IsCancelled() {
		t.Error("a weaker hit during the attack cooldown was not cancelled")
	}
	attack(h, entityevent.CauseMagic, 6)
	if h.GetHealth() != 14 {
		t.Errorf("health = %v after a 6 damage hit during cooldown, want 14 (only the 2 extra)", h.GetHealth())
	}
}

func TestArmorReducesDamage(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))
	chestplate := item.NewArmor(item.NewItemIdentifier(item.DIAMOND_CHESTPLATE), "Diamond Chestplate",
		item.NewArmorTypeInfo(8, 529, 1, item.NewArmorMaterial(10, nil)))
	h.GetArmorInventory().SetChestplate(chestplate)

	if got := h.GetArmorPoints(); got != 8 {
		t.Fatalf("GetArmorPoints() = %d, want 8", got)
	}
	ev := attack(h, entityevent.CauseEntityAttack, 10)
	if got := ev.GetModifier(entityevent.ModifierArmor); got != -3.2 {
		t.Errorf("armor modifier = %v, want -3.2 (10 * 8 * 0.04)", got)
	}
	// Fall damage bypasses armor.
	h2 := newTestHuman(t, newTestWorld(t))
	h2.GetArmorInventory().SetChestplate(chestplate.Clone())
	if ev := attack(h2, entityevent.CauseFall, 10); ev.IsApplicable(entityevent.ModifierArmor) && ev.GetModifier(entityevent.ModifierArmor) != 0 {
		t.Errorf("fall damage was reduced by armor: %v", ev.GetModifier(entityevent.ModifierArmor))
	}
}

func TestResistanceAndFireResistance(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))
	h.GetEffects().Add(effect.NewEffectInstanceWith(effect.VanillaResistance(), 600, 1))
	ev := attack(h, entityevent.CauseMagic, 10)
	if got := ev.GetFinalDamage(); got != 6 {
		t.Errorf("final damage with Resistance II = %v, want 6", got)
	}

	h2 := newTestHuman(t, newTestWorld(t))
	h2.GetEffects().Add(effect.NewEffectInstanceWith(effect.VanillaFireResistance(), 600, 0))
	if ev := attack(h2, entityevent.CauseLava, 4); !ev.IsCancelled() {
		t.Error("lava damage was not cancelled by Fire Resistance")
	}
}

func TestHealthBoostRaisesMaxHealth(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))
	boost := effect.VanillaHealthBoost()
	h.GetEffects().Add(effect.NewEffectInstanceWith(boost, 600, 1))
	if got := h.GetMaxHealth(); got != 28 {
		t.Errorf("max health with Health Boost II = %d, want 28", got)
	}
	h.GetEffects().Remove(boost)
	if got := h.GetMaxHealth(); got != 20 {
		t.Errorf("max health after removing Health Boost = %d, want 20", got)
	}
}

func TestRegenerationHealsOverTime(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))
	h.SetHealth(10)
	h.GetHungerManager().SetEnabled(false) // food regen (HungerManager::tick) would heal too
	h.GetEffects().Add(effect.NewEffectInstanceWith(effect.VanillaRegeneration(), 200, 0))
	h.SetHasGravity(false)
	for tick := range int64(200) {
		h.OnUpdate(tick + 1)
	}
	// Regeneration I heals 1 every 50 ticks.
	if got := h.GetHealth(); got != 14 {
		t.Errorf("health after 200 ticks of Regeneration I = %v, want 14", got)
	}
	if h.GetEffects().Has(effect.VanillaRegeneration()) {
		t.Error("the effect is still active after its duration ran out")
	}
}

func TestPoisonNeverKills(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))
	h.SetHealth(3)
	h.GetEffects().Add(effect.NewEffectInstanceWith(effect.VanillaPoison(), 1000, 3))
	h.SetHasGravity(false)
	// Tick the whole entity (not just its effects), so the attack cooldown counts down like it
	// does in PHP's Living::entityBaseTick.
	for tick := range int64(1000) {
		h.OnUpdate(tick + 1)
	}
	if got := h.GetHealth(); got != 1 {
		t.Errorf("health after a long Poison IV = %v, want 1 (poison stops at 1)", got)
	}
}

func TestHungerExhaustion(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))
	hunger := h.GetHungerManager()

	// Every 4 exhaustion removes 1 saturation, then 1 food once saturation is gone.
	hunger.Exhaust(4*20, playerevent.ExhaustCauseCustom)
	if hunger.GetSaturation() != 0 || hunger.GetFood() != 20 {
		t.Fatalf("after 80 exhaustion: saturation=%v food=%v, want 0/20", hunger.GetSaturation(), hunger.GetFood())
	}
	hunger.Exhaust(4*3, playerevent.ExhaustCauseCustom)
	if hunger.GetFood() != 17 {
		t.Errorf("food = %v, want 17", hunger.GetFood())
	}
	if !hunger.IsHungry() {
		t.Error("IsHungry() = false with food below max")
	}
}

func TestFoodRegeneration(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))
	h.SetHasGravity(false)
	h.SetHealth(10)
	for tick := range int64(160) {
		h.OnUpdate(tick + 1)
	}
	// HungerManager::tick heals 1 every 80 ticks while food >= 18, costing 6 exhaustion each.
	if got := h.GetHealth(); got != 12 {
		t.Errorf("health after 160 ticks with full food = %v, want 12", got)
	}
	if got := h.GetHungerManager().GetExhaustion(); got != 0 || h.GetHungerManager().GetSaturation() != 17 {
		t.Errorf("exhaustion=%v saturation=%v, want 0/17 after 12 exhaustion", got, h.GetHungerManager().GetSaturation())
	}
}

func TestExperienceManager(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))
	xp := h.GetXpManager()

	xp.AddXp(7, false)
	if xp.GetXpLevel() != 1 || xp.GetXpProgress() != 0 {
		t.Errorf("after 7 XP: level=%d progress=%v, want 1/0", xp.GetXpLevel(), xp.GetXpProgress())
	}
	xp.AddXp(352-7, false)
	if xp.GetXpLevel() != 16 {
		t.Errorf("after 352 XP: level=%d, want 16", xp.GetXpLevel())
	}
	if xp.GetCurrentTotalXp() != 352 || xp.GetLifetimeTotalXp() != 352 {
		t.Errorf("total XP = %d (lifetime %d), want 352", xp.GetCurrentTotalXp(), xp.GetLifetimeTotalXp())
	}
	// Human::getXpDropAmount: min(100, 7 * level) for non-spectators.
	if got := h.GetXpDropAmount(); got != 100 {
		t.Errorf("GetXpDropAmount() = %d, want 100", got)
	}
}

func TestFallDamage(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))
	if got := h.CalculateFallDamage(3); got != 0 {
		t.Errorf("CalculateFallDamage(3) = %v, want 0", got)
	}
	if got := h.CalculateFallDamage(10); got != 7 {
		t.Errorf("CalculateFallDamage(10) = %v, want 7", got)
	}
	h.GetEffects().Add(effect.NewEffectInstanceWith(effect.VanillaJumpBoost(), 600, 1))
	if got := h.CalculateFallDamage(10); got != 5 {
		t.Errorf("CalculateFallDamage(10) with Jump Boost II = %v, want 5", got)
	}
}

func TestDeath(t *testing.T) {
	h := newTestHuman(t, newTestWorld(t))
	h.GetInventory().SetItem(0, item.VanillaApple())
	attack(h, entityevent.CauseVoid, 100)
	if h.IsAlive() {
		t.Fatal("human is alive after 100 damage")
	}
	drops := h.GetDrops()
	if len(drops) != 1 || !drops[0].Equals(item.VanillaApple(), true) {
		t.Errorf("GetDrops() = %v, want the apple from the inventory", drops)
	}
}

func TestHumanNBTRoundTrip(t *testing.T) {
	w := newTestWorld(t)
	h := newTestHuman(t, w)
	h.SetHealth(13)
	h.GetHungerManager().SetFood(9)
	h.GetXpManager().SetXpLevel(5)
	h.SetNameTag("Alex")
	tag := h.SaveNBT()

	loaded := entity.NewHuman(entity.NewLocation(0.5, 100, 0.5, w, 0, 0), newTestSkin(t), tag)
	if loaded.GetHealth() != 13 {
		t.Errorf("health = %v after reload, want 13", loaded.GetHealth())
	}
	if loaded.GetHungerManager().GetFood() != 9 {
		t.Errorf("food = %v after reload, want 9", loaded.GetHungerManager().GetFood())
	}
	if loaded.GetXpManager().GetXpLevel() != 5 {
		t.Errorf("XP level = %d after reload, want 5", loaded.GetXpManager().GetXpLevel())
	}
	if loaded.GetNameTag() != "Alex" {
		t.Errorf("name tag = %q after reload, want Alex", loaded.GetNameTag())
	}
}

func TestEntityFactoryRecreatesSavedEntities(t *testing.T) {
	w := newTestWorld(t)
	z := entity.NewZombie(entity.NewLocation(3, 100, 4, w, 90, 0), nil)
	z.SetHealth(7)
	tag := z.SaveNBT()

	created, err := entity.GetEntityFactory().CreateFromData(w, tag)
	if err != nil {
		t.Fatal(err)
	}
	loaded, ok := created.(*entity.Zombie)
	if !ok {
		t.Fatalf("CreateFromData returned %T, want *entity.Zombie", created)
	}
	if loaded.GetHealth() != 7 {
		t.Errorf("health = %v, want 7", loaded.GetHealth())
	}
	if pos := loaded.GetPosition(); pos.X != 3 || pos.Z != 4 {
		t.Errorf("position = %v, want x=3 z=4", pos)
	}
}
