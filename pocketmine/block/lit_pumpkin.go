package block

// LitPumpkin is a port of pocketmine\block\LitPumpkin.
type LitPumpkin struct {
	CarvedPumpkin
}

func NewLitPumpkin(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *LitPumpkin {
	l := &LitPumpkin{
		CarvedPumpkin{
			Opaque:                    Opaque{NewBlock(idInfo, name, typeInfo)},
			HorizontalFacingComponent: NewHorizontalFacingComponent(),
		},
	}
	l.Init(l)
	return l
}

func (l *LitPumpkin) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *LitPumpkin) GetLightLevel() int { return 15 }
