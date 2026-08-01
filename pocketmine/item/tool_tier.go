package item

// ToolTier is a port of pocketmine\item\ToolTier.
type ToolTier int

const (
	ToolTierWood ToolTier = iota
	ToolTierGold
	ToolTierStone
	ToolTierCopper
	ToolTierIron
	ToolTierDiamond
	ToolTierNetherite
)

// toolTierMetadata mirrors ToolTier::meta()'s 5-tuple.
type toolTierMetadata struct {
	harvestLevel     int
	maxDurability    int
	baseAttackPoints int
	baseEfficiency   int
	enchantability   int
}

var toolTierMetadataTable = map[ToolTier]toolTierMetadata{
	ToolTierWood:      {harvestLevel: 1, maxDurability: 60, baseAttackPoints: 5, baseEfficiency: 2, enchantability: 15},
	ToolTierGold:      {harvestLevel: 2, maxDurability: 33, baseAttackPoints: 5, baseEfficiency: 12, enchantability: 22},
	ToolTierStone:     {harvestLevel: 3, maxDurability: 132, baseAttackPoints: 6, baseEfficiency: 4, enchantability: 5},
	ToolTierCopper:    {harvestLevel: 3, maxDurability: 191, baseAttackPoints: 6, baseEfficiency: 5, enchantability: 13},
	ToolTierIron:      {harvestLevel: 4, maxDurability: 251, baseAttackPoints: 7, baseEfficiency: 6, enchantability: 14},
	ToolTierDiamond:   {harvestLevel: 5, maxDurability: 1562, baseAttackPoints: 8, baseEfficiency: 8, enchantability: 10},
	ToolTierNetherite: {harvestLevel: 6, maxDurability: 2032, baseAttackPoints: 9, baseEfficiency: 9, enchantability: 15},
}

func (t ToolTier) GetHarvestLevel() int { return toolTierMetadataTable[t].harvestLevel }

func (t ToolTier) GetMaxDurability() int { return toolTierMetadataTable[t].maxDurability }

func (t ToolTier) GetBaseAttackPoints() int { return toolTierMetadataTable[t].baseAttackPoints }

func (t ToolTier) GetBaseEfficiency() int { return toolTierMetadataTable[t].baseEfficiency }

func (t ToolTier) GetEnchantability() int { return toolTierMetadataTable[t].enchantability }
