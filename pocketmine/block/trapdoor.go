package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// Trapdoor is a port of pocketmine\block\Trapdoor.
type Trapdoor struct {
	Transparent
	HorizontalFacingComponent

	Open bool
	Top  bool
}

func NewTrapdoor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Trapdoor {
	t := &Trapdoor{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	t.Init(t)
	return t
}

func (t *Trapdoor) Clone() Behavior {
	c := *t
	c.rebind(&c)
	return &c
}

func (t *Trapdoor) DescribeBlockOnlyState(w runtime.DataDescriber) {
	t.DescribeHorizontalFacing(w)
	w.Bool(&t.Top)
	w.Bool(&t.Open)
}

func (t *Trapdoor) IsOpen() bool { return t.Open }

func (t *Trapdoor) SetOpen(open bool) { t.Open = open }

func (t *Trapdoor) IsTop() bool { return t.Top }

func (t *Trapdoor) SetTop(top bool) { t.Top = top }

// RecalculateCollisionBoxes: TODO (from the PHP original) - like doors, these are slightly too
// thin in Bedrock (0.1825 instead of 0.1875).
func (t *Trapdoor) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	trimFace := math.Up
	if t.Open {
		trimFace = t.Facing
	} else if t.Top {
		trimFace = math.Down
	}
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(trimFace, 1-0.1825)}
}

func (t *Trapdoor) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (t *Trapdoor) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		t.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	if (clickVector.Y > 0.5 && face != math.Up) || face == math.Down {
		t.Top = true
	}
	return t.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (t *Trapdoor) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	t.Open = !t.Open
	world, err := t.position.GetWorld()
	if err != nil {
		panic(err)
	}
	if err := world.SetBlock(t.position, t.self); err != nil {
		panic(err)
	}
	world.AddSound(t.position.AsVector3(), sound.DoorSound{})
	return true
}
