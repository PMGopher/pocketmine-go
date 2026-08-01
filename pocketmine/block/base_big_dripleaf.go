package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// dripleafShaper lets concrete leaf types (BigDripleafStem, BigDripleafHead) report whether they
// are the head half, both for BaseBigDripleaf's own self-dispatched isHead() calls and for
// recognising a neighbouring dripleaf block via type assertion (PHP's `instanceof
// BaseBigDripleaf`) - same self-dispatch shape as RailShaper/pressurePlateShaper/candleShaper/
// stemShaper.
type dripleafShaper interface {
	IsHead() bool
}

// BaseBigDripleaf is a port of pocketmine\block\BaseBigDripleaf. Like Crops/Stem, this isn't
// meant to be instantiated directly - a concrete leaf type (BigDripleafStem, BigDripleafHead)
// must embed it, implement Clone, and satisfy dripleafShaper.
type BaseBigDripleaf struct {
	Transparent
	HorizontalFacingComponent
}

func (b *BaseBigDripleaf) DescribeBlockOnlyState(w runtime.DataDescriber) {
	b.DescribeHorizontalFacing(w)
}

// bigDripleafCanBeSupportedBy is a port of BaseBigDripleaf::canBeSupportedBy. Moss block support
// is a TODO in the PHP original too.
func bigDripleafCanBeSupportedBy(blk Behavior, head bool) bool {
	if shaper, ok := blk.(dripleafShaper); ok && shaper.IsHead() == head {
		return true
	}
	if blk.GetTypeId() == CLAY {
		return true
	}
	if geo, ok := blk.(blockGeometry); ok {
		return geo.HasTypeTag(BlockTypeTagsDirt) || geo.HasTypeTag(BlockTypeTagsMud)
	}
	return false
}

func (b *BaseBigDripleaf) OnNearbyBlockChange() {
	geo := b.self.(blockGeometry)
	isHead := b.self.(dripleafShaper).IsHead()
	if (!isHead && !isBaseBigDripleaf(geo.GetSide(math.Up, 1))) || !bigDripleafCanBeSupportedBy(geo.GetSide(math.Down, 1), false) {
		if world, err := b.position.GetWorld(); err == nil {
			world.UseBreakOn(b.position.AsVector3())
		}
	}
}

func isBaseBigDripleaf(blk Behavior) bool {
	_, ok := blk.(dripleafShaper)
	return ok
}

// Place is a port of BaseBigDripleaf::place. The PHP original also converts the block below (if
// it's another BaseBigDripleaf) into a BIG_DRIPLEAF_STEM via VanillaBlocks - that part needs the
// unported block registry, so it's skipped here (documented gap, same category as everywhere else
// VanillaBlocks is needed).
func (b *BaseBigDripleaf) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	below := blockReplace.(blockGeometry).GetSide(math.Down, 1)
	if !bigDripleafCanBeSupportedBy(below, true) {
		return false
	}
	if player != nil {
		b.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	// If `below` is itself a BaseBigDripleaf, the PHP original also converts it into a
	// BIG_DRIPLEAF_STEM with this facing via VanillaBlocks - skipped, see this method's doc
	// comment.
	return b.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract should fertilize-grow the dripleaf stack (BaseBigDripleaf::grow) - needs a
// Fertilizer item marker, StructureGrowEvent, World.IsInWorld/GetBlock(Position), and the block
// registry, none ported yet, so this is a no-op for now.
func (b *BaseBigDripleaf) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return false
}

func (b *BaseBigDripleaf) GetFlameEncouragement() int { return 15 }

func (b *BaseBigDripleaf) GetFlammability() int { return 100 }

func (b *BaseBigDripleaf) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
