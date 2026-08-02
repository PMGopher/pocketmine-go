package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/event"
	"pocketmine-go/pocketmine/math"
)

// BambooSapling is a port of pocketmine\block\BambooSapling.
type BambooSapling struct {
	Flowable

	Ready bool
}

func NewBambooSapling(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *BambooSapling {
	b := &BambooSapling{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	b.Init(b)
	return b
}

func (b *BambooSapling) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *BambooSapling) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Bool(&b.Ready) }

func (b *BambooSapling) IsReady() bool { return b.Ready }

func (b *BambooSapling) SetReady(ready bool) { b.Ready = ready }

func (b *BambooSapling) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	geo := support.(blockGeometry)
	return support.GetTypeId() == GRAVEL || geo.HasTypeTag(BlockTypeTagsDirt) ||
		geo.HasTypeTag(BlockTypeTagsMud) || geo.HasTypeTag(BlockTypeTagsSand)
}

func (b *BambooSapling) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return b.canBeSupportedAt(blockReplace) && b.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (b *BambooSapling) OnNearbyBlockChange() {
	if !b.canBeSupportedAt(b.self) {
		if world, err := b.position.GetWorld(); err == nil {
			world.UseBreakOn(b.position.AsVector3())
		}
	} else {
		b.Flowable.OnNearbyBlockChange()
	}
}

// OnInteract's fertilizer/bamboo-item-driven grow needs Fertilizer/item.Bamboo item markers, not
// ported yet. Block's default OnInteract (return false) already matches this gap, so there's
// nothing to override here.

// grow is a port of BambooSapling::grow. Only ever called with a nil player currently
// (OnRandomTick, below) - same reasoning as Bamboo.grow not taking a player parameter yet.
func (b *BambooSapling) grow() bool {
	world, err := b.position.GetWorld()
	if err != nil {
		return false
	}
	if !b.self.(blockGeometry).GetSide(math.Up, 1).CanBeReplaced() {
		return false
	}

	tx := NewBlockTransaction(world)
	bamboo := VanillaBamboo().(*Bamboo)
	above := bamboo.Clone().(*Bamboo)
	above.LeafSize = BambooSmallLeaves
	tx.AddBlock(b.position, bamboo)
	tx.AddBlock(b.position.GetSide(math.Up, 1), above)

	ev := &StructureGrowEvent{Block: b.self, Transaction: tx, Player: nil}
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}

	return tx.Apply()
}

func (b *BambooSapling) TicksRandomly() bool { return true }

func (b *BambooSapling) OnRandomTick() {
	world, err := b.position.GetWorld()
	if err != nil {
		return
	}
	geo := b.self.(blockGeometry)
	above := geo.GetSide(math.Up, 1)

	if b.Ready {
		b.Ready = false
		pos := b.position.AsVector3()
		if world.GetFullLightAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()) < 9 || !b.grow() {
			if err := world.SetBlock(b.position, b.self); err != nil {
				panic(err)
			}
		}
	} else if above.CanBeReplaced() {
		b.Ready = true
		if err := world.SetBlock(b.position, b.self); err != nil {
			panic(err)
		}
	}
}

// AsItem should return VanillaItems.BAMBOO() — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.
