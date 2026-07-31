package block

// TintedGlass is a port of pocketmine\block\TintedGlass.
type TintedGlass struct {
	Transparent
}

func NewTintedGlass(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *TintedGlass {
	t := &TintedGlass{Transparent{NewBlock(idInfo, name, typeInfo)}}
	t.Init(t)
	return t
}

func (t *TintedGlass) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *TintedGlass) GetLightFilter() int { return 15 }
