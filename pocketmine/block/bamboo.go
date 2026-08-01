package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	BambooNoLeaves    = 0
	BambooSmallLeaves = 1
	BambooLargeLeaves = 2
)

// Bamboo is a port of pocketmine\block\Bamboo.
type Bamboo struct {
	Transparent

	Thick    bool
	Ready    bool
	LeafSize int
}

func NewBamboo(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Bamboo {
	b := &Bamboo{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, LeafSize: BambooNoLeaves}
	b.Init(b)
	return b
}

func (b *Bamboo) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Bamboo) DescribeBlockOnlyState(w runtime.DataDescriber) {
	leafSize := b.LeafSize
	w.BoundedIntAuto(BambooNoLeaves, BambooLargeLeaves, &leafSize)
	b.LeafSize = leafSize
	w.Bool(&b.Thick)
	w.Bool(&b.Ready)
}

func (b *Bamboo) IsThick() bool { return b.Thick }

func (b *Bamboo) SetThick(thick bool) { b.Thick = thick }

func (b *Bamboo) IsReady() bool { return b.Ready }

func (b *Bamboo) SetReady(ready bool) { b.Ready = ready }

func (b *Bamboo) GetLeafSize() int { return b.LeafSize }

func (b *Bamboo) SetLeafSize(leafSize int) { b.LeafSize = leafSize }

func (b *Bamboo) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	// This places the BB at the northwest corner, not the center.
	thickness := 2.0
	if b.Thick {
		thickness = 3.0
	}
	inset := 1 - thickness/16
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.South, inset).TrimmedCopy(math.East, inset)}
}

func (b *Bamboo) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

// bambooOffsetSeed is a port of Bamboo::getOffsetSeed. The PHP original uses GMP (arbitrary
// precision) purely to get exact two's-complement bitwise/multiplication behaviour before masking
// to 32 bits; Go's int64 arithmetic wraps in two's complement the same way on overflow (this is
// guaranteed by the language spec), and since the final result is masked to the low 32 bits,
// truncating intermediate products to 64 bits first doesn't change that low-32-bit outcome - 2^64
// is a multiple of 2^32, so arithmetic mod 2^64 preserves the true result mod 2^32 exactly.
func bambooOffsetSeed(x, y, z int) uint32 {
	p1 := int64(z) * 0x6ebfff5
	p2 := int64(x) * 0x2fc20f
	p3 := int64(y)
	xord := (p1 ^ p2) ^ p3
	fullResult := (xord*0x285b825 + 0xb) * xord
	return uint32(fullResult & 0xffffffff)
}

func bambooMaxHeight(x, z int) int {
	return 12 + int(bambooOffsetSeed(x, 0, z)%5)
}

func (b *Bamboo) GetModelPositionOffset() (math.Vector3, bool) {
	seed := bambooOffsetSeed(b.position.FloorX(), 0, b.position.FloorZ())
	retX := float64((seed%12)+1) / 16
	retZ := float64(((seed>>8)%12)+1) / 16
	return math.Vector3{X: retX, Y: 0, Z: retZ}, true
}

func (b *Bamboo) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	geo := support.(blockGeometry)
	return geo.HasSameTypeId(b.self) || support.GetTypeId() == GRAVEL ||
		geo.HasTypeTag(BlockTypeTagsDirt) || geo.HasTypeTag(BlockTypeTagsMud) || geo.HasTypeTag(BlockTypeTagsSand)
}

func (b *Bamboo) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return b.canBeSupportedAt(blockReplace) && b.Transparent.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (b *Bamboo) OnNearbyBlockChange() {
	if !b.canBeSupportedAt(b.self) {
		if world, err := b.position.GetWorld(); err == nil {
			world.UseBreakOn(b.position.AsVector3())
		}
	} else {
		b.Transparent.OnNearbyBlockChange()
	}
}

// seekToTop is a port of Bamboo::seekToTop.
func (b *Bamboo) seekToTop() *Bamboo {
	top := b
	for {
		next, ok := top.self.(blockGeometry).GetSide(math.Up, 1).(*Bamboo)
		if !ok || !next.HasSameTypeId(top.self) {
			break
		}
		top = next
	}
	return top
}

// OnInteract's fertilizer/bamboo-item-driven grow needs Fertilizer/item.Bamboo item markers, not
// ported yet. Block's default OnInteract (return false) already matches this gap, so there's
// nothing to override here.

// grow is a port of Bamboo::grow. It needs a fresh per-call BlockTransaction, the block registry,
// and StructureGrowEvent, none ported yet, so this is a no-op stub returning false for now - see
// Sugarcane.grow's doc comment for the same category of gap.
func (b *Bamboo) grow(maxHeight int, growAmount int) bool { return false }

func (b *Bamboo) TicksRandomly() bool { return true }

func (b *Bamboo) OnRandomTick() {
	world, err := b.position.GetWorld()
	if err != nil {
		return
	}
	if b.Ready {
		b.Ready = false
		pos := b.position.AsVector3()
		maxHeight := bambooMaxHeight(pos.FloorX(), pos.FloorZ())
		if world.GetFullLightAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()) < 9 || !b.grow(maxHeight, 1) {
			if err := world.SetBlock(b.position, b.self); err != nil {
				panic(err)
			}
		}
	} else if b.self.(blockGeometry).GetSide(math.Up, 1).CanBeReplaced() {
		b.Ready = true
		if err := world.SetBlock(b.position, b.self); err != nil {
			panic(err)
		}
	}
}

// AsItem should return VanillaItems.BAMBOO() — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.
