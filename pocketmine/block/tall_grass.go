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
	DoublePlantVariant func() Behavior
}

func NewTallGrass(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, doublePlantVariant func() Behavior) *TallGrass {
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

// GetDropsForIncompatibleTool is a port of TallGrassTrait::getDropsForIncompatibleTool.
func (t *TallGrass) GetDropsForIncompatibleTool(item Item) []Item {
	return tallGrassDropsForIncompatibleTool(item)
}

// tallGrassDropsForIncompatibleTool is TallGrassTrait::getDropsForIncompatibleTool.
func tallGrassDropsForIncompatibleTool(item Item) []Item {
	if FortuneBonusChanceDivisor(item, 8, 2) {
		return itemDrops(vanillaItem("wheat_seeds"))
	}
	return nil
}

// OnInteract is a port of TallGrass::onInteract: bone meal grows it into its double plant.
func (t *TallGrass) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	world, err := t.position.GetWorld()
	if err != nil {
		return false
	}
	upX, upY, upZ := t.position.FloorX(), t.position.FloorY()+1, t.position.FloorZ()
	if !world.IsInWorld(upX, upY, upZ) || t.self.(blockGeometry).GetSide(math.Up, 1).GetTypeId() != AIR {
		return false
	}
	if item.GetTypeId() == itemTypeIDsBoneMeal && t.DoublePlantVariant != nil {
		if doubleVariant := t.DoublePlantVariant(); doubleVariant != nil {
			bottom := doubleVariant.Clone()
			bottom.(interface{ SetTop(bool) }).SetTop(false)
			top := doubleVariant.Clone()
			top.(interface{ SetTop(bool) }).SetTop(true)
			_ = world.SetBlock(t.position, bottom)
			_ = world.SetBlock(NewPosition(float64(upX), float64(upY), float64(upZ), world), top)
			item.Pop()
			return true
		}
	}
	return false
}
