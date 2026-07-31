package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Dirt is a port of pocketmine\block\Dirt.
type Dirt struct {
	Opaque

	DirtTypeValue blockutils.DirtType
}

func NewDirt(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Dirt {
	d := &Dirt{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}, DirtTypeValue: blockutils.DirtTypeNormal}
	d.Init(d)
	return d
}

func (d *Dirt) Clone() Behavior {
	c := *d
	c.rebind(&c)
	return &c
}

func (d *Dirt) DescribeBlockItemState(w runtime.DataDescriber) {
	dirtType := int(d.DirtTypeValue)
	w.BoundedIntAuto(int(blockutils.DirtTypeNormal), int(blockutils.DirtTypeRooted), &dirtType)
	d.DirtTypeValue = blockutils.DirtType(dirtType)
}

func (d *Dirt) GetDirtType() blockutils.DirtType { return d.DirtTypeValue }

func (d *Dirt) SetDirtType(dirtType blockutils.DirtType) { d.DirtTypeValue = dirtType }

// OnInteract should turn into Farmland/Dirt (hoe), spawn hanging roots (fertilizer on rooted
// dirt), or turn into Mud (water potion) — needs Hoe/Fertilizer/Potion item types and the block
// registry (VanillaBlocks), none ported yet, so this is a no-op for now (see
// Block.GetDropsForCompatibleTool's doc comment for the same category of gap).
func (d *Dirt) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return false
}
