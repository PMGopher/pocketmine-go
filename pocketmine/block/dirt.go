package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
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

// OnInteract is a port of Dirt::onInteract: a hoe tills it (coarse and rooted dirt become dirt,
// rooted dirt drops hanging roots), bone meal grows hanging roots under rooted dirt, and a water
// potion turns it into mud.
func (d *Dirt) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	world, err := d.position.GetWorld()
	if err != nil {
		return false
	}
	if face != math.Down && isHoe(item) {
		if d.self.(blockGeometry).GetSide(math.Up, 1).GetTypeId() != AIR {
			return true
		}
		applyDamage(item, 1)
		newBlock := VanillaDirt()
		if d.DirtTypeValue == blockutils.DirtTypeNormal {
			newBlock = VanillaBlock("farmland")
		}
		center := d.position.Add(0.5, 0.5, 0.5)
		world.AddSound(center, sound.ItemUseOnBlockSound{BlockStateID: newBlock.GetStateId()})
		_ = world.SetBlock(d.position, newBlock)
		if d.DirtTypeValue == blockutils.DirtTypeRooted {
			if roots := asItemOrNil(VanillaBlock("hanging_roots")); roots != nil {
				if dropper, ok := world.(blockItemDropper); ok {
					dropper.DropBlockItem(center, roots)
				}
			}
		}
		return true
	} else if d.DirtTypeValue == blockutils.DirtTypeRooted && isFertilizer(item) {
		down := d.self.(blockGeometry).GetSide(math.Down, 1)
		if down.GetTypeId() != AIR {
			return true
		}
		item.Pop()
		_ = world.SetBlock(down.GetPosition(), VanillaBlock("hanging_roots"))
		//TODO: bonemeal particles, growth sounds
	} else if potion, ok := item.(waterPotionChecker); ok && potion.IsWaterPotion() {
		item.Pop()
		_ = world.SetBlock(d.position, VanillaBlock("mud"))
		world.AddSound(d.position.Vector3, sound.NewWaterSplashSound(0.5))
		return true
	}
	return false
}
