package blockutils

// LeavesType is a port of pocketmine\block\utils\LeavesType.
type LeavesType int

const (
	LeavesTypeOak LeavesType = iota
	LeavesTypeSpruce
	LeavesTypeBirch
	LeavesTypeJungle
	LeavesTypeAcacia
	LeavesTypeDarkOak
	LeavesTypeMangrove
	LeavesTypeAzalea
	LeavesTypeFloweringAzalea
	LeavesTypeCherry
	LeavesTypePaleOak
)

func (l LeavesType) GetDisplayName() string {
	switch l {
	case LeavesTypeOak:
		return "Oak"
	case LeavesTypeSpruce:
		return "Spruce"
	case LeavesTypeBirch:
		return "Birch"
	case LeavesTypeJungle:
		return "Jungle"
	case LeavesTypeAcacia:
		return "Acacia"
	case LeavesTypeDarkOak:
		return "Dark Oak"
	case LeavesTypeMangrove:
		return "Mangrove"
	case LeavesTypeAzalea:
		return "Azalea"
	case LeavesTypeFloweringAzalea:
		return "Flowering Azalea"
	case LeavesTypeCherry:
		return "Cherry"
	case LeavesTypePaleOak:
		return "Pale Oak"
	default:
		panic("invalid LeavesType value")
	}
}
