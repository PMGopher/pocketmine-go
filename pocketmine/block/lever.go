package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// Lever is a port of pocketmine\block\Lever.
type Lever struct {
	Flowable

	Facing    blockutils.LeverFacing
	Activated bool
}

func NewLever(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Lever {
	l := &Lever{
		Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
		Facing:   blockutils.LeverFacingUpAxisX,
	}
	l.Init(l)
	return l
}

func (l *Lever) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *Lever) DescribeBlockOnlyState(w runtime.DataDescriber) {
	facing := int(l.Facing)
	w.BoundedIntAuto(int(blockutils.LeverFacingUpAxisX), int(blockutils.LeverFacingWest), &facing)
	l.Facing = blockutils.LeverFacing(facing)
	w.Bool(&l.Activated)
}

func (l *Lever) GetFacing() blockutils.LeverFacing { return l.Facing }

func (l *Lever) SetFacing(facing blockutils.LeverFacing) { l.Facing = facing }

func (l *Lever) IsActivated() bool { return l.Activated }

func (l *Lever) SetActivated(activated bool) { l.Activated = activated }

func (l *Lever) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if !l.canBeSupportedAt(blockReplace, math.Opposite(face)) {
		return false
	}

	selectUpDownPos := func(x, z blockutils.LeverFacing) blockutils.LeverFacing {
		if player != nil && math.FacingAxis(player.GetHorizontalFacing()) == math.AxisX {
			return x
		}
		return z
	}

	switch face {
	case math.Down:
		l.Facing = selectUpDownPos(blockutils.LeverFacingDownAxisX, blockutils.LeverFacingDownAxisZ)
	case math.Up:
		l.Facing = selectUpDownPos(blockutils.LeverFacingUpAxisX, blockutils.LeverFacingUpAxisZ)
	case math.North:
		l.Facing = blockutils.LeverFacingNorth
	case math.South:
		l.Facing = blockutils.LeverFacingSouth
	case math.West:
		l.Facing = blockutils.LeverFacingWest
	case math.East:
		l.Facing = blockutils.LeverFacingEast
	default:
		panic("Bad facing value")
	}

	return l.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (l *Lever) OnNearbyBlockChange() {
	if !l.canBeSupportedAt(l.self, math.Opposite(l.Facing.GetFacing())) {
		if world, err := l.position.GetWorld(); err == nil {
			world.UseBreakOn(l.position.AsVector3())
		}
	}
}

func (l *Lever) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	l.Activated = !l.Activated
	world, err := l.position.GetWorld()
	if err != nil {
		panic(err)
	}
	if err := world.SetBlock(l.position, l.self); err != nil {
		panic(err)
	}
	var s sound.Sound
	if l.Activated {
		s = sound.RedstonePowerOnSound{}
	} else {
		s = sound.RedstonePowerOffSound{}
	}
	world.AddSound(l.position.AsVector3().Add(0.5, 0.5, 0.5), s)
	return true
}

func (l *Lever) canBeSupportedAt(blk Behavior, face math.Facing) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(face).HasCenterSupport()
}
