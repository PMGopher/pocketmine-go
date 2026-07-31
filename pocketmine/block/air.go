package block

// Air is a port of pocketmine\block\Air.
type Air struct {
	Flowable
}

func NewAir(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Air {
	a := &Air{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	a.Init(a)
	return a
}

func (a *Air) Clone() Behavior {
	c := *a
	c.rebind(&c)
	return &c
}

func (a *Air) CanBeReplaced() bool { return true }

func (a *Air) CanBePlaced() bool { return false }
