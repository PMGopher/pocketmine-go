package block

// Item is the minimal surface Block/BlockBreakInfo need from an item. Declared locally so the
// future item package satisfies it automatically with no import needed here — the same
// forward-compatible-local-interface pattern already used for Tile, permission.Plugin and
// command.Server elsewhere in this port.
type Item interface {
	GetBlockToolType() ToolType
	GetBlockToolHarvestLevel() int
	GetMiningEfficiency(isCompatibleToolType bool) float64
}

// ToolTier is the minimal surface needed from an item tool tier (wood/stone/iron/diamond/netherite).
type ToolTier interface {
	GetHarvestLevel() int
}
