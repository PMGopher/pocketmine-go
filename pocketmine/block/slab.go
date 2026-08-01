package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Slab is a port of pocketmine\block\Slab.
type Slab struct {
	Transparent

	SlabTypeValue blockutils.SlabType
}

// NewSlab mirrors the PHP constructor, which appends " Slab" to the given name.
func NewSlab(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Slab {
	s := &Slab{Transparent: Transparent{NewBlock(idInfo, name+" Slab", typeInfo)}, SlabTypeValue: blockutils.SlabTypeBottom}
	s.Init(s)
	return s
}

func (s *Slab) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Slab) DescribeBlockOnlyState(w runtime.DataDescriber) {
	slabType := int(s.SlabTypeValue)
	w.BoundedIntAuto(int(blockutils.SlabTypeBottom), int(blockutils.SlabTypeDouble), &slabType)
	s.SlabTypeValue = blockutils.SlabType(slabType)
}

func (s *Slab) IsTransparent() bool { return s.SlabTypeValue != blockutils.SlabTypeDouble }

func (s *Slab) GetSlabType() blockutils.SlabType { return s.SlabTypeValue }

func (s *Slab) SetSlabType(slabType blockutils.SlabType) { s.SlabTypeValue = slabType }

func (s *Slab) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	if s.Transparent.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock) {
		return true
	}

	if replace, ok := blockReplace.(*Slab); ok && replace.SlabTypeValue != blockutils.SlabTypeDouble && replace.HasSameTypeId(s.self) {
		if replace.SlabTypeValue == blockutils.SlabTypeTop {
			return clickVector.Y <= 0.5 || (!isClickedBlock && face == math.Up)
		}
		return clickVector.Y >= 0.5 || (!isClickedBlock && face == math.Down)
	}

	return false
}

func (s *Slab) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if replace, ok := blockReplace.(*Slab); ok && replace.SlabTypeValue != blockutils.SlabTypeDouble && replace.HasSameTypeId(s.self) &&
		((replace.SlabTypeValue == blockutils.SlabTypeTop && (clickVector.Y <= 0.5 || face == math.Up)) ||
			(replace.SlabTypeValue == blockutils.SlabTypeBottom && (clickVector.Y >= 0.5 || face == math.Down))) {
		// Clicked in empty half of existing slab
		s.SlabTypeValue = blockutils.SlabTypeDouble
	} else if (face != math.Up && clickVector.Y > 0.5) || face == math.Down {
		s.SlabTypeValue = blockutils.SlabTypeTop
	} else {
		s.SlabTypeValue = blockutils.SlabTypeBottom
	}

	return s.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (s *Slab) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	if s.SlabTypeValue == blockutils.SlabTypeDouble {
		return []math.AxisAlignedBB{math.OneAABB()}
	}
	trimFace := math.Up
	if s.SlabTypeValue == blockutils.SlabTypeTop {
		trimFace = math.Down
	}
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(trimFace, 0.5)}
}

func (s *Slab) GetSupportType(facing math.Facing) blockutils.SupportType {
	if s.SlabTypeValue == blockutils.SlabTypeDouble {
		return blockutils.SupportTypeFull
	}
	if (facing == math.Up && s.SlabTypeValue == blockutils.SlabTypeTop) || (facing == math.Down && s.SlabTypeValue == blockutils.SlabTypeBottom) {
		return blockutils.SupportTypeFull
	}
	return blockutils.SupportTypeNone
}

// GetDropsForCompatibleTool should return [s.AsItem().SetCount(...)] — needs real Item
// construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (s *Slab) GetDropsForCompatibleTool(item Item) []Item { return nil }
