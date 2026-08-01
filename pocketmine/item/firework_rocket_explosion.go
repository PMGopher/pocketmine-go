package item

import (
	"fmt"

	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/nbt"
)

const (
	fireworkTagType       = "FireworkType"
	fireworkTagColors     = "FireworkColor"
	fireworkTagFadeColors = "FireworkFade"
	fireworkTagTwinkle    = "FireworkFlicker"
	fireworkTagTrail      = "FireworkTrail"
)

// FireworkRocketExplosion is a port of pocketmine\item\FireworkRocketExplosion.
//
// The PHP constructor's validation (colors must be non-empty, both slices must contain only
// DyeColor values - trivially true here since Go is statically typed) partially carries over:
// NewFireworkRocketExplosion panics on an empty colors slice, matching the PHP original's
// InvalidArgumentException (a programmer error at the call site), the same convention used
// elsewhere in this port (e.g. block.PitcherCrop.SetAge).
type FireworkRocketExplosion struct {
	RocketType FireworkRocketType
	Colors     []blockutils.DyeColor
	FadeColors []blockutils.DyeColor
	Twinkle    bool
	Trail      bool
}

func NewFireworkRocketExplosion(rocketType FireworkRocketType, colors, fadeColors []blockutils.DyeColor, twinkle, trail bool) FireworkRocketExplosion {
	if len(colors) == 0 {
		panic("Colors list cannot be empty")
	}
	return FireworkRocketExplosion{RocketType: rocketType, Colors: colors, FadeColors: fadeColors, Twinkle: twinkle, Trail: trail}
}

func (e FireworkRocketExplosion) GetType() FireworkRocketType { return e.RocketType }

func (e FireworkRocketExplosion) GetColors() []blockutils.DyeColor { return e.Colors }

// GetFlashColor is a port of FireworkRocketExplosion::getFlashColor.
func (e FireworkRocketExplosion) GetFlashColor() blockutils.DyeColor { return e.Colors[0] }

// GetColorMix is a port of FireworkRocketExplosion::getColorMix.
func (e FireworkRocketExplosion) GetColorMix() color.Color {
	rgb := make([]color.Color, len(e.Colors))
	for i, c := range e.Colors {
		rgb[i] = c.GetRgbValue()
	}
	return color.Mix(rgb[0], rgb[1:]...)
}

func (e FireworkRocketExplosion) GetFadeColors() []blockutils.DyeColor { return e.FadeColors }

func (e FireworkRocketExplosion) WillTwinkle() bool { return e.Twinkle }

func (e FireworkRocketExplosion) GetTrail() bool { return e.Trail }

func decodeFireworkColors(colorsBytes []byte) ([]blockutils.DyeColor, error) {
	colors := make([]blockutils.DyeColor, 0, len(colorsBytes))
	for _, b := range colorsBytes {
		c, ok := blockutils.DyeColorFromInvertedID(int(b))
		if !ok {
			return nil, fmt.Errorf("unknown color %d", b)
		}
		colors = append(colors, c)
	}
	return colors, nil
}

func encodeFireworkColors(colors []blockutils.DyeColor) []byte {
	result := make([]byte, len(colors))
	for i, c := range colors {
		result[i] = byte(blockutils.DyeColorToInvertedID(c) & 0xff)
	}
	return result
}

// FireworkRocketExplosionFromCompoundTag is a port of FireworkRocketExplosion::fromCompoundTag.
func FireworkRocketExplosionFromCompoundTag(tag *nbt.CompoundTag) (FireworkRocketExplosion, error) {
	colorBytes, err := tag.GetByteArray(fireworkTagColors)
	if err != nil {
		return FireworkRocketExplosion{}, err
	}
	colors, err := decodeFireworkColors(colorBytes)
	if err != nil {
		return FireworkRocketExplosion{}, err
	}
	if len(colors) == 0 {
		return FireworkRocketExplosion{}, fmt.Errorf("colors list cannot be empty")
	}

	fadeBytes, _ := tag.GetByteArray(fireworkTagFadeColors)
	fadeColors, err := decodeFireworkColors(fadeBytes)
	if err != nil {
		return FireworkRocketExplosion{}, err
	}

	rocketTypeByte, err := tag.GetByte(fireworkTagType)
	if err != nil {
		return FireworkRocketExplosion{}, err
	}
	if int(rocketTypeByte) < int(FireworkRocketTypeSmallBall) || int(rocketTypeByte) > int(FireworkRocketTypeBurst) {
		return FireworkRocketExplosion{}, fmt.Errorf("invalid firework type %d", rocketTypeByte)
	}

	return FireworkRocketExplosion{
		RocketType: FireworkRocketType(rocketTypeByte),
		Colors:     colors,
		FadeColors: fadeColors,
		Twinkle:    tag.GetByteOr(fireworkTagTwinkle, 0) != 0,
		Trail:      tag.GetByteOr(fireworkTagTrail, 0) != 0,
	}, nil
}

// ToCompoundTag is a port of FireworkRocketExplosion::toCompoundTag.
func (e FireworkRocketExplosion) ToCompoundTag() *nbt.CompoundTag {
	twinkle, trail := nbt.ByteTag(0), nbt.ByteTag(0)
	if e.Twinkle {
		twinkle = 1
	}
	if e.Trail {
		trail = 1
	}
	return nbt.NewCompoundTag().
		SetByte(fireworkTagType, nbt.ByteTag(e.RocketType)).
		SetByteArray(fireworkTagColors, encodeFireworkColors(e.Colors)).
		SetByteArray(fireworkTagFadeColors, encodeFireworkColors(e.FadeColors)).
		SetByte(fireworkTagTwinkle, twinkle).
		SetByte(fireworkTagTrail, trail)
}
