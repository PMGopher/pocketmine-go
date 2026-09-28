package block

// StoneButton is a port of pocketmine\block\StoneButton.
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
