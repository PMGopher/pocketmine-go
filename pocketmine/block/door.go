package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// Door is a port of pocketmine\block\Door.
type Door struct {
	Transparent
	HorizontalFacingComponent

	Top        bool
	HingeRight bool
	Open       bool
}

func NewDoor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Door {
	d := &Door{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	d.Init(d)
	return d
}

func (d *Door) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *Door) DescribeBlockOnlyState(w runtime.DataDescriber) {
	d.DescribeHorizontalFacing(w)
	w.Bool(&d.Top)
	w.Bool(&d.HingeRight)
	w.Bool(&d.Open)
}

func (d *Door) ReadStateFromWorld() Behavior {
	d.Block.ReadStateFromWorld()

	d.collisionBoxes = nil
	d.haveCollisionBoxes = false

	otherSide := math.Up
	if d.Top {
		otherSide = math.Down
	}
	if other, ok := d.GetSide(otherSide, 1).(*Door); ok && other.HasSameTypeId(d.self) {
		if d.Top {
			d.Facing = other.Facing
			d.Open = other.Open
		} else {
			d.HingeRight = other.HingeRight
		}
	}

	return d.self
}

func (d *Door) IsTop() bool { return d.Top }

func (d *Door) SetTop(top bool) { d.Top = top }

func (d *Door) IsHingeRight() bool { return d.HingeRight }

func (d *Door) SetHingeRight(hingeRight bool) { d.HingeRight = hingeRight }

func (d *Door) IsOpen() bool { return d.Open }

func (d *Door) SetOpen(open bool) { d.Open = open }

func (d *Door) IsSolid() bool { return false }

// RecalculateCollisionBoxes: TODO (from the PHP original) doors are 0.1825 blocks thick, instead
// of 0.1875 like Java Edition (https://bugs.mojang.com/browse/MCPE-19214).
func (d *Door) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	trimFace := d.Facing
	if d.Open {
		trimFace = math.RotateY(d.Facing, !d.HingeRight)
	}
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(trimFace, 327.0/400)}
}

func (d *Door) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (d *Door) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Down).HasEdgeSupport()
}

func (d *Door) OnNearbyBlockChange() {
	_, downIsDoor := d.GetSide(math.Down, 1).(*Door)
	if !d.canBeSupportedAt(d.self) && !downIsDoor {
		if world, err := d.position.GetWorld(); err == nil {
			world.UseBreakOn(d.position.AsVector3()) // this will delete both halves if they exist
		}
	}
}

func (d *Door) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if face != math.Up {
		return false
	}

	blockUp := d.GetSide(math.Up, 1)
	if !blockUp.CanBeReplaced() || !d.canBeSupportedAt(blockReplace) {
		return false
	}

	if player != nil {
		d.Facing = player.GetHorizontalFacing()
	}

	next := d.GetSide(math.RotateY(d.Facing, false), 1)
	next2 := d.GetSide(math.RotateY(d.Facing, true), 1)

	nextIsDoor := next.(blockGeometry).HasSameTypeId(d.self)
	if nextIsDoor || (!next2.IsTransparent() && next.IsTransparent()) {
		d.HingeRight = true
	}

	topHalf := d.Clone().(*Door)
	topHalf.Top = true

	tx.AddBlock(blockReplace.GetPosition(), d.self)
	tx.AddBlock(blockUp.GetPosition(), topHalf)
	return true
}

func (d *Door) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	d.Open = !d.Open

	otherSide := math.Up
	if d.Top {
		otherSide = math.Down
	}

	world, err := d.position.GetWorld()
	if err != nil {
		panic(err)
	}
	if other, ok := d.GetSide(otherSide, 1).(*Door); ok && other.HasSameTypeId(d.self) {
		other.Open = d.Open
		if err := world.SetBlock(other.GetPosition(), other); err != nil {
			panic(err)
		}
	}

	if err := world.SetBlock(d.position, d.self); err != nil {
		panic(err)
	}
	world.AddSound(d.position.AsVector3(), sound.DoorSound{})

	return true
}

func (d *Door) GetDrops(item Item) []Item {
	if !d.Top {
		return d.Block.GetDrops(item)
	}
	return nil
}

func (d *Door) GetAffectedBlocks() []Behavior {
	otherSide := math.Up
	if d.Top {
		otherSide = math.Down
	}
	other := d.GetSide(otherSide, 1)
	if other.(blockGeometry).HasSameTypeId(d.self) {
		return []Behavior{d.self, other}
	}
	return d.Block.GetAffectedBlocks()
}
