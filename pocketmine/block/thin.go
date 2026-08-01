package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// Thin is a port of pocketmine\block\Thin.
//
// Thin blocks behave like glass panes. They connect to full-cube blocks horizontally adjacent to
// them if possible.
type Thin struct {
	Transparent

	Connections map[math.Facing]bool
}

func NewThin(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Thin {
	t := &Thin{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Connections: map[math.Facing]bool{}}
	t.Init(t)
	return t
}

// Clone can't use the usual one-line pattern: Connections is a map, a reference type - see
// Vine.Clone's doc comment for the same reasoning.
func (t *Thin) Clone() Behavior {
	c := *t
	c.Connections = make(map[math.Facing]bool, len(t.Connections))
	for k, v := range t.Connections {
		c.Connections[k] = v
	}
	c.rebind(&c)
	return &c
}

func (t *Thin) ReadStateFromWorld() Behavior {
	t.Block.ReadStateFromWorld()

	t.collisionBoxes = nil
	t.haveCollisionBoxes = false

	connections := map[math.Facing]bool{}
	for _, facing := range math.HorizontalFacing {
		side := t.GetSide(facing, 1)
		_, isThin := side.(*Thin)
		_, isWall := side.(*Wall)
		if isThin || isWall || side.GetSupportType(math.Opposite(facing)) == blockutils.SupportTypeFull {
			connections[facing] = true
		}
	}
	t.Connections = connections

	return t.self
}

func (t *Thin) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	const inset = 7.0 / 16

	var bbs []math.AxisAlignedBB

	if t.Connections[math.West] || t.Connections[math.East] {
		bb := math.OneAABB().SquashedCopy(math.AxisZ, inset)
		if !t.Connections[math.West] {
			bb = bb.TrimmedCopy(math.West, inset)
		} else if !t.Connections[math.East] {
			bb = bb.TrimmedCopy(math.East, inset)
		}
		bbs = append(bbs, bb)
	}

	if t.Connections[math.North] || t.Connections[math.South] {
		bb := math.OneAABB().SquashedCopy(math.AxisX, inset)
		if !t.Connections[math.North] {
			bb = bb.TrimmedCopy(math.North, inset)
		} else if !t.Connections[math.South] {
			bb = bb.TrimmedCopy(math.South, inset)
		}
		bbs = append(bbs, bb)
	}

	if len(bbs) == 0 {
		return []math.AxisAlignedBB{math.OneAABB().ContractedCopy(inset, 0, inset)}
	}

	return bbs
}

func (t *Thin) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
