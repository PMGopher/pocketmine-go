package block

import "pocketmine-go/pocketmine/math"

// Flower is a port of pocketmine\block\Flower.
//
// PHP's StaticSupportTrait provides CanBePlacedAt/OnNearbyBlockChange in terms of an abstract
// canBeSupportedAt(Block) — traits are copy-pasted per using class rather than inherited, so
// there's no virtual-dispatch pitfall to route through Behavior for this; each concrete type using
// it (Flower, DeadBush, WitherRose, ...) just implements the same two methods directly, calling
// its own canBeSupportedAt.
type Flower struct {
	Flowable
}

func NewFlower(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Flower {
	f := &Flower{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	f.Init(f)
	return f
}

func (f *Flower) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *Flower) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1).(blockGeometry)
	return support.HasTypeTag(BlockTypeTagsDirt) || support.HasTypeTag(BlockTypeTagsMud)
}

func (f *Flower) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return f.canBeSupportedAt(blockReplace) && f.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (f *Flower) OnNearbyBlockChange() {
	if !f.canBeSupportedAt(f.self) {
		if world, err := f.position.GetWorld(); err == nil {
			world.UseBreakOn(f.position.AsVector3())
		}
	} else {
		f.Flowable.OnNearbyBlockChange()
	}
}

func (f *Flower) GetFlameEncouragement() int { return 60 }

func (f *Flower) GetFlammability() int { return 100 }
