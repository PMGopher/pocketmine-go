package block

// Item is the minimal surface Block/BlockBreakInfo need from an item. Declared locally so the
// future item package satisfies it automatically with no import needed here — the same
// forward-compatible-local-interface pattern already used for Tile, permission.Plugin and
// command.Server elsewhere in this port.
type Item interface {
	GetTypeId() int
	GetBlockToolType() ToolType
	GetBlockToolHarvestLevel() int
	GetMiningEfficiency(isCompatibleToolType bool) float64
	// Pop reduces this item stack's count by one — a base Item method in the PHP original (unlike
	// ApplyDamage, which only applies to Durable items and stays on the separate Axe marker below).
	Pop()
	// IsNull reports whether this represents "no item" (air / zero count) — a base Item method in
	// the PHP original.
	IsNull() bool
}

// ToolTier is the minimal surface needed from an item tool tier (wood/stone/iron/diamond/netherite).
type ToolTier interface {
	GetHarvestLevel() int
}
