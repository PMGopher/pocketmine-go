package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// FenceGate is a port of pocketmine\block\FenceGate.
type FenceGate struct {
	Transparent
	HorizontalFacingComponent
	WoodTypeComponent

	Open   bool
	InWall bool
}

func NewFenceGate(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, woodType blockutils.WoodType) *FenceGate {
	f := &FenceGate{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
		WoodTypeComponent:         NewWoodTypeComponent(woodType),
	}
	f.Init(f)
	return f
}

func (f *FenceGate) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FenceGate) DescribeBlockOnlyState(w runtime.DataDescriber) {
	f.DescribeHorizontalFacing(w)
	w.Bool(&f.Open)
	w.Bool(&f.InWall)
}

func (f *FenceGate) IsOpen() bool { return f.Open }

func (f *FenceGate) SetOpen(open bool) { f.Open = open }

func (f *FenceGate) IsInWall() bool { return f.InWall }

func (f *FenceGate) SetInWall(inWall bool) { f.InWall = inWall }

func (f *FenceGate) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	if f.Open {
		return nil
	}
	bb := math.OneAABB().ExtendedCopy(math.Up, 0.5).SquashedCopy(math.FacingAxis(f.Facing), 6.0/16)
	return []math.AxisAlignedBB{bb}
}

func (f *FenceGate) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (f *FenceGate) checkInWall() bool {
	left := f.GetSide(math.RotateY(f.Facing, false), 1)
	right := f.GetSide(math.RotateY(f.Facing, true), 1)
	_, leftIsWall := left.(*Wall)
	_, rightIsWall := right.(*Wall)
	return leftIsWall || rightIsWall
}

func (f *FenceGate) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		f.Facing = player.GetHorizontalFacing()
	}
	f.InWall = f.checkInWall()
	return f.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (f *FenceGate) OnNearbyBlockChange() {
	inWall := f.checkInWall()
	if inWall != f.InWall {
		f.InWall = inWall
		if world, err := f.position.GetWorld(); err == nil {
			if err := world.SetBlock(f.position, f.self); err != nil {
				panic(err)
			}
		}
	}
}

func (f *FenceGate) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	f.Open = !f.Open
	if f.Open && player != nil {
		playerFacing := player.GetHorizontalFacing()
		if playerFacing == math.Opposite(f.Facing) {
			f.Facing = playerFacing
		}
	}

	world, err := f.position.GetWorld()
	if err != nil {
		panic(err)
	}
	if err := world.SetBlock(f.position, f.self); err != nil {
		panic(err)
	}
	world.AddSound(f.position.AsVector3(), sound.DoorSound{})
	return true
}

func (f *FenceGate) GetFuelTime() int {
	if f.WoodType.IsFlammable() {
		return 300
	}
	return 0
}

func (f *FenceGate) GetFlameEncouragement() int {
	if f.WoodType.IsFlammable() {
		return 5
	}
	return 0
}

func (f *FenceGate) GetFlammability() int {
	if f.WoodType.IsFlammable() {
		return 20
	}
	return 0
}
