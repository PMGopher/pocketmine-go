package blockutils

// WoodType is a port of pocketmine\block\utils\WoodType.
type WoodType int

const (
	WoodTypeOak WoodType = iota
	WoodTypeSpruce
	WoodTypeBirch
	WoodTypeJungle
	WoodTypeAcacia
	WoodTypeDarkOak
	WoodTypeMangrove
	WoodTypeCrimson
	WoodTypeWarped
	WoodTypeCherry
	WoodTypePaleOak
	WoodTypeBamboo
)

func (w WoodType) GetDisplayName() string {
	switch w {
	case WoodTypeOak:
		return "Oak"
	case WoodTypeSpruce:
		return "Spruce"
	case WoodTypeBirch:
		return "Birch"
	case WoodTypeJungle:
		return "Jungle"
	case WoodTypeAcacia:
		return "Acacia"
	case WoodTypeDarkOak:
		return "Dark Oak"
	case WoodTypeMangrove:
		return "Mangrove"
	case WoodTypeCrimson:
		return "Crimson"
	case WoodTypeWarped:
		return "Warped"
	case WoodTypeCherry:
		return "Cherry"
	case WoodTypePaleOak:
		return "Pale Oak"
	case WoodTypeBamboo:
		return "Bamboo"
	default:
		panic("invalid WoodType value")
	}
}

func (w WoodType) IsFlammable() bool {
	return w != WoodTypeCrimson && w != WoodTypeWarped
}

// GetStandardLogSuffix returns the suffix used in a standard (bark-on-the-side) log's name, and
// ok=false where the PHP original returns null (no suffix).
func (w WoodType) GetStandardLogSuffix() (suffix string, ok bool) {
	switch w {
	case WoodTypeCrimson, WoodTypeWarped:
		return "Stem", true
	case WoodTypeBamboo:
		return "Block", true
	default:
		return "", false
	}
}

// GetAllSidedLogSuffix returns the suffix used in an all-sided (bark-everywhere) log's name, and
// ok=false where the PHP original returns null.
func (w WoodType) GetAllSidedLogSuffix() (suffix string, ok bool) {
	if w == WoodTypeCrimson || w == WoodTypeWarped {
		return "Hyphae", true
	}
	return "", false
}
