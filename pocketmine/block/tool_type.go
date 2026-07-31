package block

// ToolType is a port of pocketmine\block\BlockToolType: bitflags for the tool types that can be
// used to break a block. A block may allow multiple tool types by combining these.
type ToolType int

const (
	ToolTypeNone    ToolType = 0
	ToolTypeSword   ToolType = 1 << 0
	ToolTypeShovel  ToolType = 1 << 1
	ToolTypePickaxe ToolType = 1 << 2
	ToolTypeAxe     ToolType = 1 << 3
	ToolTypeShears  ToolType = 1 << 4
	ToolTypeHoe     ToolType = 1 << 5
)
