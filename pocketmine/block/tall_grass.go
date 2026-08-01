package block

import "pocketmine-go/pocketmine/math"

// TallGrass is a port of pocketmine\block\TallGrass.
//
// PHP's StaticSupportTrait provides CanBePlacedAt/OnNearbyBlockChange in terms of an abstract
// canBeSupportedAt(Block) - see Flower's doc comment for why this is inlined per type rather than
// shared.
type TallGrass struct {
	Flowable

	// DoublePlantVariant mirrors the PHP constructor's `?Closure(): DoublePlant` - the double-tall
	// plant this grows into when fertilized (nil for grass with no double variant).
	DoublePlantVariant func() *DoublePlant
}

func NewTallGrass(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, doublePlantVariant func() *DoublePlant) *TallGrass {
	t := &TallGrass{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, DoublePlantVariant: doublePlantVariant}
	t.Init(t)
	return t
}

func (t *TallGrass) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *TallGrass) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1).(blockGeometry)
	return support.HasTypeTag(BlockTypeTagsDirt) || support.HasTypeTag(BlockTypeTagsMud)
}

func (t *TallGrass) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return t.canBeSupportedAt(blockReplace) && t.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (t *TallGrass) OnNearbyBlockChange() {
	if !t.canBeSupportedAt(t.self) {
		if world, err := t.position.GetWorld(); err == nil {
			world.UseBreakOn(t.position.AsVector3())
		}
	} else {
		t.Flowable.OnNearbyBlockChange()
	}
}

func (t *TallGrass) GetFlameEncouragement() int { return 60 }

func (t *TallGrass) GetFlammability() int { return 100 }

// GetDropsForIncompatibleTool's wheat-seed chance (TallGrassTrait, via FortuneDropHelper) needs
// the unported item package for real Item construction - see Gravel's GetDropsForCompatibleTool
// doc comment for the same category of gap - so this returns nil for now.
func (t *TallGrass) GetDropsForIncompatibleTool(item Item) []Item { return nil }

// OnInteract should grow into DoublePlantVariant when fertilized - needs a Fertilizer item
// marker, not ported yet, so this is a no-op for now (same gap category as Azalea's OnInteract).
func (t *TallGrass) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return false
}
