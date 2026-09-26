package block

import (
	"fmt"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

const (
	FillableCauldronMinFillLevel = 1
	FillableCauldronMaxFillLevel = 6
)

// fillableCauldron is the part of FillableCauldron's subclasses Cauldron.fill uses.
type fillableCauldron interface {
	Behavior
	SetFillLevel(fillLevel int)
	GetFillSound() sound.Sound
}

// fillableCauldronHooks is FillableCauldron's abstract methods.
type fillableCauldronHooks interface {
	GetFillSound() sound.Sound
	GetEmptySound() sound.Sound
}

// FillableCauldron is a port of pocketmine\block\FillableCauldron.
type FillableCauldron struct {
	Transparent

	FillLevel int
}

func newFillableCauldron(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) FillableCauldron {
	return FillableCauldron{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, FillLevel: FillableCauldronMinFillLevel}
}

func (f *FillableCauldron) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.BoundedIntAuto(FillableCauldronMinFillLevel, FillableCauldronMaxFillLevel, &f.FillLevel)
}

func (f *FillableCauldron) GetFillLevel() int { return f.FillLevel }

// SetFillLevel is a port of FillableCauldron::setFillLevel.
func (f *FillableCauldron) SetFillLevel(fillLevel int) {
	if fillLevel < FillableCauldronMinFillLevel || fillLevel > FillableCauldronMaxFillLevel {
		panic(fmt.Sprintf("Fill level must be in range %d ... %d", FillableCauldronMinFillLevel, FillableCauldronMaxFillLevel))
	}
	f.FillLevel = fillLevel
}

func (f *FillableCauldron) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return cauldronCollisionBoxes()
}

func (f *FillableCauldron) GetSupportType(facing math.Facing) blockutils.SupportType {
	if facing == math.Up {
		return blockutils.SupportTypeEdge
	}
	return blockutils.SupportTypeNone
}

// withFillLevel is a port of FillableCauldron::withFillLevel.
func (f *FillableCauldron) withFillLevel(fillLevel int) Behavior {
	if fillLevel == 0 {
		return VanillaBlock("cauldron")
	}
	f.SetFillLevel(min(FillableCauldronMaxFillLevel, fillLevel))
	return f.self
}

// addFillLevels is a port of FillableCauldron::addFillLevels.
func (f *FillableCauldron) addFillLevels(amount int, usedItem Item, returnedItem Item, returnedItems *[]Item) {
	if f.FillLevel >= FillableCauldronMaxFillLevel {
		return
	}
	world, err := f.position.GetWorld()
	if err != nil {
		return
	}
	_ = world.SetBlock(f.position, f.withFillLevel(f.FillLevel+amount))
	world.AddSound(f.position.Add(0.5, 0.5, 0.5), f.self.(fillableCauldronHooks).GetFillSound())

	usedItem.Pop()
	appendItem(returnedItems, returnedItem)
}

// removeFillLevels is a port of FillableCauldron::removeFillLevels.
func (f *FillableCauldron) removeFillLevels(amount int, usedItem Item, returnedItem Item, returnedItems *[]Item) {
	if f.FillLevel < amount {
		return
	}
	world, err := f.position.GetWorld()
	if err != nil {
		return
	}
	_ = world.SetBlock(f.position, f.withFillLevel(f.FillLevel-amount))
	world.AddSound(f.position.Add(0.5, 0.5, 0.5), f.self.(fillableCauldronHooks).GetEmptySound())

	usedItem.Pop()
	appendItem(returnedItems, returnedItem)
}

// mix is a port of FillableCauldron::mix.
func (f *FillableCauldron) mix(usedItem Item, returnedItem Item, returnedItems *[]Item) {
	world, err := f.position.GetWorld()
	if err != nil {
		return
	}
	_ = world.SetBlock(f.position, VanillaBlock("cauldron"))
	//TODO: sounds and particles

	usedItem.Pop()
	appendItem(returnedItems, returnedItem)
}

// AsItem is a port of FillableCauldron::asItem: every cauldron drops as an empty cauldron.
func (f *FillableCauldron) AsItem() (Item, error) {
	return VanillaBlock("cauldron").(interface{ AsItem() (Item, error) }).AsItem()
}
