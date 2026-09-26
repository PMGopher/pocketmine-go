package tile

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	flowerPotTagItem       = "item"
	flowerPotTagItemData   = "mData"
	flowerPotTagPlantBlock = "PlantBlock"
)

// FlowerPot is a port of pocketmine\block\tile\FlowerPot: the plant in a flower pot.
type FlowerPot struct {
	SpawnableBase

	plant Block
}

func NewFlowerPot(world World, pos math.Vector3) *FlowerPot {
	f := &FlowerPot{}
	f.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	f.Init(f)
	return f
}

func (f *FlowerPot) SaveID() string { return "FlowerPot" }

// ReadSaveData is a port of FlowerPot::readSaveData: the legacy item ID + meta, or the plant's
// blockstate.
func (f *FlowerPot) ReadSaveData(tag *nbt.CompoundTag) error {
	itemIDTag, _ := tag.GetTag(flowerPotTagItem)
	itemMetaTag, _ := tag.GetTag(flowerPotTagItemData)
	itemID, isShort := itemIDTag.(nbt.ShortTag)
	itemMeta, isInt := itemMetaTag.(nbt.IntTag)
	var plant Block
	if isShort && isInt {
		if BlockFromLegacyIdMetaFunc == nil {
			return nil
		}
		b, err := BlockFromLegacyIdMetaFunc(int(itemID), int(itemMeta))
		if err != nil {
			return &SavedDataLoadingError{"Error loading legacy flower pot item data: " + err.Error()}
		}
		plant = b
	} else if plantBlockTag, ok, _ := tag.GetCompoundTag(flowerPotTagPlantBlock); ok {
		if BlockFromNbtFunc == nil {
			return nil
		}
		b, err := BlockFromNbtFunc(plantBlockTag)
		if err != nil {
			return &SavedDataLoadingError{"Error loading " + flowerPotTagPlantBlock + " tag for flower pot: " + err.Error()}
		}
		plant = b
	}
	if plant != nil {
		f.SetPlant(plant)
	}
	return nil
}

// WriteSaveData is a port of FlowerPot::writeSaveData.
func (f *FlowerPot) WriteSaveData(tag *nbt.CompoundTag) {
	if f.plant != nil && BlockToNbtFunc != nil {
		if plantTag, err := BlockToNbtFunc(f.plant); err == nil {
			tag.SetTag(flowerPotTagPlantBlock, plantTag)
		}
	}
}

// GetPlant is a port of FlowerPot::getPlant (nil for no plant).
func (f *FlowerPot) GetPlant() Block { return f.plant }

// SetPlant is a port of FlowerPot::setPlant: air counts as no plant.
func (f *FlowerPot) SetPlant(plant Block) {
	if plant == nil || (IsAirFunc != nil && IsAirFunc(plant)) {
		f.plant = nil
		return
	}
	f.plant = plant
}

// AddAdditionalSpawnData is a port of FlowerPot::addAdditionalSpawnData.
func (f *FlowerPot) AddAdditionalSpawnData(tag *nbt.CompoundTag) {
	if f.plant != nil && NetworkBlockStateNbtFunc != nil {
		if plantTag := NetworkBlockStateNbtFunc(f.plant); plantTag != nil {
			tag.SetTag(flowerPotTagPlantBlock, plantTag)
		}
	}
}

// GetRenderUpdateBugWorkaroundStateProperties is a port of
// FlowerPot::getRenderUpdateBugWorkaroundStateProperties.
func (f *FlowerPot) GetRenderUpdateBugWorkaroundStateProperties(b Block) map[string]any {
	return map[string]any{"update_bit": uint8(1)} // BlockStateNames::UPDATE_BIT
}
