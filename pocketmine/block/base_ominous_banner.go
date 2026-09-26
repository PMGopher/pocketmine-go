package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// BaseOminousBanner is a port of pocketmine\block\BaseOminousBanner. Like BaseBanner, this isn't
// meant to be instantiated directly - a concrete leaf type (OminousFloorBanner,
// OminousWallBanner) must embed it, implement Clone, and satisfy bannerShaper.
type BaseOminousBanner struct {
	Transparent
}

func (b *BaseOminousBanner) IsSolid() bool { return false }

func (b *BaseOminousBanner) GetMaxStackSize() int { return 16 }

func (b *BaseOminousBanner) GetFuelTime() int { return 300 }

func (b *BaseOminousBanner) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (b *BaseOminousBanner) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (b *BaseOminousBanner) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	supportingFace := b.self.(bannerShaper).GetSupportingFace()
	if !bannerCanBeSupportedBy(blockReplace.(blockGeometry).GetSide(supportingFace, 1)) {
		return false
	}
	return b.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (b *BaseOminousBanner) OnNearbyBlockChange() {
	supportingFace := b.self.(bannerShaper).GetSupportingFace()
	if !bannerCanBeSupportedBy(b.self.(blockGeometry).GetSide(supportingFace, 1)) {
		if world, err := b.position.GetWorld(); err == nil {
			world.UseBreakOn(b.position.AsVector3())
		}
	}
}

// WriteStateToWorld is a port of BaseOminousBanner::writeStateToWorld.
func (b *BaseOminousBanner) WriteStateToWorld() {
	b.Block.WriteStateToWorld()
	if t, ok := b.tileAt(); ok {
		if bannerTile, ok := t.(*tile.Banner); ok {
			bannerTile.SetBaseColor(blockutils.DyeColorWhite)
			bannerTile.SetPatterns(nil)
			bannerTile.SetType(tile.BannerTypeOminous)
		}
	}
}
