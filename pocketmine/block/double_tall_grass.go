package block

// DoubleTallGrass is a port of pocketmine\block\DoubleTallGrass.
type DoubleTallGrass struct {
	DoublePlant
}

func NewDoubleTallGrass(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *DoubleTallGrass {
	d := &DoubleTallGrass{DoublePlant: DoublePlant{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}}
	d.Init(d)
	return d
}

func (d *DoubleTallGrass) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *DoubleTallGrass) CanBeReplaced() bool { return true }

// GetDropsForIncompatibleTool is a port of DoubleTallGrass::getDropsForIncompatibleTool. The
// bottom half always drops nothing (fully portable). The top half should defer to
// TallGrassTrait's version (a FortuneDropHelper-based wheat seed chance), but that needs the
// unported item package for real Item construction - see Gravel's GetDropsForCompatibleTool doc
// comment for the same category of gap - so both halves return nil for now.
func (d *DoubleTallGrass) GetDropsForIncompatibleTool(item Item) []Item { return nil }

func (d *DoubleTallGrass) GetFlameEncouragement() int { return 60 }

func (d *DoubleTallGrass) GetFlammability() int { return 100 }
