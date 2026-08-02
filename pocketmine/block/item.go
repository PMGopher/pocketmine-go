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
	// GetCustomName is needed by Chest.OnInteract's CanOpenWith(item.GetCustomName()) check — a
	// base Item method in the PHP original.
	GetCustomName() string
	// GetCount/SetCount are needed by AsItem() callers that scale a drop's stack size (e.g.
	// mushroom blocks' random 0-2 count, ore blocks' fortune-scaled count) - base Item
	// methods in the PHP original.
	GetCount() int
	SetCount(count int)
}

// NewItemBlockFunc constructs a real Item wrapping the given block, using item type ID
// -blk.GetTypeId() (ItemTypeIds::fromBlockTypeId in the PHP original is exactly this negation,
// not a lookup table - see ItemTypeIds.php). The item package can't be imported directly here (it
// already imports block for ItemBlock/tool types, an import block can't reverse without a cycle),
// so the item package's own init() sets this instead - the standard Go dependency-inversion
// pattern for breaking a two-way need between packages. Nil until the item package is imported
// somewhere in the running program; AsItem() handles that case explicitly.
var NewItemBlockFunc func(blk Behavior) Item

// ToolTier is the minimal surface needed from an item tool tier (wood/stone/iron/diamond/netherite).
type ToolTier interface {
	GetHarvestLevel() int
}
