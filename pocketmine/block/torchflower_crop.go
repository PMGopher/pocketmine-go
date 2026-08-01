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

// getNextState would return VanillaBlocks.TORCHFLOWER() or VanillaBlocks.TORCHFLOWER_CROP() with
// Ready set - needs the unported block registry, same gap as Sapling.grow, so it's not ported.

// OnInteract's fertilizer-driven growth (via BlockEventHelper.Grow) needs a Fertilizer item
// marker and the block registry, neither ported yet - same gap documented on
// Crops/SweetBerryBush/CocoaBlock/Sapling's OnInteract. Block's default OnInteract (return false)
// already matches this gap, so there's nothing to override here.

func (t *TorchflowerCrop) TicksRandomly() bool { return true }

// OnRandomTick should use CropGrowthHelper.CanGrow and BlockEventHelper.Grow (block/utils, not
// ported) to decide whether to advance to the next growth state - same gap as
// NetherWartPlant.OnRandomTick, so this is a no-op for now.
func (t *TorchflowerCrop) OnRandomTick() {}

// AsItem should return VanillaItems.TORCHFLOWER_SEEDS() — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.
