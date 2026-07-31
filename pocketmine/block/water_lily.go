package block

import "pocketmine-go/pocketmine/math"

// Water is a forward-compatible marker for pocketmine\block\Water — same pattern as Flowable.go's
// Liquid marker, just narrower (WaterLily needs to specifically distinguish Water, not any
// liquid). The future Water block type just needs to satisfy this trivially.
type Water interface {
	IsWater() bool
}

// WaterLily is a port of pocketmine\block\WaterLily.
type WaterLily struct {
	Flowable
}

func NewWaterLily(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *WaterLily {
	w := &WaterLily{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	w.Init(w)
	return w
}

func (w *WaterLily) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WaterLily) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	bb := math.OneAABB().ContractedCopy(1.0/16, 0, 1.0/16).TrimmedCopy(math.Up, 63.0/64)
	return []math.AxisAlignedBB{bb}
}

func (w *WaterLily) canBeSupportedAt(blk Behavior) bool {
	_, isWater := blk.(blockGeometry).GetSide(math.Down, 1).(Water)
	return isWater
}

func (w *WaterLily) supportedWhenPlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return w.canBeSupportedAt(blockReplace) && w.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (w *WaterLily) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	if _, isWater := blockReplace.(Water); isWater {
		return false
	}
	return w.supportedWhenPlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (w *WaterLily) OnNearbyBlockChange() {
	if !w.canBeSupportedAt(w.self) {
		if world, err := w.position.GetWorld(); err == nil {
			world.UseBreakOn(w.position.AsVector3())
		}
	} else {
		w.Flowable.OnNearbyBlockChange()
	}
}
