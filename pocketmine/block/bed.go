package block

import (
	"pocketmine-go/pocketmine/block/tile"
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Bed is a port of pocketmine\block\Bed.
type Bed struct {
	Transparent
	ColorComponent
	HorizontalFacingComponent

	Occupied bool
	Head     bool
}

func NewBed(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Bed {
	b := &Bed{
		Transparent:               Transparent{NewBlock(idInfo, name, typeInfo)},
		ColorComponent:            NewColorComponent(),
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	b.Init(b)
	return b
}

func (b *Bed) Clone() Behavior {
	c := *b
	c.rebind(&c)
	return &c
}

func (b *Bed) DescribeBlockOnlyState(w runtime.DataDescriber) {
	b.DescribeHorizontalFacing(w)
	w.Bool(&b.Occupied)
	w.Bool(&b.Head)
}

// ReadStateFromWorld is a port of Bed::readStateFromWorld.
func (b *Bed) ReadStateFromWorld() Behavior {
	b.Block.ReadStateFromWorld()

	b.Color = blockutils.DyeColorRed // legacy pre-1.1 beds don't have tiles
	world, err := b.position.GetWorld()
	if err != nil {
		return b.self
	}
	t, _ := world.GetTile(b.position)
	if bedTile, ok := t.(*tile.Bed); ok {
		b.Color = bedTile.GetColor()
	}
	return b.self
}

// WriteStateToWorld's tile sync (writing Color back to the tile.Bed on placement) is skipped:
// there's no WriteStateToWorld hook in Behavior yet - same documented gap as
// Note/RedstoneComparator/BaseBanner/MobHead.

func (b *Bed) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 7.0/16)}
}

func (b *Bed) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (b *Bed) IsHeadPart() bool { return b.Head }

func (b *Bed) SetHead(head bool) { b.Head = head }

func (b *Bed) IsOccupied() bool { return b.Occupied }

func (b *Bed) SetOccupied(occupied bool) { b.Occupied = occupied }

func (b *Bed) getOtherHalfSide() math.Facing {
	if b.Head {
		return math.Opposite(b.Facing)
	}
	return b.Facing
}

// GetOtherHalf is a port of Bed::getOtherHalf.
func (b *Bed) GetOtherHalf() (*Bed, bool) {
	other, ok := b.self.(blockGeometry).GetSide(b.getOtherHalfSide(), 1).(*Bed)
	if !ok || other.Head == b.Head || other.Facing != b.Facing {
		return nil, false
	}
	return other, true
}

// OnInteract should send the player a status message and put them to sleep - needs
// Player.SendMessage/SleepOn, World.GetTimeOfDay, and the lang/KnownTranslationFactory machinery
// wired to a real Player, none ported to that depth yet, so this is a documented no-op for now
// (still returns true when a player is present, matching the PHP original's control flow).
func (b *Bed) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return player != nil
}

func (b *Bed) OnNearbyBlockChange() {
	if b.Head {
		return
	}
	other, ok := b.GetOtherHalf()
	if !ok || other.Occupied == b.Occupied {
		return
	}
	b.Occupied = other.Occupied
	if world, err := b.position.GetWorld(); err == nil {
		if err := world.SetBlock(b.position, b.self); err != nil {
			panic(err)
		}
	}
}

func (b *Bed) OnEntityLand(entity Entity) (float64, bool) {
	if living, ok := entity.(Living); ok && living.IsSneaking() {
		return 0, false
	}
	entity.SetFallDistance(entity.GetFallDistance() * 0.5)
	return entity.GetMotion().Y * -3 / 4, true // 2/3 in Java, according to the wiki
}

func (b *Bed) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Down) != blockutils.SupportTypeNone
}

func (b *Bed) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if !b.canBeSupportedAt(blockReplace) {
		return false
	}
	if player != nil {
		b.Facing = player.GetHorizontalFacing()
	} else {
		b.Facing = math.North
	}

	next := b.self.(blockGeometry).GetSide(b.getOtherHalfSide(), 1)
	if !next.CanBeReplaced() || !b.canBeSupportedAt(next) {
		return false
	}
	nextState := b.self.Clone().(*Bed)
	nextState.Head = true
	tx.AddBlock(blockReplace.GetPosition(), b.self)
	tx.AddBlock(next.GetPosition(), nextState)
	return true
}

func (b *Bed) GetDrops(item Item) []Item {
	if b.Head {
		return b.Block.GetDrops(item)
	}
	return nil
}

func (b *Bed) GetAffectedBlocks() []Behavior {
	if other, ok := b.GetOtherHalf(); ok {
		return []Behavior{b.self, other}
	}
	return b.Block.GetAffectedBlocks()
}

func (b *Bed) GetMaxStackSize() int { return 1 }
