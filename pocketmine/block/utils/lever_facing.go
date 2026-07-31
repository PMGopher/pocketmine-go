package blockutils

import "pocketmine-go/pocketmine/math"

// LeverFacing is a port of pocketmine\block\utils\LeverFacing.
type LeverFacing int

const (
	LeverFacingUpAxisX LeverFacing = iota
	LeverFacingUpAxisZ
	LeverFacingDownAxisX
	LeverFacingDownAxisZ
	LeverFacingNorth
	LeverFacingEast
	LeverFacingSouth
	LeverFacingWest
)

func (l LeverFacing) GetFacing() math.Facing {
	switch l {
	case LeverFacingUpAxisX, LeverFacingUpAxisZ:
		return math.Up
	case LeverFacingDownAxisX, LeverFacingDownAxisZ:
		return math.Down
	case LeverFacingNorth:
		return math.North
	case LeverFacingEast:
		return math.East
	case LeverFacingSouth:
		return math.South
	case LeverFacingWest:
		return math.West
	default:
		panic("invalid LeverFacing value")
	}
}
