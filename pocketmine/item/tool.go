package item

import "pocketmine-go/pocketmine/item/enchantment"

// Tool is a port of pocketmine\item\Tool.
type Tool struct {
	Durable
}

func (t *Tool) GetMaxStackSize() int { return 1 }

// GetMiningEfficiency is a port of Tool::getMiningEfficiency.
func (t *Tool) GetMiningEfficiency(isCorrectTool bool) float64 {
	efficiency := 1.0
	if isCorrectTool {
		efficiency = t.self.(baseMiningEfficiencyShaper).GetBaseMiningEfficiency()
		if enchantmentLevel := t.self.GetEnchantmentLevel(enchantment.VanillaEfficiency()); enchantmentLevel > 0 {
			efficiency += float64(enchantmentLevel*enchantmentLevel + 1)
		}
	}
	return efficiency
}

// baseMiningEfficiencyShaper lets concrete tool types override GetBaseMiningEfficiency - same
// narrow self-dispatch shape as durableShaper.
type baseMiningEfficiencyShaper interface {
	GetBaseMiningEfficiency() float64
}

func (t *Tool) GetBaseMiningEfficiency() float64 { return 1 }
