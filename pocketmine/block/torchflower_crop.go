package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// TorchflowerCrop is a port of pocketmine\block\TorchflowerCrop.
type TorchflowerCrop struct {
	Flowable

	Ready bool
}

func NewTorchflowerCrop(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *TorchflowerCrop {
	t := &TorchflowerCrop{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	t.Init(t)
	return t
}

func (t *TorchflowerCrop) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *TorchflowerCrop) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Bool(&t.Ready) }

func (t *TorchflowerCrop) IsReady() bool { return t.Ready }

func (t *TorchflowerCrop) SetReady(ready bool) { t.Ready = ready }

func (t *TorchflowerCrop) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetSide(math.Down, 1).GetTypeId() == FARMLAND
}

func (t *TorchflowerCrop) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return t.canBeSupportedAt(blockReplace) && t.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (t *TorchflowerCrop) OnNearbyBlockChange() {
	if !t.canBeSupportedAt(t.self) {
		if world, err := t.position.GetWorld(); err == nil {
			world.UseBreakOn(t.position.AsVector3())
		}
	} else {
		t.Flowable.OnNearbyBlockChange()
	}
}

// getNextState is a port of TorchflowerCrop::getNextState.
func (t *TorchflowerCrop) getNextState() Behavior {
	if t.Ready {
		return VanillaTorchflower()
	}
	crop := VanillaTorchflowerCrop().(*TorchflowerCrop)
	crop.SetReady(true)
	return crop
}

// OnInteract is a port of TorchflowerCrop::onInteract. `$item instanceof Fertilizer` is checked
// via item type ID (bone meal is the only Fertilizer-marked item in the PHP original), same
// structural-marker convention as Crops.OnInteract.
func (t *TorchflowerCrop) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if item.GetTypeId() != itemTypeIDsBoneMeal {
		return false
	}
	if Grow(t.self, t.getNextState(), player) {
		item.Pop()
	}
	return true
}

func (t *TorchflowerCrop) TicksRandomly() bool { return true }

// OnRandomTick is a port of TorchflowerCrop::onRandomTick.
func (t *TorchflowerCrop) OnRandomTick() {
	if CropGrowthCanGrow(t.self) {
		Grow(t.self, t.getNextState(), nil)
	}
}

// AsItem should return VanillaItems.TORCHFLOWER_SEEDS() — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.
