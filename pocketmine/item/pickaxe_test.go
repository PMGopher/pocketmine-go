package item

import (
	"testing"

	"pocketmine-go/pocketmine/block"
)

var (
	_ Item          = (*Pickaxe)(nil)
	_ block.Item    = (*Pickaxe)(nil)
	_ block.Durable = (*Pickaxe)(nil)
	_ Item          = (*Shovel)(nil)
	_ block.Item    = (*Shovel)(nil)
	_ block.Durable = (*Shovel)(nil)
	_ Item          = (*Sword)(nil)
	_ block.Item    = (*Sword)(nil)
	_ block.Durable = (*Sword)(nil)
	_ Item          = (*Hoe)(nil)
	_ block.Item    = (*Hoe)(nil)
	_ block.Durable = (*Hoe)(nil)
)

func TestPickaxeGetAttackPointsAndBlockToolType(t *testing.T) {
	p := NewPickaxe(NewItemIdentifier(IRON_PICKAXE), "Iron Pickaxe", ToolTierIron)
	if p.GetAttackPoints() != 5 {
		t.Errorf("GetAttackPoints() = %d, want 5 (iron base 7 - 2)", p.GetAttackPoints())
	}
	if p.GetBlockToolType() != block.ToolTypePickaxe {
		t.Errorf("GetBlockToolType() = %v, want ToolTypePickaxe", p.GetBlockToolType())
	}
}

func TestShovelGetAttackPointsAndBlockToolType(t *testing.T) {
	s := NewShovel(NewItemIdentifier(IRON_SHOVEL), "Iron Shovel", ToolTierIron)
	if s.GetAttackPoints() != 4 {
		t.Errorf("GetAttackPoints() = %d, want 4 (iron base 7 - 3)", s.GetAttackPoints())
	}
	if s.GetBlockToolType() != block.ToolTypeShovel {
		t.Errorf("GetBlockToolType() = %v, want ToolTypeShovel", s.GetBlockToolType())
	}
}

func TestSwordGetAttackPointsAndHarvestLevel(t *testing.T) {
	sw := NewSword(NewItemIdentifier(IRON_SWORD), "Iron Sword", ToolTierIron)
	if sw.GetAttackPoints() != 7 {
		t.Errorf("GetAttackPoints() = %d, want 7 (full tier base)", sw.GetAttackPoints())
	}
	if sw.GetBlockToolHarvestLevel() != 1 {
		t.Errorf("GetBlockToolHarvestLevel() = %d, want 1 (fixed for swords)", sw.GetBlockToolHarvestLevel())
	}
}

func TestSwordMiningEfficiencyScalesByOneAndHalf(t *testing.T) {
	sw := NewSword(NewItemIdentifier(IRON_SWORD), "Iron Sword", ToolTierIron)
	if eff := sw.GetMiningEfficiency(true); eff != 15 {
		t.Errorf("GetMiningEfficiency(true) = %v, want 15 (base 10 * 1.5)", eff)
	}
	if eff := sw.GetMiningEfficiency(false); eff != 1.5 {
		t.Errorf("GetMiningEfficiency(false) = %v, want 1.5 (1 * 1.5)", eff)
	}
}

func TestHoeGetBlockToolType(t *testing.T) {
	h := NewHoe(NewItemIdentifier(IRON_HOE), "Iron Hoe", ToolTierIron)
	if h.GetBlockToolType() != block.ToolTypeHoe {
		t.Errorf("GetBlockToolType() = %v, want ToolTypeHoe", h.GetBlockToolType())
	}
	if h.GetAttackPoints() != 1 {
		t.Errorf("GetAttackPoints() = %d, want 1 (Hoe doesn't override the Item default)", h.GetAttackPoints())
	}
}
