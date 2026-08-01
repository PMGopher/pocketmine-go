package item

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/nbt"
)

const (
	bannerTagPatterns     = "Patterns"
	bannerTagPatternColor = "Color"
	bannerTagPatternName  = "Pattern"
)

// Banner is a port of pocketmine\item\Banner. In PHP this extends ItemBlockWallOrFloor, whose
// GetBlock needs RuntimeBlockStateRegistry (the full block registry, not ported) - since GetBlock
// isn't part of Item here at all yet (see the Item interface's doc comment), this embeds ItemBase
// directly instead of porting that base class.
type Banner struct {
	ItemBase

	Color    blockutils.DyeColor
	Patterns []blockutils.BannerPatternLayer
}

func NewBanner(identifier ItemIdentifier, name string) *Banner {
	b := &Banner{Color: blockutils.DyeColorBlack}
	b.Init(b, identifier, name)
	return b
}

func (b *Banner) Clone() Item {
	c := *b
	c.Patterns = append([]blockutils.BannerPatternLayer(nil), b.Patterns...)
	c.rebind(&c)
	return &c
}

func (b *Banner) GetColor() blockutils.DyeColor { return b.Color }

func (b *Banner) SetColor(color blockutils.DyeColor) { b.Color = color }

func (b *Banner) GetPatterns() []blockutils.BannerPatternLayer { return b.Patterns }

func (b *Banner) SetPatterns(patterns []blockutils.BannerPatternLayer) { b.Patterns = patterns }

func (b *Banner) GetFuelTime() int { return 300 }

func (b *Banner) describeState(w runtime.DataDescriber) {
	col := int(b.Color)
	w.BoundedIntAuto(int(blockutils.DyeColorWhite), int(blockutils.DyeColorBlack), &col)
	b.Color = blockutils.DyeColor(col)
}

// deserializeCompoundTag/serializeCompoundTag extend ItemBase's own pair, the same self-dispatch
// participation described on Durable's. The inverted-DyeColor-ID byte encoding matches
// tile.Banner's (see that file's doc comment).
func (b *Banner) deserializeCompoundTag(tag *nbt.CompoundTag) {
	b.ItemBase.deserializeCompoundTag(tag)
	b.Patterns = nil

	patterns, ok, _ := tag.GetListTag(bannerTagPatterns)
	if !ok {
		return
	}
	for _, v := range patterns.Values() {
		patternTag, ok := v.(*nbt.CompoundTag)
		if !ok {
			continue
		}
		patternColor, ok := blockutils.DyeColorFromInvertedID(int(patternTag.GetIntOr(bannerTagPatternColor, 0)))
		if !ok {
			patternColor = blockutils.DyeColorBlack
		}
		patternType, ok := blockutils.BannerPatternTypeFromID(string(patternTag.GetStringOr(bannerTagPatternName, "")))
		if !ok {
			continue
		}
		b.Patterns = append(b.Patterns, blockutils.NewBannerPatternLayer(patternType, patternColor))
	}
}

func (b *Banner) serializeCompoundTag(tag *nbt.CompoundTag) {
	b.ItemBase.serializeCompoundTag(tag)

	if len(b.Patterns) == 0 {
		tag.RemoveTag(bannerTagPatterns)
		return
	}
	values := make([]nbt.Tag, len(b.Patterns))
	for i, p := range b.Patterns {
		id, _ := blockutils.BannerPatternTypeToID(p.GetType())
		values[i] = nbt.NewCompoundTag().
			SetString(bannerTagPatternName, nbt.StringTag(id)).
			SetInt(bannerTagPatternColor, nbt.IntTag(blockutils.DyeColorToInvertedID(p.GetColor())))
	}
	patternsTag, err := nbt.NewListTag(values, nbt.TagCompound)
	if err != nil {
		panic(err)
	}
	tag.SetTag(bannerTagPatterns, patternsTag)
}
