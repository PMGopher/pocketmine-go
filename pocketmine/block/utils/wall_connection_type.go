package blockutils

// WallConnectionType is a port of pocketmine\block\utils\WallConnectionType.
type WallConnectionType int

const (
	WallConnectionTypeShort WallConnectionType = iota
	WallConnectionTypeTall
)
