package block

import blockutils "pocketmine-go/pocketmine/block/utils"

// WoodenButton is a port of pocketmine\block\WoodenButton.
type WoodenButton struct {
	Button
	WoodTypeComponent
}

func NewWoodenButton(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *WoodenButton {
	b := &WoodenButton{
		Button: Button{
			Flowable:        Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
			FacingComponent: NewFacingComponent(),
			ActivationTime:  30,
		},
		WoodTypeComponent: NewWoodTypeComponent(woodType),
	}
	b.Init(b)
	return b
}

func (b *WoodenButton) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

// HasEntityCollision is false pending arrow-activation support (matches the PHP original's
// `//TODO: arrows activate wooden buttons`).
func (b *WoodenButton) HasEntityCollision() bool { return false }

func (b *WoodenButton) GetFuelTime() int {
	if b.WoodType.IsFlammable() {
		return 100
	}
	return 0
}
