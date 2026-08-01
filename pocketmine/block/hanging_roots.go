package block

import "pocketmine-go/pocketmine/math"

// HangingRoots is a port of pocketmine\block\HangingRoots.
type HangingRoots struct {
	Flowable
}

func NewHangingRoots(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *HangingRoots {
	h := &HangingRoots{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	h.Init(h)
	return h
}

func (h *HangingRoots) Clone() Behavior {
	c := *h
	c.rebind(&c)
	return &c
}

// canBeSupportedAt is deliberately checking Up (not Down) - matches the PHP original's comment:
// "weird I know, but they can be placed on the bottom of fences".
func (h *HangingRoots) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Up).HasCenterSupport()
}

func (h *HangingRoots) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return h.canBeSupportedAt(blockReplace) && h.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (h *HangingRoots) OnNearbyBlockChange() {
	if !h.canBeSupportedAt(h.self) {
		if world, err := h.position.GetWorld(); err == nil {
			world.UseBreakOn(h.position.AsVector3())
		}
	} else {
		h.Flowable.OnNearbyBlockChange()
	}
}

// GetDropsForIncompatibleTool should check Item.HasEnchantment(SilkTouch) — needs the unported
// enchantment package, so this always returns nil (matching the "no silk touch" branch) for now.
func (h *HangingRoots) GetDropsForIncompatibleTool(item Item) []Item { return nil }
