package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// SporeBlossom is a port of pocketmine\block\SporeBlossom.
//
// PHP's StaticSupportTrait provides CanBePlacedAt/OnNearbyBlockChange in terms of an abstract
// canBeSupportedAt(Block) - see Flower's doc comment for why this is inlined per type rather than
// shared.
type SporeBlossom struct {
	Flowable
}

func NewSporeBlossom(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SporeBlossom {
	s := &SporeBlossom{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	s.Init(s)
	return s
}

func (s *SporeBlossom) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SporeBlossom) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Up) == blockutils.SupportTypeFull
}

func (s *SporeBlossom) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return s.canBeSupportedAt(blockReplace) && s.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (s *SporeBlossom) OnNearbyBlockChange() {
	if !s.canBeSupportedAt(s.self) {
		if world, err := s.position.GetWorld(); err == nil {
			world.UseBreakOn(s.position.AsVector3())
		}
	} else {
		s.Flowable.OnNearbyBlockChange()
	}
}
