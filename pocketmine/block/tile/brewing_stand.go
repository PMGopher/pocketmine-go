package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	BrewingStandTagBrewTime                = "CookTime"
	BrewingStandTagBrewTimeLegacy          = "BrewTime"
	BrewingStandTagMaxFuelTime             = "FuelTotal"
	BrewingStandTagRemainingFuelTime       = "FuelAmount"
	BrewingStandTagRemainingFuelTimeLegacy = "Fuel"

	BrewingStandBrewTimeTicks = 400
)

// BrewingStandOnUpdateFunc is BrewingStand::onUpdate (the brewing simulation), set by
// block/inventory, which can import the packages it needs.
var BrewingStandOnUpdateFunc func(b *BrewingStand) bool

// BrewingStand is a port of pocketmine\block\tile\BrewingStand.
type BrewingStand struct {
	SpawnableBase
	NameableComponent
	ContainerComponent

	BrewTime          int
	MaxFuelTime       int
	RemainingFuelTime int
}

func NewBrewingStand(world World, pos math.Vector3) *BrewingStand {
	b := &BrewingStand{}
	b.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	b.Init(b)
	return b
}

func (b *BrewingStand) SaveID() string { return "BrewingStand" }

func (b *BrewingStand) GetDefaultName() string { return "Brewing Stand" }

func (b *BrewingStand) GetName() string { return b.NameableComponent.GetName(b) }

// ReadSaveData is a port of BrewingStand::readSaveData.
func (b *BrewingStand) ReadSaveData(tag *nbt.CompoundTag) error {
	b.LoadName(tag)
	b.loadItems(b, tag)

	// PHP checks the legacy "BrewTime" tag first, falling back to the PE "CookTime" tag - which is
	// the only one WriteSaveData/AddAdditionalSpawnData below actually write, so on a save
	// produced by this port "BrewTime" is always absent and "CookTime" is used.
	if v, err := tag.GetShort(BrewingStandTagBrewTimeLegacy); err == nil {
		b.BrewTime = int(v)
	} else if v, err := tag.GetShort(BrewingStandTagBrewTime); err == nil {
		b.BrewTime = int(v)
	} else {
		b.BrewTime = 0
	}

	b.MaxFuelTime = int(tag.GetShortOr(BrewingStandTagMaxFuelTime, 0))

	if v, err := tag.GetByte(BrewingStandTagRemainingFuelTimeLegacy); err == nil {
		b.RemainingFuelTime = int(v)
	} else if v, err := tag.GetShort(BrewingStandTagRemainingFuelTime); err == nil {
		b.RemainingFuelTime = int(v)
	} else {
		b.RemainingFuelTime = 0
	}

	if b.MaxFuelTime == 0 {
		b.MaxFuelTime = b.RemainingFuelTime
	}
	if b.RemainingFuelTime == 0 {
		b.MaxFuelTime, b.RemainingFuelTime, b.BrewTime = 0, 0, 0
	}
	return nil
}

func (b *BrewingStand) writeState(tag *nbt.CompoundTag) {
	tag.SetShort(BrewingStandTagBrewTime, nbt.ShortTag(b.BrewTime))
	tag.SetShort(BrewingStandTagMaxFuelTime, nbt.ShortTag(b.MaxFuelTime))
	tag.SetShort(BrewingStandTagRemainingFuelTime, nbt.ShortTag(b.RemainingFuelTime))
}

func (b *BrewingStand) WriteSaveData(tag *nbt.CompoundTag) {
	b.SaveName(tag)
	b.saveItems(b, tag)
	b.writeState(tag)
}

func (b *BrewingStand) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	b.NameableComponent.AddAdditionalSpawnData(tag)
	b.writeState(tag)
}

// GetInventory is a port of BrewingStand::getInventory (a BrewingStandInventory).
func (b *BrewingStand) GetInventory() Inventory { return b.realInventory(b) }

// GetRealInventory is a port of BrewingStand::getRealInventory.
func (b *BrewingStand) GetRealInventory() Inventory { return b.realInventory(b) }

// CloseHook is BrewingStand::close's removal of the inventory's viewers.
func (b *BrewingStand) CloseHook() { b.removeAllViewers() }

// OnBlockDestroyedHook is ContainerTrait::onBlockDestroyedHook.
func (b *BrewingStand) OnBlockDestroyedHook() { b.dropContents(b) }

// OnUpdate is a port of BrewingStand::onUpdate: whether it's still brewing.
func (b *BrewingStand) OnUpdate() bool {
	if b.closed || BrewingStandOnUpdateFunc == nil {
		return false
	}
	return BrewingStandOnUpdateFunc(b)
}

// CopyDataFromItem must be defined here rather than relying on promotion - see
// NameableComponent.ApplyItemCustomName's doc comment for why.
func (b *BrewingStand) CopyDataFromItem(item Item) {
	b.TileBase.CopyDataFromItem(item)
	b.ApplyItemCustomName(item)
}
