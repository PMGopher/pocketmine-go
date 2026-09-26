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

// AllLeavesTypes is LeavesType::cases().
var AllLeavesTypes = []LeavesType{LeavesTypeOak, LeavesTypeSpruce, LeavesTypeBirch, LeavesTypeJungle, LeavesTypeAcacia, LeavesTypeDarkOak, LeavesTypeMangrove, LeavesTypeAzalea, LeavesTypeFloweringAzalea, LeavesTypeCherry, LeavesTypePaleOak}

// IDName is strtolower($case->name), e.g. "dark_oak" (used for registry names).
func (t LeavesType) IDName() string {
	switch t {
	case LeavesTypeOak:
		return "oak"
	case LeavesTypeSpruce:
		return "spruce"
	case LeavesTypeBirch:
		return "birch"
	case LeavesTypeJungle:
		return "jungle"
	case LeavesTypeAcacia:
		return "acacia"
	case LeavesTypeDarkOak:
		return "dark_oak"
	case LeavesTypeMangrove:
		return "mangrove"
	case LeavesTypeAzalea:
		return "azalea"
	case LeavesTypeFloweringAzalea:
		return "flowering_azalea"
	case LeavesTypeCherry:
		return "cherry"
	case LeavesTypePaleOak:
		return "pale_oak"
	}
	panic("invalid LeavesType value")
}
