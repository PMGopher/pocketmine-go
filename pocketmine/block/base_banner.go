package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// bannerShaper lets concrete leaf types (FloorBanner, WallBanner) report which side supports them
// - same self-dispatch shape as every other *Shaper in this port.
type bannerShaper interface {
	GetSupportingFace() math.Facing
}

// ominousBannerVersioner is satisfied by BaseBanner's concrete leaves (FloorBanner, WallBanner)
// that override getOminousVersion - separate from bannerShaper since BaseOminousBanner's own
// leaves (OminousFloorBanner, OminousWallBanner) satisfy bannerShaper but have no ominous version
// of their own to report.
type ominousBannerVersioner interface {
	GetOminousVersion() Behavior
}

// GetOminousVersion is a port of BaseBanner::getOminousVersion's default (VanillaBlocks::AIR()).
// FloorBanner/WallBanner override this to their real ominous counterpart.
func (b *BaseBanner) GetOminousVersion() Behavior { return VanillaAir() }

// BaseBanner is a port of pocketmine\block\BaseBanner. Like Crops/Stem/BaseCoral, this isn't
// meant to be instantiated directly - a concrete leaf type (FloorBanner, WallBanner) must embed
// it, implement Clone, and satisfy bannerShaper.
type BaseBanner struct {
	Transparent
	ColorComponent

	Patterns []blockutils.BannerPatternLayer
}

// ReadStateFromWorld is a port of BaseBanner::readStateFromWorld.
func (b *BaseBanner) ReadStateFromWorld() Behavior {
	b.Block.ReadStateFromWorld()

	world, err := b.position.GetWorld()
	if err != nil {
		return b.self
	}
	t, _ := world.GetTile(b.position)
	if bannerTile, ok := t.(*tile.Banner); ok {
		if bannerTile.GetType() == tile.BannerTypeOminous {
			// illager banner is implemented as a separate block, as it doesn't support base color
			// or custom patterns
			return b.self.(ominousBannerVersioner).GetOminousVersion()
		}
		b.Color = bannerTile.GetBaseColor()
		b.Patterns = bannerTile.GetPatterns()
	}
	return b.self
}

func (b *BaseBanner) IsSolid() bool { return false }

func (b *BaseBanner) GetMaxStackSize() int { return 16 }

func (b *BaseBanner) GetPatterns() []blockutils.BannerPatternLayer { return b.Patterns }

func (b *BaseBanner) SetPatterns(patterns []blockutils.BannerPatternLayer) { b.Patterns = patterns }

func (b *BaseBanner) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (b *BaseBanner) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func bannerCanBeSupportedBy(blk Behavior) bool { return blk.IsSolid() }

func (b *BaseBanner) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	supportingFace := b.self.(bannerShaper).GetSupportingFace()
	if !bannerCanBeSupportedBy(blockReplace.(blockGeometry).GetSide(supportingFace, 1)) {
		return false
	}
	// The PHP original also copies colour/patterns from an ItemBanner here - needs the unported
	// item package, so it's skipped (documented gap, same category as everywhere else real Item
	// construction is needed).
	return b.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (b *BaseBanner) OnNearbyBlockChange() {
	supportingFace := b.self.(bannerShaper).GetSupportingFace()
	if !bannerCanBeSupportedBy(b.self.(blockGeometry).GetSide(supportingFace, 1)) {
		if world, err := b.position.GetWorld(); err == nil {
			world.UseBreakOn(b.position.AsVector3())
		}
	}
}

// GetDropsForCompatibleTool/GetPickedItem/AsItem should return an ItemBanner carrying Patterns -
// needs real Item construction from the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so all three are left as Block's defaults for
// now.
