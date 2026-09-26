package bedrockblock

// Block state properties the 1.26.50 palette has but PocketMine-MP 5.44.4's (1.26.30)
// BlockStateNames doesn't: stairs got a corner shape and fences, glass panes, bars and tripwire got
// connection flags.
const (
	MC_CORNER           = "minecraft:corner"
	MC_CONNECTION_EAST  = "minecraft:connection_east"
	MC_CONNECTION_NORTH = "minecraft:connection_north"
	MC_CONNECTION_SOUTH = "minecraft:connection_south"
	MC_CONNECTION_WEST  = "minecraft:connection_west"

	MC_CORNER_NONE        = "none"
	MC_CORNER_INNER_LEFT  = "inner_left"
	MC_CORNER_INNER_RIGHT = "inner_right"
	MC_CORNER_OUTER_LEFT  = "outer_left"
	MC_CORNER_OUTER_RIGHT = "outer_right"
)
