package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// NetherPortal is a port of pocketmine\block\NetherPortal.
type NetherPortal struct {
	Transparent

	Axis math.Axis
}

func NewNetherPortal(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *NetherPortal {
	n := &NetherPortal{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Axis: math.AxisX}
	n.Init(n)
	return n
}

func (n *NetherPortal) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *NetherPortal) DescribeBlockOnlyState(w runtime.DataDescriber) { w.HorizontalAxis(&n.Axis) }

func (n *NetherPortal) GetAxis() math.Axis { return n.Axis }

// SetAxis panics for a non-horizontal axis, mirroring the PHP original's InvalidArgumentException
// (a programmer error at the call site).
func (n *NetherPortal) SetAxis(axis math.Axis) {
	if axis != math.AxisX && axis != math.AxisZ {
		panic("Invalid axis")
	}
	n.Axis = axis
}

func (n *NetherPortal) GetLightLevel() int { return 11 }

func (n *NetherPortal) IsSolid() bool { return false }

func (n *NetherPortal) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (n *NetherPortal) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (n *NetherPortal) GetDrops(item Item) []Item { return nil }

// OnEntityInside is a TODO in the PHP original too - nether portal teleportation logic isn't
// implemented upstream either.
func (n *NetherPortal) OnEntityInside(entity Entity) bool { return true }
