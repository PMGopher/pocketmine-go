package blockutils

// SaplingType is a port of pocketmine\block\utils\SaplingType. The getTreeType() mapping to
// pocketmine\world\generator\object\TreeType isn't ported here since the world-gen package
// (TreeFactory/TreeType) isn't ported yet - see Sapling.grow's doc comment.
type SaplingType int

const (
	SaplingTypeOak SaplingType = iota
	SaplingTypeSpruce
	SaplingTypeBirch
	SaplingTypeJungle
	SaplingTypeAcacia
	SaplingTypeDarkOak
)

// AllSaplingTypes is SaplingType::cases().
var AllSaplingTypes = []SaplingType{SaplingTypeOak, SaplingTypeSpruce, SaplingTypeBirch, SaplingTypeJungle, SaplingTypeAcacia, SaplingTypeDarkOak}

// IDName is strtolower($case->name), e.g. "dark_oak" (used for registry names).
func (t SaplingType) IDName() string {
	switch t {
	case SaplingTypeOak:
		return "oak"
	case SaplingTypeSpruce:
		return "spruce"
	case SaplingTypeBirch:
		return "birch"
	case SaplingTypeJungle:
		return "jungle"
	case SaplingTypeAcacia:
		return "acacia"
	case SaplingTypeDarkOak:
		return "dark_oak"
	}
	panic("invalid SaplingType value")
}

// GetDisplayName is a port of SaplingType::getDisplayName (TreeType::getDisplayName of its tree).
func (t SaplingType) GetDisplayName() string {
	switch t {
	case SaplingTypeOak:
		return "Oak"
	case SaplingTypeSpruce:
		return "Spruce"
	case SaplingTypeBirch:
		return "Birch"
	case SaplingTypeJungle:
		return "Jungle"
	case SaplingTypeAcacia:
		return "Acacia"
	case SaplingTypeDarkOak:
		return "Dark Oak"
	}
	panic("invalid SaplingType value")
}
