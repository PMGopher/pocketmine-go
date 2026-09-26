package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
	"pocketmine-go/pocketmine/math"
)

// dripleafShaper lets concrete leaf types (BigDripleafStem, BigDripleafHead) report whether they
// are the head half, both for BaseBigDripleaf's own self-dispatched isHead() calls and for
// recognising a neighbouring dripleaf block via type assertion (PHP's `instanceof
// BaseBigDripleaf`) - same self-dispatch shape as RailShaper/pressurePlateShaper/candleShaper/
// stemShaper.
type dripleafShaper interface {
	IsHead() bool
	GetFacing() math.Facing
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

// Place is a port of BaseBigDripleaf::place.
func (b *BaseBigDripleaf) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	below := blockReplace.(blockGeometry).GetSide(math.Down, 1)
	if !bigDripleafCanBeSupportedBy(below, true) {
		return false
	}
	if player != nil {
		b.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	if belowDripleaf, ok := below.(dripleafShaper); ok {
		b.Facing = belowDripleaf.GetFacing()
		stem := VanillaBigDripleafStem().(*BigDripleafStem)
		stem.SetFacing(b.Facing)
		tx.AddBlock(below.GetPosition(), stem)
	}
	return b.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract is a port of BaseBigDripleaf::onInteract.
func (b *BaseBigDripleaf) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if isFertilizer(item) && b.grow(player) {
		item.Pop()
		return true
	}
	return false
}

// seekToHead is a port of BaseBigDripleaf::seekToHead (nil if there's no head above).
func (b *BaseBigDripleaf) seekToHead() Behavior {
	if b.self.(dripleafShaper).IsHead() {
		return b.self
	}
	for step := 1; ; step++ {
		next := b.self.(blockGeometry).GetSide(math.Up, step)
		shaper, ok := next.(dripleafShaper)
		if !ok {
			return nil
		}
		if shaper.IsHead() {
			return next
		}
	}
}

// grow is a port of BaseBigDripleaf::grow.
func (b *BaseBigDripleaf) grow(player Player) bool {
	head := b.seekToHead()
	if head == nil {
		return false
	}
	pos := head.GetPosition()
	world, err := pos.GetWorld()
	if err != nil {
		return false
	}
	up := pos.GetSide(math.Up, 1)
	if !world.IsInWorld(up.FloorX(), up.FloorY(), up.FloorZ()) || world.GetBlockAt(up.FloorX(), up.FloorY(), up.FloorZ()).GetTypeId() != AIR {
		return false
	}
	facing := head.(interface{ GetFacing() math.Facing }).GetFacing()
	stem := VanillaBigDripleafStem().(*BigDripleafStem)
	stem.SetFacing(facing)
	newHead := VanillaBigDripleafHead().(*BigDripleafHead)
	newHead.SetFacing(facing)
	tx := NewBlockTransaction(world)
	tx.AddBlock(pos, stem)
	tx.AddBlock(up, newHead)
	ev := blockevent.NewStructureGrowEvent(head, tx, eventPlayer(player))
	event.Call(ev)
	if !ev.IsCancelled() {
		return tx.Apply()
	}
	return false
}

func (b *BaseBigDripleaf) GetFlameEncouragement() int { return 15 }

func (b *BaseBigDripleaf) GetFlammability() int { return 100 }

func (b *BaseBigDripleaf) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}
