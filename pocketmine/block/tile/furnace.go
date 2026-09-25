package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	FurnaceTagBurnTime = "BurnTime"
	FurnaceTagCookTime = "CookTime"
	FurnaceTagMaxTime  = "MaxTime"
)

// FurnaceType is a port of pocketmine\crafting\FurnaceType (crafting.FurnaceType is an alias):
// it lives here because the furnace tiles need it and the crafting package imports this one.
// getCookSound isn't ported (the furnace sounds aren't).
type FurnaceType int

const (
	FurnaceTypeFurnace FurnaceType = iota
	FurnaceTypeBlastFurnace
	FurnaceTypeSmoker
	FurnaceTypeCampfire
	FurnaceTypeSoulCampfire
)

// AllFurnaceTypes is FurnaceType::cases().
var AllFurnaceTypes = []FurnaceType{FurnaceTypeFurnace, FurnaceTypeBlastFurnace, FurnaceTypeSmoker, FurnaceTypeCampfire, FurnaceTypeSoulCampfire}

// GetCookDurationTicks is a port of FurnaceType::getCookDurationTicks.
func (t FurnaceType) GetCookDurationTicks() int {
	switch t {
	case FurnaceTypeBlastFurnace, FurnaceTypeSmoker:
		return 100
	case FurnaceTypeCampfire, FurnaceTypeSoulCampfire:
		return 600
	}
	return 200
}

// Furnace is a port of pocketmine\block\tile\Furnace, minus its inventory/Container half - see
// ContainerComponent's doc comment for why the inventory package can't be imported here.
//
// The burn/cook/max time STATE and its NBT round trip are fully real, including the PHP
// original's load-time consistency fixups (cookTime forced to 0 if there's no fuel left,
// maxFuelTime defaulting to the current remaining fuel time if it was never saved). checkFuel/
// onStartSmelting/onStopSmelting/onUpdate (the actual smelting simulation) all need the fuel/
// smelting/result inventory slots plus FurnaceRecipe/CraftingManager (crafting package, not
// ported), so none of that is ported.
type Furnace struct {
	SpawnableBase
	NameableComponent
	ContainerComponent

	RemainingFuelTime int
	CookTime          int
	MaxFuelTime       int
}

func NewFurnace(world World, pos math.Vector3) *Furnace {
	f := &Furnace{}
	f.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	f.Init(f)
	return f
}

func (f *Furnace) SaveID() string { return "Furnace" }

func (f *Furnace) GetDefaultName() string { return "Furnace" }

func (f *Furnace) GetName() string { return f.NameableComponent.GetName(f) }

// ReadSaveData is a port of Furnace::readSaveData.
func (f *Furnace) ReadSaveData(tag *nbt.CompoundTag) error {
	f.RemainingFuelTime = max(0, int(tag.GetShortOr(FurnaceTagBurnTime, nbt.ShortTag(f.RemainingFuelTime))))

	f.CookTime = int(tag.GetShortOr(FurnaceTagCookTime, nbt.ShortTag(f.CookTime)))
	if f.RemainingFuelTime == 0 {
		f.CookTime = 0
	}

	f.MaxFuelTime = int(tag.GetShortOr(FurnaceTagMaxTime, nbt.ShortTag(f.MaxFuelTime)))
	if f.MaxFuelTime == 0 {
		f.MaxFuelTime = f.RemainingFuelTime
	}

	f.LoadName(tag)
	return nil
}

func (f *Furnace) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetShort(FurnaceTagBurnTime, nbt.ShortTag(f.RemainingFuelTime))
	tag.SetShort(FurnaceTagCookTime, nbt.ShortTag(f.CookTime))
	tag.SetShort(FurnaceTagMaxTime, nbt.ShortTag(f.MaxFuelTime))
	f.SaveName(tag)
}

// CopyDataFromItem must be defined here rather than relying on promotion - see
// NameableComponent.ApplyItemCustomName's doc comment for why.
func (f *Furnace) CopyDataFromItem(item Item) {
	f.TileBase.CopyDataFromItem(item)
	f.ApplyItemCustomName(item)
}

// NormalFurnace is a port of pocketmine\block\tile\NormalFurnace.
type NormalFurnace struct{ Furnace }

func NewNormalFurnace(world World, pos math.Vector3) *NormalFurnace {
	n := &NormalFurnace{}
	n.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	n.Init(n)
	return n
}

func (n *NormalFurnace) GetFurnaceType() FurnaceType { return FurnaceTypeFurnace }

// BlastFurnace is a port of pocketmine\block\tile\BlastFurnace.
type BlastFurnace struct{ Furnace }

func NewBlastFurnace(world World, pos math.Vector3) *BlastFurnace {
	b := &BlastFurnace{}
	b.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	b.Init(b)
	return b
}

func (b *BlastFurnace) SaveID() string { return "BlastFurnace" }

func (b *BlastFurnace) GetFurnaceType() FurnaceType { return FurnaceTypeBlastFurnace }

// Smoker is a port of pocketmine\block\tile\Smoker.
type Smoker struct{ Furnace }

func NewSmoker(world World, pos math.Vector3) *Smoker {
	s := &Smoker{}
	s.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	s.Init(s)
	return s
}

func (s *Smoker) SaveID() string { return "Smoker" }

func (s *Smoker) GetFurnaceType() FurnaceType { return FurnaceTypeSmoker }
