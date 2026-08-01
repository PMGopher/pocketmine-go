package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// Fence is a port of pocketmine\block\Fence.
type Fence struct {
	Transparent

	Connections map[math.Facing]bool
}

func NewFence(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Fence {
	f := &Fence{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Connections: map[math.Facing]bool{}}
	f.Init(f)
	return f
}

// Clone can't use the usual one-line pattern: Connections is a map, a reference type - see
// Vine.Clone's doc comment for the same reasoning.
func (f *Fence) Clone() Behavior {
	c := *f
	c.Connections = make(map[math.Facing]bool, len(f.Connections))
	for k, v := range f.Connections {
		c.Connections[k] = v
	}
	c.rebind(&c)
	return &c
}

func (f *Fence) GetThickness() float64 { return 0.25 }

func (f *Fence) ReadStateFromWorld() Behavior {
	f.Block.ReadStateFromWorld()

	f.collisionBoxes = nil
	f.haveCollisionBoxes = false

	connections := map[math.Facing]bool{}
	for _, facing := range math.HorizontalFacing {
		side := f.GetSide(facing, 1)
		_, isFence := side.(*Fence)
		_, isFenceGate := side.(*FenceGate)
		if isFence || isFenceGate || side.GetSupportType(math.Opposite(facing)) == blockutils.SupportTypeFull {
			connections[facing] = true
		}
	}
	f.Connections = connections

	return f.self
}

func (f *Fence) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	inset := 0.5 - f.GetThickness()/2

	var bbs []math.AxisAlignedBB

	connectWest := f.Connections[math.West]
	connectEast := f.Connections[math.East]
	if connectWest || connectEast {
		bb := math.OneAABB().SquashedCopy(math.AxisZ, inset).ExtendedCopy(math.Up, 0.5)
		if !connectWest {
			bb = bb.TrimmedCopy(math.West, inset)
		}
		if !connectEast {
			bb = bb.TrimmedCopy(math.East, inset)
		}
		bbs = append(bbs, bb)
	}

	connectNorth := f.Connections[math.North]
	connectSouth := f.Connections[math.South]
	if connectNorth || connectSouth {
		bb := math.OneAABB().SquashedCopy(math.AxisX, inset).ExtendedCopy(math.Up, 0.5)
		if !connectNorth {
			bb = bb.TrimmedCopy(math.North, inset)
		}
		if !connectSouth {
			bb = bb.TrimmedCopy(math.South, inset)
		}
		bbs = append(bbs, bb)
	}

	if len(bbs) == 0 {
		return []math.AxisAlignedBB{math.OneAABB().ExtendedCopy(math.Up, 0.5).ContractedCopy(inset, 0, inset)}
	}

	return bbs
}

func (f *Fence) GetSupportType(facing math.Facing) blockutils.SupportType {
	if math.FacingAxis(facing) == math.AxisY {
		return blockutils.SupportTypeCenter
	}
	return blockutils.SupportTypeNone
}
