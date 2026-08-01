package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// TripwireHook is a port of pocketmine\block\TripwireHook.
type TripwireHook struct {
	Flowable
	HorizontalFacingComponent

	Connected bool
	Powered   bool
}

func NewTripwireHook(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *TripwireHook {
	t := &TripwireHook{
		Flowable:                  Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	t.Init(t)
	return t
}

func (t *TripwireHook) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *TripwireHook) DescribeBlockOnlyState(w runtime.DataDescriber) {
	t.DescribeHorizontalFacing(w)
	w.Bool(&t.Connected)
	w.Bool(&t.Powered)
}

func (t *TripwireHook) IsConnected() bool { return t.Connected }

func (t *TripwireHook) SetConnected(connected bool) { t.Connected = connected }

func (t *TripwireHook) IsPowered() bool { return t.Powered }

func (t *TripwireHook) SetPowered(powered bool) { t.Powered = powered }

// Place: TODO (from the PHP original) check face is valid.
func (t *TripwireHook) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if math.FacingAxis(face) != math.AxisY {
		t.Facing = face
		return t.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
	}
	return false
}
