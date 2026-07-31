package block

import "pocketmine-go/pocketmine/math"

// WitherRose is a port of pocketmine\block\WitherRose.
type WitherRose struct {
	Flowable
}

func NewWitherRose(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *WitherRose {
	w := &WitherRose{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	w.Init(w)
	return w
}

func (w *WitherRose) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WitherRose) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	geo := support.(blockGeometry)
	if geo.HasTypeTag(BlockTypeTagsDirt) || geo.HasTypeTag(BlockTypeTagsMud) {
		return true
	}
	switch support.GetTypeId() {
	case NETHERRACK, SOUL_SAND, SOUL_SOIL:
		return true
	default:
		return false
	}
}

func (w *WitherRose) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return w.canBeSupportedAt(blockReplace) && w.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (w *WitherRose) OnNearbyBlockChange() {
	if !w.canBeSupportedAt(w.self) {
		if world, err := w.position.GetWorld(); err == nil {
			world.UseBreakOn(w.position.AsVector3())
		}
	} else {
		w.Flowable.OnNearbyBlockChange()
	}
}

func (w *WitherRose) HasEntityCollision() bool { return true }

// OnEntityInside should apply the Wither effect to Living entities that don't already have it —
// needs entity.Living/EffectInstance/VanillaEffects from the unported entity package, so this is
// currently a no-op (matches the "entity just doesn't take the effect yet" gap already present
// for other entity-dependent block behaviors like Cobweb.OnEntityInside).
func (w *WitherRose) OnEntityInside(entity Entity) bool { return true }

func (w *WitherRose) GetFlameEncouragement() int { return 60 }

func (w *WitherRose) GetFlammability() int { return 100 }
