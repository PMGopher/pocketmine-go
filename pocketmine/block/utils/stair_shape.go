package blockutils

// StairShape is a port of pocketmine\block\utils\StairShape.
type StairShape int

const (
	StairShapeStraight StairShape = iota
	StairShapeInnerLeft
	StairShapeInnerRight
	StairShapeOuterLeft
	StairShapeOuterRight
)
