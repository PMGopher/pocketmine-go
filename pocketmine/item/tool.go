package item

// Tool is a port of pocketmine\item\Tool. GetMiningEfficiency's enchantment-driven bonus
// (VanillaEnchantments::EFFICIENCY) isn't ported for the same reason ApplyDamage's unbreaking
// reduction isn't (see Durable's doc comment) - the enchantment level is always 0, so efficiency
// is just GetBaseMiningEfficiency() when isCorrectTool, matching the PHP formula with that term
// zeroed out.
type Tool struct {
	Durable
}

func (t *Tool) GetMaxStackSize() int { return 1 }

func (t *Tool) GetMiningEfficiency(isCorrectTool bool) float64 {
	if isCorrectTool {
		return t.self.(baseMiningEfficiencyShaper).GetBaseMiningEfficiency()
	}
	return 1
}

// baseMiningEfficiencyShaper lets concrete tool types override GetBaseMiningEfficiency - same
// narrow self-dispatch shape as durableShaper.
type baseMiningEfficiencyShaper interface {
	GetBaseMiningEfficiency() float64
}

func (t *Tool) GetBaseMiningEfficiency() float64 { return 1 }
