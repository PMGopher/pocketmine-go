package block

import "pocketmine-go/pocketmine/math"

// StructureVoid is a port of pocketmine\block\StructureVoid.
type StructureVoid struct {
	Transparent
}

func NewStructureVoid(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *StructureVoid {
	s := &StructureVoid{Transparent{NewBlock(idInfo, name, typeInfo)}}
	s.Init(s)
	return s
}

func (s *StructureVoid) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *StructureVoid) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (s *StructureVoid) IsSolid() bool { return false }
