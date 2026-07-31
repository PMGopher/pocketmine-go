package block

// StoneButton is a port of pocketmine\block\StoneButton.
//
// The first real concrete leaf block type in this port (as opposed to testButtonType, which only
// exists to exercise the Behavior pattern in tests). PHP constructs this — like every concrete
// block type — via a registry (VanillaBlocksInputs.php: `new StoneButton($id, "Stone Button", new
// Info(BreakInfo::pickaxe(0.5)))`) that isn't ported yet, so NewStoneButton takes the same
// idInfo/name/typeInfo triple its PHP constructor (inherited from Block) does, to be called from
// that registry once it exists.
type StoneButton struct {
	Button
}

func NewStoneButton(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *StoneButton {
	b := &StoneButton{Button: Button{
		Flowable:        Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		FacingComponent: NewFacingComponent(),
		ActivationTime:  20,
	}}
	b.Init(b)
	return b
}

func (b *StoneButton) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}
