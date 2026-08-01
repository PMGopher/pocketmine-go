package blockutils

// MushroomBlockType is a port of pocketmine\block\utils\MushroomBlockType.
type MushroomBlockType int

const (
	MushroomBlockTypePores MushroomBlockType = iota
	MushroomBlockTypeCapNorthwest
	MushroomBlockTypeCapNorth
	MushroomBlockTypeCapNortheast
	MushroomBlockTypeCapWest
	MushroomBlockTypeCapMiddle
	MushroomBlockTypeCapEast
	MushroomBlockTypeCapSouthwest
	MushroomBlockTypeCapSouth
	MushroomBlockTypeCapSoutheast
	MushroomBlockTypeAllCap
)
