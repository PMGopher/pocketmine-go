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
// getCookSound lives in the block package (block.furnaceCookSound), which has the sounds.
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

// FurnaceOnUpdateFunc is Furnace::onUpdate (the smelting simulation: fuel, recipes, events and
// viewers' progress bars), set by block/inventory, which can import the packages it needs.
var FurnaceOnUpdateFunc func(f *Furnace) bool

// Furnace is a port of pocketmine\block\tile\Furnace (abstract in PHP: NormalFurnace,
// BlastFurnace and Smoker set the furnace type).
type Furnace struct {
	SpawnableBase
	NameableComponent
	ContainerComponent

	RemainingFuelTime int
	CookTime          int
	MaxFuelTime       int

	furnaceType FurnaceType
}

// GetFurnaceType is a port of Furnace::getFurnaceType.
func (f *Furnace) GetFurnaceType() FurnaceType { return f.furnaceType }

// GetInventory is a port of Furnace::getInventory (a FurnaceInventory).
func (f *Furnace) GetInventory() Inventory { return f.realInventory(f.self) }

// GetRealInventory is a port of Furnace::getRealInventory.
func (f *Furnace) GetRealInventory() Inventory { return f.realInventory(f.self) }

// CloseHook is Furnace::close's removal of the inventory's viewers.
func (f *Furnace) CloseHook() { f.removeAllViewers() }

// OnBlockDestroyedHook is ContainerTrait::onBlockDestroyedHook.
func (f *Furnace) OnBlockDestroyedHook() { f.dropContents(f.self) }

// OnUpdate is a port of Furnace::onUpdate: whether the furnace is still smelting.
func (f *Furnace) OnUpdate() bool {
	if f.closed || FurnaceOnUpdateFunc == nil {
		return false
	}
	return FurnaceOnUpdateFunc(f)
}

func NewFurnace(world World, pos math.Vector3) *Furnace {
	f := &Furnace{furnaceType: FurnaceTypeFurnace}
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
	f.loadItems(f.self, tag)

	if f.RemainingFuelTime > 0 {
		if world, ok := f.position.GetWorld(); ok {
			if scheduler, ok := world.(blockUpdateScheduler); ok {
				scheduler.ScheduleDelayedBlockUpdate(f.position.Vector3, 1)
			}
		}
	}
	return nil
}

// blockUpdateScheduler is World::scheduleDelayedBlockUpdate.
type blockUpdateScheduler interface {
	ScheduleDelayedBlockUpdate(pos math.Vector3, delay int)
}

func (f *Furnace) WriteSaveData(tag *nbt.CompoundTag) {
	tag.SetShort(FurnaceTagBurnTime, nbt.ShortTag(f.RemainingFuelTime))
	tag.SetShort(FurnaceTagCookTime, nbt.ShortTag(f.CookTime))
	tag.SetShort(FurnaceTagMaxTime, nbt.ShortTag(f.MaxFuelTime))
	f.SaveName(tag)
	f.saveItems(f.self, tag)
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
	n := &NormalFurnace{Furnace{furnaceType: FurnaceTypeFurnace}}
	n.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	n.Init(n)
	return n
}

// BlastFurnace is a port of pocketmine\block\tile\BlastFurnace.
type BlastFurnace struct{ Furnace }

func NewBlastFurnace(world World, pos math.Vector3) *BlastFurnace {
	b := &BlastFurnace{Furnace{furnaceType: FurnaceTypeBlastFurnace}}
	b.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	b.Init(b)
	return b
}

func (b *BlastFurnace) SaveID() string { return "BlastFurnace" }

// Smoker is a port of pocketmine\block\tile\Smoker.
type Smoker struct{ Furnace }

func NewSmoker(world World, pos math.Vector3) *Smoker {
	s := &Smoker{Furnace{furnaceType: FurnaceTypeSmoker}}
	s.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	s.Init(s)
	return s
}

func (s *Smoker) SaveID() string { return "Smoker" }
