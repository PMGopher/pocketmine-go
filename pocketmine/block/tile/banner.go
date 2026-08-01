package tile

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
)

const (
	BannerTagBase         = "Base"
	BannerTagPatterns     = "Patterns"
	BannerTagPatternColor = "Color"
	BannerTagPatternName  = "Pattern"
	BannerTagType         = "Type"

	BannerTypeNormal  = 0
	BannerTypeOminous = 1
)

// Banner is a port of pocketmine\block\tile\Banner.
//
// Deprecated in the PHP original too - see block.BaseBanner (not ported yet).
type Banner struct {
	SpawnableBase

	BaseColor  blockutils.DyeColor
	Patterns   []blockutils.BannerPatternLayer
	BannerType int
}

func NewBanner(world World, pos math.Vector3) *Banner {
	b := &Banner{SpawnableBase: SpawnableBase{TileBase: NewTileBase(world, pos)}, BaseColor: blockutils.DyeColorBlack}
	b.Init(b)
	return b
}

func (b *Banner) SaveID() string { return "Banner" }

// ReadSaveData is a port of Banner::readSaveData.
//
// TODO (from the PHP original): a missing/invalid base colour silently falls back to Black rather
// than erroring; similarly, a pattern with a missing colour falls back to Black, and a pattern
// with an unrecognised type is silently skipped.
func (b *Banner) ReadSaveData(tag *nbt.CompoundTag) error {
	b.BaseColor = blockutils.DyeColorBlack
	if baseColorTag, ok := tag.GetTag(BannerTagBase); ok {
		if intTag, ok := baseColorTag.(nbt.IntTag); ok {
			if c, ok := blockutils.DyeColorFromInvertedID(int(intTag)); ok {
				b.BaseColor = c
			}
		}
	}

	b.Patterns = nil
	if patterns, ok, err := tag.GetListTag(BannerTagPatterns); err == nil && ok {
		for _, entry := range patterns.Values() {
			pattern, ok := entry.(*nbt.CompoundTag)
			if !ok {
				continue
			}
			patternColor := blockutils.DyeColorBlack
			if c, ok := blockutils.DyeColorFromInvertedID(int(pattern.GetIntOr(BannerTagPatternColor, 0))); ok {
				patternColor = c
			}
			name, err := pattern.GetString(BannerTagPatternName)
			if err != nil {
				continue
			}
			patternType, ok := blockutils.BannerPatternTypeFromID(string(name))
			if !ok {
				continue
			}
			b.Patterns = append(b.Patterns, blockutils.NewBannerPatternLayer(patternType, patternColor))
		}
	}

	b.BannerType = int(tag.GetIntOr(BannerTagType, BannerTypeNormal))
	return nil
}

func (b *Banner) writePatterns(tag *nbt.CompoundTag) {
	tag.SetInt(BannerTagBase, nbt.IntTag(blockutils.DyeColorToInvertedID(b.BaseColor)))
	patterns, _ := nbt.NewListTag(nil, nbt.TagCompound)
	for _, pattern := range b.Patterns {
		id, _ := blockutils.BannerPatternTypeToID(pattern.GetType())
		entry := nbt.NewCompoundTag()
		entry.SetString(BannerTagPatternName, nbt.StringTag(id))
		entry.SetInt(BannerTagPatternColor, nbt.IntTag(blockutils.DyeColorToInvertedID(pattern.GetColor())))
		_ = patterns.Push(entry)
	}
	tag.SetTag(BannerTagPatterns, patterns)
	tag.SetInt(BannerTagType, nbt.IntTag(b.BannerType))
}

func (b *Banner) WriteSaveData(tag *nbt.CompoundTag) { b.writePatterns(tag) }

func (b *Banner) AddAdditionalSpawnData(tag *nbt.CompoundTag) { b.writePatterns(tag) }

func (b *Banner) GetBaseColor() blockutils.DyeColor { return b.BaseColor }

func (b *Banner) SetBaseColor(c blockutils.DyeColor) { b.BaseColor = c }

func (b *Banner) GetPatterns() []blockutils.BannerPatternLayer { return b.Patterns }

func (b *Banner) SetPatterns(patterns []blockutils.BannerPatternLayer) { b.Patterns = patterns }

func (b *Banner) GetType() int { return b.BannerType }

func (b *Banner) SetType(t int) { b.BannerType = t }

func (b *Banner) GetDefaultName() string { return "Banner" }
