package object

import "pocketmine-go/pocketmine/utils"

// TreeType is a port of the slice of pocketmine\world\generator\object\TreeType this port's
// TreeFactory actually needs - only the 3 species any registered biome (Mountains/Forest/Taiga)
// populates trees with. Jungle/Acacia/DarkOak/Crimson/Warped/Azalea aren't ported (see Tree's doc
// comment).
type TreeType int

const (
	TreeTypeOak TreeType = iota
	TreeTypeSpruce
	TreeTypeBirch
)

// NewTreeFromType is a port of the slice of TreeFactory::get this port needs (real TreeFactory
// also builds big-oak/jungle/acacia/azalea/nether trees, none of which are ported).
func NewTreeFromType(random *utils.Random, t TreeType) *Tree {
	switch t {
	case TreeTypeSpruce:
		return NewSpruceTree()
	case TreeTypeBirch:
		return NewBirchTree(random.NextBoundedInt(39) == 0)
	default:
		return NewOakTree()
	}
}
