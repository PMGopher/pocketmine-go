package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	SeaPickleMinCount = 1
	SeaPickleMaxCount = 4
)

// SeaPickle is a port of pocketmine\block\SeaPickle.
type SeaPickle struct {
	Transparent

	Count      int
	Underwater bool
}

func NewSeaPickle(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SeaPickle {
	s := &SeaPickle{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Count: SeaPickleMinCount}
	s.Init(s)
	return s
}

func (s *SeaPickle) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SeaPickle) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.BoundedIntAuto(SeaPickleMinCount, SeaPickleMaxCount, &s.Count)
	w.Bool(&s.Underwater)
}

func (s *SeaPickle) GetCount() int { return s.Count }

// SetCount panics if count is out of range, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (s *SeaPickle) SetCount(count int) {
	if count < SeaPickleMinCount || count > SeaPickleMaxCount {
		panic("Count must be in range 1 ... 4")
	}
	s.Count = count
}

func (s *SeaPickle) IsUnderwater() bool { return s.Underwater }

func (s *SeaPickle) SetUnderwater(underwater bool) { s.Underwater = underwater }

func (s *SeaPickle) IsSolid() bool { return false }

func (s *SeaPickle) GetLightLevel() int {
	if s.Underwater {
		return (s.Count + 1) * 3
	}
	return 0
}

func (s *SeaPickle) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (s *SeaPickle) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

// CanBePlacedAt: TODO (from the PHP original) proper placement logic needs a supporting face below.
func (s *SeaPickle) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	if replace, ok := blockReplace.(*SeaPickle); ok && replace.Count < SeaPickleMaxCount {
		return true
	}
	return s.Transparent.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

// Place doesn't set Underwater=true yet — TODO (from the PHP original): implement once new water
// logic is in place.
func (s *SeaPickle) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	s.Underwater = false
	if replace, ok := blockReplace.(*SeaPickle); ok && replace.Count < SeaPickleMaxCount {
		s.Count = replace.Count + 1
	}
	return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract: TODO (from the PHP original) bonemeal logic requires coral, not yet ported.
func (s *SeaPickle) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return s.Block.OnInteract(item, face, clickVector, player, returnedItems)
}

// GetDropsForCompatibleTool should return [s.AsItem().SetCount(s.Count)] — needs real Item
// construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (s *SeaPickle) GetDropsForCompatibleTool(item Item) []Item { return nil }
