package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	PinkPetalsMinCount = 1
	PinkPetalsMaxCount = 4
)

// PinkPetals is a port of pocketmine\block\PinkPetals.
type PinkPetals struct {
	Flowable
	HorizontalFacingComponent

	Count int
}

func NewPinkPetals(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *PinkPetals {
	p := &PinkPetals{
		Flowable:                  Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
		Count:                     PinkPetalsMinCount,
	}
	p.Init(p)
	return p
}

func (p *PinkPetals) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}

func (p *PinkPetals) DescribeBlockOnlyState(w runtime.DataDescriber) {
	p.DescribeHorizontalFacing(w)
	w.BoundedIntAuto(PinkPetalsMinCount, PinkPetalsMaxCount, &p.Count)
}

func (p *PinkPetals) GetCount() int { return p.Count }

// SetCount panics if count is out of range, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (p *PinkPetals) SetCount(count int) {
	if count < PinkPetalsMinCount || count > PinkPetalsMaxCount {
		panic("Count must be in range 1 ... 4")
	}
	p.Count = count
}

func (p *PinkPetals) canBeSupportedAt(blk Behavior) bool {
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	geo := support.(blockGeometry)
	return geo.HasTypeTag(BlockTypeTagsDirt) || geo.HasTypeTag(BlockTypeTagsMud)
}

func (p *PinkPetals) supportedWhenPlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return p.canBeSupportedAt(blockReplace) && p.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (p *PinkPetals) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	if replace, ok := blockReplace.(*PinkPetals); ok && replace.Count < PinkPetalsMaxCount {
		return true
	}
	return p.supportedWhenPlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (p *PinkPetals) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if replace, ok := blockReplace.(*PinkPetals); ok && replace.Count < PinkPetalsMaxCount {
		p.Count = replace.Count + 1
		p.Facing = replace.Facing
	} else if player != nil {
		p.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return p.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract should grow the plant (BlockEventHelper.Grow) or drop an item copy when fertilized —
// needs Item.Fertilizer type-checking and World.DropItem from the unported item package, so this
// is a no-op for now (see Block.GetDropsForCompatibleTool's doc comment for the same kind of gap).
func (p *PinkPetals) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return false
}

func (p *PinkPetals) GetFlameEncouragement() int { return 60 }

func (p *PinkPetals) GetFlammability() int { return 100 }

// GetDropsForCompatibleTool should return [p.AsItem().SetCount(p.Count)] — needs real Item
// construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (p *PinkPetals) GetDropsForCompatibleTool(item Item) []Item { return nil }
