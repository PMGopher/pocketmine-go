package item

import (
	"testing"

	"pocketmine-go/pocketmine/block"
)

// Compile-time proof that a real Axe satisfies both block.Axe (wood.go) and block.Durable
// (candle_component.go) via structural typing.
var (
	_ Item          = (*Axe)(nil)
	_ block.Item    = (*Axe)(nil)
	_ block.Axe     = (*Axe)(nil)
	_ block.Durable = (*Axe)(nil)
)

func newTestAxe(tier ToolTier) *Axe {
	return NewAxe(NewItemIdentifier(IRON_AXE), "Iron Axe", tier)
}

func TestToolTierMetadata(t *testing.T) {
	if ToolTierIron.GetHarvestLevel() != 4 {
		t.Errorf("Iron GetHarvestLevel() = %d, want 4", ToolTierIron.GetHarvestLevel())
	}
	if ToolTierWood.GetMaxDurability() != 60 {
		t.Errorf("Wood GetMaxDurability() = %d, want 60", ToolTierWood.GetMaxDurability())
	}
	if ToolTierNetherite.GetEnchantability() != 15 {
		t.Errorf("Netherite GetEnchantability() = %d, want 15", ToolTierNetherite.GetEnchantability())
	}
}

func TestAxeGetMaxDurabilityFollowsTier(t *testing.T) {
	a := newTestAxe(ToolTierDiamond)
	if a.GetMaxDurability() != 1562 {
		t.Errorf("GetMaxDurability() = %d, want 1562", a.GetMaxDurability())
	}
}

func TestAxeGetAttackPointsIsTierBaseMinusOne(t *testing.T) {
	a := newTestAxe(ToolTierIron)
	if a.GetAttackPoints() != 6 {
		t.Errorf("GetAttackPoints() = %d, want 6 (iron base 7 - 1)", a.GetAttackPoints())
	}
}

func TestAxeGetBlockToolType(t *testing.T) {
	a := newTestAxe(ToolTierIron)
	if a.GetBlockToolType() != block.ToolTypeAxe {
		t.Errorf("GetBlockToolType() = %v, want ToolTypeAxe", a.GetBlockToolType())
	}
	if a.GetBlockToolHarvestLevel() != 4 {
		t.Errorf("GetBlockToolHarvestLevel() = %d, want 4", a.GetBlockToolHarvestLevel())
	}
}

func TestTieredToolWoodenFuelTime(t *testing.T) {
	wood := newTestAxe(ToolTierWood)
	if wood.GetFuelTime() != 200 {
		t.Errorf("wood GetFuelTime() = %d, want 200", wood.GetFuelTime())
	}
	iron := newTestAxe(ToolTierIron)
	if iron.GetFuelTime() != 0 {
		t.Errorf("iron GetFuelTime() = %d, want 0", iron.GetFuelTime())
	}
}

func TestTieredToolNetheriteIsFireProof(t *testing.T) {
	netherite := newTestAxe(ToolTierNetherite)
	if !netherite.IsFireProof() {
		t.Error("expected a netherite tool to be fireproof")
	}
	iron := newTestAxe(ToolTierIron)
	if iron.IsFireProof() {
		t.Error("expected a non-netherite tool not to be fireproof")
	}
}

func TestAxeApplyDamageDamages(t *testing.T) {
	a := newTestAxe(ToolTierWood)
	if !a.ApplyDamage(1) {
		t.Fatal("expected ApplyDamage to succeed")
	}
	if a.GetDamage() != 1 {
		t.Errorf("GetDamage() = %d, want 1", a.GetDamage())
	}
}
