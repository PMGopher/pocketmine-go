package tile

import (
	"fmt"

	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

// Cauldron potion container types and potion ID, a port of Cauldron's private constants.
const (
	CauldronPotionContainerTypeNone      = -1
	CauldronPotionContainerTypeNormal    = 0
	CauldronPotionContainerTypeSplash    = 1
	CauldronPotionContainerTypeLingering = 2

	cauldronPotionIDNone = -1

	cauldronTagPotionID            = "PotionId"    // TAG_Short
	cauldronTagPotionContainerType = "PotionType"  // TAG_Short
	cauldronTagCustomColor         = "CustomColor" // TAG_Int
)

// Cauldron is a port of pocketmine\block\tile\Cauldron: the potion or custom water colour of a
// water/potion cauldron.
type Cauldron struct {
	SpawnableBase

	potionItem       Item
	customWaterColor *color.Color
}

func NewCauldron(world World, pos math.Vector3) *Cauldron {
	c := &Cauldron{}
	c.SpawnableBase = SpawnableBase{TileBase: NewTileBase(world, pos)}
	c.Init(c)
	return c
}

func (c *Cauldron) SaveID() string { return "Cauldron" }

func (c *Cauldron) GetPotionItem() Item               { return c.potionItem }
func (c *Cauldron) SetPotionItem(potionItem Item)     { c.potionItem = potionItem }
func (c *Cauldron) GetCustomWaterColor() *color.Color { return c.customWaterColor }

func (c *Cauldron) SetCustomWaterColor(customWaterColor *color.Color) {
	c.customWaterColor = customWaterColor
}

// potionInfo is the container type and potion ID Cauldron writes for its potion item.
func (c *Cauldron) potionInfo() (int, int) {
	if c.potionItem == nil || PotionItemInfoFunc == nil {
		return CauldronPotionContainerTypeNone, cauldronPotionIDNone
	}
	return PotionItemInfoFunc(c.potionItem)
}

func (c *Cauldron) writePotionData(tag *nbt.CompoundTag) {
	containerType, potionID := c.potionInfo()
	tag.SetShort(cauldronTagPotionContainerType, nbt.ShortTag(containerType))
	tag.SetShort(cauldronTagPotionID, nbt.ShortTag(potionID))
	if c.customWaterColor != nil {
		tag.SetInt(cauldronTagCustomColor, nbt.IntTag(c.customWaterColor.ToARGB()))
	}
}

func (c *Cauldron) AddAdditionalSpawnData(tag *nbt.CompoundTag) { c.writePotionData(tag) }

// ReadSaveData is a port of Cauldron::readSaveData.
func (c *Cauldron) ReadSaveData(tag *nbt.CompoundTag) error {
	containerType := int(tag.GetShortOr(cauldronTagPotionContainerType, CauldronPotionContainerTypeNone))
	potionID := int(tag.GetShortOr(cauldronTagPotionID, cauldronPotionIDNone))
	if containerType != CauldronPotionContainerTypeNone && potionID != cauldronPotionIDNone {
		if NewPotionItemFunc != nil {
			potion, err := NewPotionItemFunc(containerType, potionID)
			if err != nil {
				return &SavedDataLoadingError{err.Error()}
			}
			c.potionItem = potion
		}
	} else {
		c.potionItem = nil
	}

	c.customWaterColor = nil
	if t, ok := tag.GetTag(cauldronTagCustomColor); ok {
		if v, ok := t.(nbt.IntTag); ok {
			col := color.FromARGB(int32(v))
			c.customWaterColor = &col
		}
	}
	return nil
}

func (c *Cauldron) WriteSaveData(tag *nbt.CompoundTag) { c.writePotionData(tag) }

// CauldronContainerTypeError is Cauldron's "Invalid potion container type ID" error.
func CauldronContainerTypeError(containerType int) error {
	return fmt.Errorf("Invalid potion container type ID %d", containerType)
}

// GetRenderUpdateBugWorkaroundStateProperties is a port of
// Cauldron::getRenderUpdateBugWorkaroundStateProperties (FillableCauldron: fill levels 1-6).
func (c *Cauldron) GetRenderUpdateBugWorkaroundStateProperties(b Block) map[string]any {
	if fillable, ok := b.(interface{ GetFillLevel() int }); ok {
		realFillLevel := fillable.GetFillLevel()
		fake := realFillLevel + 1
		if realFillLevel == 6 { // FillableCauldron::MAX_FILL_LEVEL
			fake = 1 // FillableCauldron::MIN_FILL_LEVEL
		}
		return map[string]any{"fill_level": int32(fake)} // BlockStateNames::FILL_LEVEL
	}
	return nil
}
