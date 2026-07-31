package block

import "fmt"

const (
	// CompatibleToolMultiplier: if the tool is the correct type and high enough harvest level
	// (tool tier), base break time is hardness multiplied by this value.
	CompatibleToolMultiplier = 1.5
	// IncompatibleToolMultiplier: if the tool is an incorrect type or too low harvest level, base
	// break time is hardness multiplied by this value instead.
	IncompatibleToolMultiplier = 5.0

	// DefaultIndestructibleBlastResistance mirrors indestructible()'s PHP default parameter value
	// (Go has no default parameters, so callers pass this explicitly).
	DefaultIndestructibleBlastResistance = 18000003.75
)

// BlockBreakInfo is a port of pocketmine\block\BlockBreakInfo.
type BlockBreakInfo struct {
	hardness             float64
	toolType             ToolType
	toolHarvestLevel     int
	blastResistance      float64
	explosionHarvestable bool
}

// NewBlockBreakInfo mirrors the constructor. Pass nil for blastResistance/explosionHarvestable
// to use their PHP defaults (5x hardness, and "true if a pickaxe is compatible" respectively).
func NewBlockBreakInfo(hardness float64, toolType ToolType, toolHarvestLevel int, blastResistance *float64, explosionHarvestable *bool) *BlockBreakInfo {
	br := hardness * 5
	if blastResistance != nil {
		br = *blastResistance
	}
	eh := toolType&ToolTypePickaxe != 0
	if explosionHarvestable != nil {
		eh = *explosionHarvestable
	}
	return &BlockBreakInfo{hardness: hardness, toolType: toolType, toolHarvestLevel: toolHarvestLevel, blastResistance: br, explosionHarvestable: eh}
}

func BlockBreakInfoTier(hardness float64, toolType ToolType, toolTier ToolTier, blastResistance *float64) *BlockBreakInfo {
	return NewBlockBreakInfo(hardness, toolType, toolTier.GetHarvestLevel(), blastResistance, nil)
}

func harvestLevelOf(toolTier ToolTier) int {
	if toolTier == nil {
		return 0
	}
	return toolTier.GetHarvestLevel()
}

func BlockBreakInfoPickaxe(hardness float64, toolTier ToolTier, blastResistance *float64) *BlockBreakInfo {
	return NewBlockBreakInfo(hardness, ToolTypePickaxe, harvestLevelOf(toolTier), blastResistance, nil)
}

func BlockBreakInfoShovel(hardness float64, toolTier ToolTier, blastResistance *float64) *BlockBreakInfo {
	return NewBlockBreakInfo(hardness, ToolTypeShovel, harvestLevelOf(toolTier), blastResistance, nil)
}

func BlockBreakInfoAxe(hardness float64, toolTier ToolTier, blastResistance *float64) *BlockBreakInfo {
	return NewBlockBreakInfo(hardness, ToolTypeAxe, harvestLevelOf(toolTier), blastResistance, nil)
}

func BlockBreakInfoInstant(toolType ToolType, toolHarvestLevel int) *BlockBreakInfo {
	zero := 0.0
	return NewBlockBreakInfo(0.0, toolType, toolHarvestLevel, &zero, nil)
}

func BlockBreakInfoIndestructible(blastResistance float64) *BlockBreakInfo {
	return NewBlockBreakInfo(-1.0, ToolTypeNone, 0, &blastResistance, nil)
}

func (b *BlockBreakInfo) GetHardness() float64 { return b.hardness }

func (b *BlockBreakInfo) IsBreakable() bool { return b.hardness >= 0 }

func (b *BlockBreakInfo) BreaksInstantly() bool { return b.hardness == 0.0 }

func (b *BlockBreakInfo) GetBlastResistance() float64 { return b.blastResistance }

func (b *BlockBreakInfo) GetToolType() ToolType { return b.toolType }

func (b *BlockBreakInfo) GetToolHarvestLevel() int { return b.toolHarvestLevel }

func (b *BlockBreakInfo) IsToolCompatible(tool Item) bool {
	if b.hardness < 0 {
		return false
	}
	return b.toolType == ToolTypeNone || b.toolHarvestLevel == 0 ||
		(b.toolType&tool.GetBlockToolType() != 0 && tool.GetBlockToolHarvestLevel() >= b.toolHarvestLevel)
}

// GetBreakTime returns the seconds needed to break this block with item. Errors if the item's
// mining efficiency isn't positive.
func (b *BlockBreakInfo) GetBreakTime(item Item) (float64, error) {
	base := b.hardness
	if b.IsToolCompatible(item) {
		base *= CompatibleToolMultiplier
	} else {
		base *= IncompatibleToolMultiplier
	}

	efficiency := item.GetMiningEfficiency(b.toolType&item.GetBlockToolType() != 0)
	if efficiency <= 0 {
		return 0, fmt.Errorf("item must have a positive mining efficiency, but got %v", efficiency)
	}

	return base / efficiency, nil
}

func (b *BlockBreakInfo) IsExplosionHarvestable() bool { return b.explosionHarvestable }
