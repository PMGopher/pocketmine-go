package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// BaseOminousBanner is a port of pocketmine\block\BaseOminousBanner. Like BaseBanner, this isn't
// meant to be instantiated directly - a concrete leaf type (OminousFloorBanner,
// OminousWallBanner) must embed it, implement Clone, and satisfy bannerShaper.
//
// WriteStateToWorld's tile sync (forcing the tile.Banner to White/no-patterns/Ominous on
// placement) is skipped: there's no WriteStateToWorld hook in Behavior yet, same documented gap
// as BaseBanner/block.Note/RedstoneComparator.
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

// AsItem should return VanillaItems.OMINOUS_BANNER() — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.
