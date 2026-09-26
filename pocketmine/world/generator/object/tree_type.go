package object

// TreeType is a port of pocketmine\world\generator\object\TreeType.
type TreeType int

const (
	TreeTypeOak TreeType = iota
	TreeTypeSpruce
	TreeTypeBirch
	TreeTypeJungle
	TreeTypeAcacia
	TreeTypeDarkOak
	TreeTypeCrimson
	TreeTypeWarped
	TreeTypeAzalea
	//TODO: cherry blossom, mangrove
	//TODO: perhaps huge mushrooms should be here too???
)

// GetDisplayName is a port of TreeType::getDisplayName.
func (t TreeType) GetDisplayName() string {
	switch t {
	case TreeTypeOak:
		return "Oak"
	case TreeTypeSpruce:
		return "Spruce"
	case TreeTypeBirch:
		return "Birch"
	case TreeTypeJungle:
		return "Jungle"
	case TreeTypeAcacia:
		return "Acacia"
	case TreeTypeDarkOak:
		return "Dark Oak"
	case TreeTypeCrimson:
		return "Crimson"
	case TreeTypeWarped:
		return "Warped"
	case TreeTypeAzalea:
		return "Azalea"
	}
	return ""
}
