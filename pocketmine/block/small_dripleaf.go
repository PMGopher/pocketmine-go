package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// SmallDripleaf is a port of pocketmine\block\SmallDripleaf.
type SmallDripleaf struct {
	Transparent
	HorizontalFacingComponent

	Top bool
}

func NewSmallDripleaf(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SmallDripleaf {
	s := &SmallDripleaf{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	s.Init(s)
	return s
}

func (s *SmallDripleaf) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SmallDripleaf) DescribeBlockOnlyState(w runtime.DataDescriber) {
	s.DescribeHorizontalFacing(w)
	w.Bool(&s.Top)
}

func (s *SmallDripleaf) IsTop() bool { return s.Top }

func (s *SmallDripleaf) SetTop(top bool) { s.Top = top }

// canBeSupportedBy is a port of SmallDripleaf::canBeSupportedBy. Moss block support, and support
// from (waterlogged) dirt/grass/podzol/etc, are both TODOs in the PHP original too.
func (s *SmallDripleaf) canBeSupportedBy(blk Behavior) bool {
	return blk.GetTypeId() == CLAY
}

func (s *SmallDripleaf) OnNearbyBlockChange() {
	geo := s.self.(blockGeometry)
	if !s.Top && !s.canBeSupportedBy(geo.GetSide(math.Down, 1)) {
		if world, err := s.position.GetWorld(); err == nil {
			world.UseBreakOn(s.position.AsVector3())
		}
		return
	}
	face := math.Up
	if s.Top {
		face = math.Down
	}
	other := geo.GetSide(face, 1)
	if !other.(blockGeometry).HasSameTypeId(s.self) {
		if world, err := s.position.GetWorld(); err == nil {
			world.UseBreakOn(s.position.AsVector3())
		}
	}
}

// Place is a port of SmallDripleaf::place.
func (s *SmallDripleaf) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	geo := blockReplace.(blockGeometry)
	above := geo.GetSide(math.Up, 1)
	if above.GetTypeId() != AIR || !s.canBeSupportedBy(geo.GetSide(math.Down, 1)) {
		return false
	}
	if player != nil {
		s.Facing = math.Opposite(player.GetHorizontalFacing())
	}

	topHalf := VanillaSmallDripleaf().(*SmallDripleaf)
	topHalf.SetFacing(s.Facing)
	topHalf.SetTop(true)
	tx.AddBlock(above.GetPosition(), topHalf)

	return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract's fertilizer-driven grow needs a Fertilizer item marker, a fresh per-call
// BlockTransaction, World.IsInWorld/GetBlock(Position), and StructureGrowEvent, none ported yet.
// Block's default OnInteract (return false) already matches this gap, so there's nothing to
// override here.

func (s *SmallDripleaf) GetAffectedBlocks() []Behavior {
	face := math.Up
	if s.Top {
		face = math.Down
	}
	other := s.self.(blockGeometry).GetSide(face, 1)
	if other.(blockGeometry).HasSameTypeId(s.self) {
		return []Behavior{s.self, other}
	}
	return s.Block.GetAffectedBlocks()
}

// GetDropsForCompatibleTool should return [s.AsItem()] for the bottom half — needs real Item
// construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (s *SmallDripleaf) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (s *SmallDripleaf) GetFlameEncouragement() int { return 15 }

func (s *SmallDripleaf) GetFlammability() int { return 100 }

func (s *SmallDripleaf) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (s *SmallDripleaf) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }
