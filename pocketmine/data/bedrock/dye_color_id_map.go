package bedrock

import (
	"sync"

	blockutils "pocketmine-go/pocketmine/block/utils"
)

// DyeColorIdMap is a port of pocketmine\data\bedrock\DyeColorIdMap: dye colour save IDs and the
// dye item ID of each colour.
type DyeColorIdMap struct {
	*IntSaveIdMap[blockutils.DyeColor]
	itemIdToEnum map[string]blockutils.DyeColor
	enumToItemId map[blockutils.DyeColor]string
}

var (
	dyeColorIdMap     *DyeColorIdMap
	dyeColorIdMapOnce sync.Once
)

// GetDyeColorIdMap is DyeColorIdMap::getInstance().
func GetDyeColorIdMap() *DyeColorIdMap {
	dyeColorIdMapOnce.Do(func() {
		m := &DyeColorIdMap{IntSaveIdMap: newIntSaveIdMap[blockutils.DyeColor](), itemIdToEnum: map[string]blockutils.DyeColor{}, enumToItemId: map[blockutils.DyeColor]string{}}
		for _, e := range []struct {
			color     blockutils.DyeColor
			id        int
			dyeItemID string
		}{
			{blockutils.DyeColorWhite, 0, "minecraft:white_dye"},
			{blockutils.DyeColorOrange, 1, "minecraft:orange_dye"},
			{blockutils.DyeColorMagenta, 2, "minecraft:magenta_dye"},
			{blockutils.DyeColorLightBlue, 3, "minecraft:light_blue_dye"},
			{blockutils.DyeColorYellow, 4, "minecraft:yellow_dye"},
			{blockutils.DyeColorLime, 5, "minecraft:lime_dye"},
			{blockutils.DyeColorPink, 6, "minecraft:pink_dye"},
			{blockutils.DyeColorGray, 7, "minecraft:gray_dye"},
			{blockutils.DyeColorLightGray, 8, "minecraft:light_gray_dye"},
			{blockutils.DyeColorCyan, 9, "minecraft:cyan_dye"},
			{blockutils.DyeColorPurple, 10, "minecraft:purple_dye"},
			{blockutils.DyeColorBlue, 11, "minecraft:blue_dye"},
			{blockutils.DyeColorBrown, 12, "minecraft:brown_dye"},
			{blockutils.DyeColorGreen, 13, "minecraft:green_dye"},
			{blockutils.DyeColorRed, 14, "minecraft:red_dye"},
			{blockutils.DyeColorBlack, 15, "minecraft:black_dye"},
		} {
			m.Register(e.id, e.color)
			m.itemIdToEnum[e.dyeItemID] = e.color
			m.enumToItemId[e.color] = e.dyeItemID
		}
		dyeColorIdMap = m
	})
	return dyeColorIdMap
}

// ToInvertedID is a port of DyeColorIdMap::toInvertedId.
func (m *DyeColorIdMap) ToInvertedID(color blockutils.DyeColor) int { return ^m.ToID(color) & 0xf }

// ToItemID is a port of DyeColorIdMap::toItemId.
func (m *DyeColorIdMap) ToItemID(color blockutils.DyeColor) string { return m.enumToItemId[color] }

// FromInvertedID is a port of DyeColorIdMap::fromInvertedId.
func (m *DyeColorIdMap) FromInvertedID(id int) (blockutils.DyeColor, bool) {
	return m.FromID(^id & 0xf)
}

// FromItemID is a port of DyeColorIdMap::fromItemId.
func (m *DyeColorIdMap) FromItemID(itemID string) (blockutils.DyeColor, bool) {
	c, ok := m.itemIdToEnum[itemID]
	return c, ok
}
