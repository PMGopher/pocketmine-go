package block

import "pocketmine-go/pocketmine/math"

// RailFlagAscend is a port of pocketmine\block\utils\RailConnectionInfo::FLAG_ASCEND — ORed into
// a Facing int to mark "this connection ascends" (the rail slopes upward in that direction).
const RailFlagAscend = 1 << 24

// Rail shape constants, a port of pocketmine\data\bedrock\block\BlockLegacyMetadata's RAIL_* consts.
const (
	RailStraightNorthSouth = 0
	RailStraightEastWest   = 1
	RailAscendingEast      = 2
	RailAscendingWest      = 3
	RailAscendingNorth     = 4
	RailAscendingSouth     = 5
	RailCurveSoutheast     = 6
	RailCurveSouthwest     = 7
	RailCurveNorthwest     = 8
	RailCurveNortheast     = 9
)

// railConnections is a port of pocketmine\block\utils\RailConnectionInfo::CONNECTIONS.
var railConnections = map[int][2]int{
	RailStraightNorthSouth: {int(math.North), int(math.South)},
	RailStraightEastWest:   {int(math.East), int(math.West)},
	RailAscendingEast:      {int(math.West), int(math.East) | RailFlagAscend},
	RailAscendingWest:      {int(math.East), int(math.West) | RailFlagAscend},
	RailAscendingNorth:     {int(math.South), int(math.North) | RailFlagAscend},
	RailAscendingSouth:     {int(math.North), int(math.South) | RailFlagAscend},
}

// railCurveConnections is a port of pocketmine\block\utils\RailConnectionInfo::CURVE_CONNECTIONS.
var railCurveConnections = map[int][2]int{
	RailCurveSoutheast: {int(math.South), int(math.East)},
	RailCurveSouthwest: {int(math.South), int(math.West)},
	RailCurveNorthwest: {int(math.North), int(math.West)},
	RailCurveNortheast: {int(math.North), int(math.East)},
}

// railSearchState is a port of BaseRail::searchState: finds the shape key whose 2-element
// connection set matches connections, in either order. connections must have exactly 2 elements
// (guaranteed by railSetConnections before this is ever called).
func railSearchState(connections []int, lookup map[int][2]int) (int, bool) {
	if len(connections) != 2 {
		return 0, false
	}
	for shape, conn := range lookup {
		if conn[0] == connections[0] && conn[1] == connections[1] {
			return shape, true
		}
	}
	for shape, conn := range lookup {
		if conn[0] == connections[1] && conn[1] == connections[0] {
			return shape, true
		}
	}
	return 0, false
}
