package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// Tripwire is a port of pocketmine\block\Tripwire.
//
// AsItem() should return VanillaItems.String() (a real item, not this block reduced to item
// form) — needs the unported item package/block registry, so it's left as Block's default for
// now (see Block.GetDropsForCompatibleTool's doc comment for the same category of gap).
type Tripwire struct {
	Flowable

	Triggered bool
	Suspended bool // unclear usage, makes hitbox bigger if set
	Connected bool
	Disarmed  bool
}

func NewTripwire(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Tripwire {
	t := &Tripwire{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	t.Init(t)
	return t
}

func (t *Tripwire) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *Tripwire) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.Bool(&t.Triggered)
	w.Bool(&t.Suspended)
	w.Bool(&t.Connected)
	w.Bool(&t.Disarmed)
}

func (t *Tripwire) IsTriggered() bool { return t.Triggered }

func (t *Tripwire) SetTriggered(triggered bool) { t.Triggered = triggered }

func (t *Tripwire) IsSuspended() bool { return t.Suspended }

func (t *Tripwire) SetSuspended(suspended bool) { t.Suspended = suspended }

func (t *Tripwire) IsConnected() bool { return t.Connected }

func (t *Tripwire) SetConnected(connected bool) { t.Connected = connected }

func (t *Tripwire) IsDisarmed() bool { return t.Disarmed }

func (t *Tripwire) SetDisarmed(disarmed bool) { t.Disarmed = disarmed }
