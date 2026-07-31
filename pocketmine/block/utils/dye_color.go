package blockutils

import "pocketmine-go/pocketmine/color"

// DyeColor is a port of pocketmine\block\utils\DyeColor.
type DyeColor int

const (
	DyeColorWhite DyeColor = iota
	DyeColorOrange
	DyeColorMagenta
	DyeColorLightBlue
	DyeColorYellow
	DyeColorLime
	DyeColorPink
	DyeColorGray
	DyeColorLightGray
	DyeColorCyan
	DyeColorPurple
	DyeColorBlue
	DyeColorBrown
	DyeColorGreen
	DyeColorRed
	DyeColorBlack
)

type dyeColorMetadata struct {
	displayName string
	rgbValue    color.Color
}

var dyeColorMetadataTable = map[DyeColor]dyeColorMetadata{
	DyeColorWhite:     {"White", color.NewColor(0xf0, 0xf0, 0xf0)},
	DyeColorOrange:    {"Orange", color.NewColor(0xf9, 0x80, 0x1d)},
	DyeColorMagenta:   {"Magenta", color.NewColor(0xc7, 0x4e, 0xbd)},
	DyeColorLightBlue: {"Light Blue", color.NewColor(0x3a, 0xb3, 0xda)},
	DyeColorYellow:    {"Yellow", color.NewColor(0xfe, 0xd8, 0x3d)},
	DyeColorLime:      {"Lime", color.NewColor(0x80, 0xc7, 0x1f)},
	DyeColorPink:      {"Pink", color.NewColor(0xf3, 0x8b, 0xaa)},
	DyeColorGray:      {"Gray", color.NewColor(0x47, 0x4f, 0x52)},
	DyeColorLightGray: {"Light Gray", color.NewColor(0x9d, 0x9d, 0x97)},
	DyeColorCyan:      {"Cyan", color.NewColor(0x16, 0x9c, 0x9c)},
	DyeColorPurple:    {"Purple", color.NewColor(0x89, 0x32, 0xb8)},
	DyeColorBlue:      {"Blue", color.NewColor(0x3c, 0x44, 0xaa)},
	DyeColorBrown:     {"Brown", color.NewColor(0x83, 0x54, 0x32)},
	DyeColorGreen:     {"Green", color.NewColor(0x5e, 0x7c, 0x16)},
	DyeColorRed:       {"Red", color.NewColor(0xb0, 0x2e, 0x26)},
	DyeColorBlack:     {"Black", color.NewColor(0x1d, 0x1d, 0x21)},
}

func (d DyeColor) GetDisplayName() string { return dyeColorMetadataTable[d].displayName }

func (d DyeColor) GetRgbValue() color.Color { return dyeColorMetadataTable[d].rgbValue }
