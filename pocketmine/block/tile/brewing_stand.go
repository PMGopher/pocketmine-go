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

// BrewingStand is a port of pocketmine\block\tile\BrewingStand, minus its inventory/Container
// half - see ContainerComponent's doc comment for why the inventory package can't be imported
// here.
//
// The brew/fuel time state and its NBT round trip (including the PHP original's legacy-tag
// fallback chain and load-time consistency fixups: maxFuelTime defaulting to remainingFuelTime,
// and everything zeroing out if remainingFuelTime is 0) are fully real. checkFuel/
// getBrewableRecipes/onUpdate (the actual brewing simulation) all need the fuel/ingredient/bottle
// inventory slots plus BrewingRecipe/CraftingManager (crafting package, not ported), so none of
// that is ported.
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
	b.writeState(tag)
}

func (b *BrewingStand) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	b.NameableComponent.AddAdditionalSpawnData(tag)
	b.writeState(tag)
}

// CopyDataFromItem must be defined here rather than relying on promotion - see
// NameableComponent.ApplyItemCustomName's doc comment for why.
func (b *BrewingStand) CopyDataFromItem(item Item) {
	b.TileBase.CopyDataFromItem(item)
	b.ApplyItemCustomName(item)
}
