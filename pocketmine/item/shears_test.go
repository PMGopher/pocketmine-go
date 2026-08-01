package item

import (
	"testing"

	"pocketmine-go/pocketmine/block"
)

var (
	_ Item          = (*Shears)(nil)
	_ block.Item    = (*Shears)(nil)
	_ block.Durable = (*Shears)(nil)
	_ Item          = (*Spyglass)(nil)
	_ Item          = (*Totem)(nil)
	_ Item          = (*Trident)(nil)
	_ block.Item    = (*Trident)(nil)
	_ block.Durable = (*Trident)(nil)
)

func TestShearsProperties(t *testing.T) {
	s := NewShears(NewItemIdentifier(SHEARS), "Shears")
	if s.GetMaxDurability() != 239 {
		t.Errorf("GetMaxDurability() = %d, want 239", s.GetMaxDurability())
	}
	if s.GetBlockToolType() != block.ToolTypeShears {
		t.Errorf("GetBlockToolType() = %v, want ToolTypeShears", s.GetBlockToolType())
	}
	if s.GetBlockToolHarvestLevel() != 1 {
		t.Errorf("GetBlockToolHarvestLevel() = %d, want 1", s.GetBlockToolHarvestLevel())
	}
	if eff := s.GetMiningEfficiency(true); eff != 15 {
		t.Errorf("GetMiningEfficiency(true) = %v, want 15 (Shears' base efficiency)", eff)
	}
}

func TestSpyglassMaxStackSize(t *testing.T) {
	s := NewSpyglass(NewItemIdentifier(SPYGLASS), "Spyglass")
	if s.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", s.GetMaxStackSize())
	}
}

func TestTotemMaxStackSize(t *testing.T) {
	tt := NewTotem(NewItemIdentifier(TOTEM), "Totem")
	if tt.GetMaxStackSize() != 1 {
		t.Errorf("GetMaxStackSize() = %d, want 1", tt.GetMaxStackSize())
	}
}

func TestTridentProperties(t *testing.T) {
	tr := NewTrident(NewItemIdentifier(TRIDENT), "Trident")
	if tr.GetMaxDurability() != 251 {
		t.Errorf("GetMaxDurability() = %d, want 251", tr.GetMaxDurability())
	}
	if tr.GetAttackPoints() != 9 {
		t.Errorf("GetAttackPoints() = %d, want 9", tr.GetAttackPoints())
	}
}
