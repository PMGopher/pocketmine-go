package blockutils

// BellAttachmentType is a port of pocketmine\block\utils\BellAttachmentType.
type BellAttachmentType int

const (
	BellAttachmentTypeCeiling BellAttachmentType = iota
	BellAttachmentTypeFloor
	BellAttachmentTypeOneWall
	BellAttachmentTypeTwoWalls
)
