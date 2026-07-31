package blockutils

// DirtType is a port of pocketmine\block\utils\DirtType.
type DirtType int

const (
	DirtTypeNormal DirtType = iota
	DirtTypeCoarse
	DirtTypeRooted
)
