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
