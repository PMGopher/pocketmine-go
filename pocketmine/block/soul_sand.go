package block

import "pocketmine-go/pocketmine/math"

// SoulSand is a port of pocketmine\block\SoulSand.
type SoulSand struct {
	Opaque
}

func NewSoulSand(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *SoulSand {
	s := &SoulSand{Opaque{NewBlock(idInfo, name, typeInfo)}}
	s.Init(s)
	return s
}

func (s *SoulSand) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *SoulSand) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 1.0/8)}
}
