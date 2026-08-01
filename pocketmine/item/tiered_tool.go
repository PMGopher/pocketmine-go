package item

// TieredTool is a port of pocketmine\item\TieredTool.
type TieredTool struct {
	Tool

	Tier ToolTier
}

func (t *TieredTool) GetMaxDurability() int { return t.Tier.GetMaxDurability() }

func (t *TieredTool) GetTier() ToolTier { return t.Tier }

func (t *TieredTool) GetBaseMiningEfficiency() float64 { return float64(t.Tier.GetBaseEfficiency()) }

func (t *TieredTool) GetEnchantability() int { return t.Tier.GetEnchantability() }

// GetFuelTime is a port of TieredTool::getFuelTime.
func (t *TieredTool) GetFuelTime() int {
	if t.Tier == ToolTierWood {
		return 200
	}
	return 0
}

func (t *TieredTool) IsFireProof() bool { return t.Tier == ToolTierNetherite }
