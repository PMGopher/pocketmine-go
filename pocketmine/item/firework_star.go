package item

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/color"
	"pocketmine-go/pocketmine/nbt"
)

const (
	fireworkStarTagExplosion   = "FireworksItem"
	fireworkStarTagCustomColor = "customColor"
)

// FireworkStar is a port of pocketmine\item\FireworkStar.
type FireworkStar struct {
	ItemBase

	Explosion      FireworkRocketExplosion
	customColor    color.Color
	hasCustomColor bool
}

func NewFireworkStar(identifier ItemIdentifier, name string) *FireworkStar {
	f := &FireworkStar{
		Explosion: NewFireworkRocketExplosion(FireworkRocketTypeSmallBall, []blockutils.DyeColor{blockutils.DyeColorBlack}, nil, false, false),
	}
	f.Init(f, identifier, name)
	return f
}

func (f *FireworkStar) Clone() Item {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *FireworkStar) GetExplosion() FireworkRocketExplosion { return f.Explosion }

func (f *FireworkStar) SetExplosion(explosion FireworkRocketExplosion) { f.Explosion = explosion }

// GetColor is a port of FireworkStar::getColor.
func (f *FireworkStar) GetColor() color.Color {
	if f.hasCustomColor {
		return f.customColor
	}
	return f.Explosion.GetColorMix()
}

func (f *FireworkStar) GetCustomColor() (color.Color, bool) { return f.customColor, f.hasCustomColor }

func (f *FireworkStar) SetCustomColor(c color.Color) { f.customColor = c; f.hasCustomColor = true }

func (f *FireworkStar) ClearCustomColor() { f.customColor = color.Color{}; f.hasCustomColor = false }

// deserializeCompoundTag/serializeCompoundTag extend ItemBase's own pair, the same self-dispatch
// participation described on Durable's. The ARGB round trip skips PHP's Binary::signInt/
// unsignInt - see Armor's doc comment for why.
func (f *FireworkStar) deserializeCompoundTag(tag *nbt.CompoundTag) {
	f.ItemBase.deserializeCompoundTag(tag)

	explosionTag, ok, _ := tag.GetCompoundTag(fireworkStarTagExplosion)
	if ok {
		if explosion, err := FireworkRocketExplosionFromCompoundTag(explosionTag); err == nil {
			f.Explosion = explosion
		}
	}

	f.hasCustomColor = false
	if customColorTag, err := tag.GetInt(fireworkStarTagCustomColor); err == nil {
		customColor := color.FromARGB(int32(customColorTag))
		if !customColor.Equals(f.Explosion.GetColorMix()) {
			f.customColor = customColor
			f.hasCustomColor = true
		}
	}
}

func (f *FireworkStar) serializeCompoundTag(tag *nbt.CompoundTag) {
	f.ItemBase.serializeCompoundTag(tag)
	tag.SetTag(fireworkStarTagExplosion, f.Explosion.ToCompoundTag())
	tag.SetInt(fireworkStarTagCustomColor, nbt.IntTag(f.GetColor().ToARGB()))
}
