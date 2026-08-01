package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Wall is a port of pocketmine\block\Wall.
type Wall struct {
	Transparent

	Connections map[math.Facing]blockutils.WallConnectionType
	Post        bool
}

func NewWall(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Wall {
	w := &Wall{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Connections: map[math.Facing]blockutils.WallConnectionType{}}
	w.Init(w)
	return w
}

// Clone can't use the usual one-line pattern: Connections is a map, a reference type - see
// Vine.Clone's doc comment for the same reasoning.
func (w *Wall) Clone() Behavior {
	c := *w
	c.Connections = make(map[math.Facing]blockutils.WallConnectionType, len(w.Connections))
	for k, v := range w.Connections {
		c.Connections[k] = v
	}
	c.rebind(&c)
	return &c
}

// DescribeBlockOnlyState is a port of RuntimeDataWriter/Reader/SizeCalculator::wallConnections:
// each of the 4 horizontal connection states (none/short/tall) is a base-3 digit, packed into one
// BoundedIntAuto(0, 3^4-1) value - rather than adding a new DataDescriber method for this one
// block, the existing BoundedIntAuto primitive already covers it exactly like Lever's enum
// encoding does.
func (w *Wall) DescribeBlockOnlyState(d runtime.DataDescriber) {
	packed := 0
	mult := 1
	for _, facing := range math.HorizontalFacing {
		digit := 0
		if t, ok := w.Connections[facing]; ok {
			digit = int(t) + 1
		}
		packed += digit * mult
		mult *= 3
	}

	d.BoundedIntAuto(0, 80, &packed)

	newConnections := make(map[math.Facing]blockutils.WallConnectionType, 4)
	mult = 1
	for _, facing := range math.HorizontalFacing {
		digit := (packed / mult) % 3
		if digit != 0 {
			newConnections[facing] = blockutils.WallConnectionType(digit - 1)
		}
		mult *= 3
	}
	w.Connections = newConnections

	d.Bool(&w.Post)
}

func (w *Wall) GetConnections() map[math.Facing]blockutils.WallConnectionType { return w.Connections }

func (w *Wall) GetConnection(face math.Facing) (blockutils.WallConnectionType, bool) {
	t, ok := w.Connections[face]
	return t, ok
}

func (w *Wall) SetConnections(connections map[math.Facing]blockutils.WallConnectionType) {
	w.Connections = connections
}

// SetConnection panics if face isn't horizontal, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (w *Wall) SetConnection(face math.Facing, connType blockutils.WallConnectionType, present bool) {
	validateHorizontalFace(face)
	if present {
		w.Connections[face] = connType
	} else {
		delete(w.Connections, face)
	}
}

func (w *Wall) IsPost() bool { return w.Post }

func (w *Wall) SetPost(post bool) { w.Post = post }

func (w *Wall) OnNearbyBlockChange() {
	if w.recalculateConnections() {
		if world, err := w.position.GetWorld(); err == nil {
			if err := world.SetBlock(w.position, w.self); err != nil {
				panic(err)
			}
		}
	}
}

// recalculateConnections doesn't yet implement tall/short connection selection - TODO (from the
// PHP original): right now only short is supported, matching pre-1.16 behavior.
func (w *Wall) recalculateConnections() bool {
	changed := 0

	for _, facing := range math.HorizontalFacing {
		side := w.GetSide(facing, 1)
		geo := side.(blockGeometry)
		_, isThin := side.(*Thin)
		_, isFenceGate := side.(*FenceGate)
		connects := geo.HasSameTypeId(w.self) || isFenceGate || isThin ||
			side.GetSupportType(math.Opposite(facing)) == blockutils.SupportTypeFull

		if connects {
			if _, ok := w.Connections[facing]; !ok {
				w.Connections[facing] = blockutils.WallConnectionTypeShort
				changed++
			}
		} else if _, ok := w.Connections[facing]; ok {
			delete(w.Connections, facing)
			changed++
		}
	}

	up := w.GetSide(math.Up, 1).GetTypeId() != AIR
	if up != w.Post {
		w.Post = up
		changed++
	}

	return changed > 0
}

func (w *Wall) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	_, north := w.Connections[math.North]
	_, south := w.Connections[math.South]
	_, west := w.Connections[math.West]
	_, east := w.Connections[math.East]

	inset := 0.25
	if !w.Post && ((north && south && !west && !east) || (!north && !south && west && east)) {
		inset = 0.3125
	}

	bb := math.OneAABB().ExtendedCopy(math.Up, 0.5)
	if !north {
		bb = bb.TrimmedCopy(math.North, inset)
	}
	if !south {
		bb = bb.TrimmedCopy(math.South, inset)
	}
	if !west {
		bb = bb.TrimmedCopy(math.West, inset)
	}
	if !east {
		bb = bb.TrimmedCopy(math.East, inset)
	}
	return []math.AxisAlignedBB{bb}
}

func (w *Wall) GetSupportType(facing math.Facing) blockutils.SupportType {
	if math.FacingAxis(facing) == math.AxisY {
		return blockutils.SupportTypeCenter
	}
	return blockutils.SupportTypeNone
}
