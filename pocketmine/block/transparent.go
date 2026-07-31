package block

// Transparent is a port of pocketmine\block\Transparent.
//
// Transparent blocks do not block any light from propagating through them. This does NOT imply
// that the block is visually transparent — chests allow light through but you can't see through
// them except at the edges.
type Transparent struct {
	Block
}

func (t *Transparent) IsTransparent() bool { return true }
