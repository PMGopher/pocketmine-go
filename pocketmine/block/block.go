package block

import (
	"fmt"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	InternalStateDataBits = 11
	InternalStateDataMask = ^(^0 << InternalStateDataBits)
)

// Block is a port of pocketmine\block\Block.
//
// See Behavior's doc comment for why this holds a `self Behavior` field: any concrete block type
// embedding Block must call Init(self) — passing itself — right after constructing its own
// fields, mirroring where PHP's constructor computes state bits and snapshots defaultState.
//
// The state-ID XOR mask PHP computes via an xxh3 hash exists purely to improve PHP array
// performance (PHP arrays use some low bits of an integer key directly rather than fully hashing
// it, so spreading those bits out helps). Go's map[int]V always fully hashes its key regardless
// of bit distribution, so that optimization has no equivalent purpose here — this just uses the
// type ID shifted into the high bits, with no hash scrambling. EMPTY_STATE_ID is computed from
// this simplified formula rather than reproducing the hash-derived PHP constant, since it's
// documented as internal/unstable in both versions.
type Block struct {
	idInfo       *BlockIdentifier
	fallbackName string
	typeInfo     *BlockTypeInfo
	position     Position

	collisionBoxes     []math.AxisAlignedBB
	haveCollisionBoxes bool

	requiredBlockItemStateBits int
	requiredBlockOnlyStateBits int
	stateIDXorMask             int

	self         Behavior
	defaultState Behavior
}

// NewBlock constructs the Block base. Concrete types must call Init(self) afterwards.
func NewBlock(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) Block {
	return Block{idInfo: idInfo, fallbackName: name, typeInfo: typeInfo}
}

// Init finishes constructing b, given self (the concrete block type embedding this Block). Must
// be called exactly once, immediately after the concrete type's own fields are set to their
// initial/default values — it snapshots those values as the block's defaultState.
func (b *Block) Init(self Behavior) {
	b.self = self

	calc := runtime.NewSizeCalculator()
	self.DescribeBlockItemState(calc)
	b.requiredBlockItemStateBits = calc.GetBitsUsed()

	calc = runtime.NewSizeCalculator()
	self.DescribeBlockOnlyState(calc)
	b.requiredBlockOnlyStateBits = calc.GetBitsUsed()

	b.stateIDXorMask = b.idInfo.GetBlockTypeID() << InternalStateDataBits

	// must be last: Clone() copies self's current (default) field values
	b.defaultState = self.Clone()
}

func (b *Block) GetIdInfo() *BlockIdentifier { return b.idInfo }

func (b *Block) GetName() string { return b.fallbackName }

func (b *Block) GetTypeId() int { return b.idInfo.GetBlockTypeID() }

func (b *Block) GetStateId() int {
	full, err := b.encodeFullState()
	if err != nil {
		panic(err)
	}
	return full ^ b.stateIDXorMask
}

func (b *Block) HasSameTypeId(other Behavior) bool { return b.GetTypeId() == other.GetTypeId() }

func (b *Block) IsSameState(other Behavior) bool { return b.GetStateId() == other.GetStateId() }

func (b *Block) GetTypeTags() []string { return b.typeInfo.GetTypeTags() }

func (b *Block) HasTypeTag(tag string) bool { return b.typeInfo.HasTypeTag(tag) }

// rebind repoints self after a raw struct copy (`c := *t`) — used by each concrete block type's
// Clone() implementation, since copying the struct copies the *old* self pointer along with it.
// Unexported: only callable from within this package, which is where every concrete block type
// lives (matching PHP's flat pocketmine\block namespace), so each type's one-line Clone() can
// call it directly.
func (b *Block) rebind(self Behavior) { b.self = self }

// DecodeBlockItemState/DecodeBlockOnlyState mirror the private PHP methods of the same names.
// Exported (see Behavior's doc comment) purely so AsItem() can call them through the interface.

func (b *Block) DecodeBlockItemState(data int) error {
	r := runtime.NewReader(b.requiredBlockItemStateBits, data)
	b.self.DescribeBlockItemState(r)
	if readBits := r.GetOffset(); readBits != b.requiredBlockItemStateBits {
		return fmt.Errorf("exactly %d bits of block-item state data were provided, but %d were read", b.requiredBlockItemStateBits, readBits)
	}
	return nil
}

func (b *Block) DecodeBlockOnlyState(data int) error {
	r := runtime.NewReader(b.requiredBlockOnlyStateBits, data)
	b.self.DescribeBlockOnlyState(r)
	if readBits := r.GetOffset(); readBits != b.requiredBlockOnlyStateBits {
		return fmt.Errorf("exactly %d bits of block-only state data were provided, but %d were read", b.requiredBlockOnlyStateBits, readBits)
	}
	return nil
}

func (b *Block) encodeBlockItemState() (int, error) {
	w := runtime.NewWriter(b.requiredBlockItemStateBits)
	b.self.DescribeBlockItemState(w)
	if writtenBits := w.GetOffset(); writtenBits != b.requiredBlockItemStateBits {
		return 0, fmt.Errorf("exactly %d bits of block-item state data were expected, but %d were written", b.requiredBlockItemStateBits, writtenBits)
	}
	return w.GetValue(), nil
}

func (b *Block) encodeBlockOnlyState() (int, error) {
	w := runtime.NewWriter(b.requiredBlockOnlyStateBits)
	b.self.DescribeBlockOnlyState(w)
	if writtenBits := w.GetOffset(); writtenBits != b.requiredBlockOnlyStateBits {
		return 0, fmt.Errorf("exactly %d bits of block-only state data were expected, but %d were written", b.requiredBlockOnlyStateBits, writtenBits)
	}
	return w.GetValue(), nil
}

func (b *Block) encodeFullState() (int, error) {
	if b.requiredBlockOnlyStateBits == 0 && b.requiredBlockItemStateBits == 0 {
		return 0, nil
	}
	result := 0
	if b.requiredBlockItemStateBits > 0 {
		v, err := b.encodeBlockItemState()
		if err != nil {
			return 0, err
		}
		result |= v
	}
	if b.requiredBlockOnlyStateBits > 0 {
		v, err := b.encodeBlockOnlyState()
		if err != nil {
			return 0, err
		}
		result |= v << b.requiredBlockItemStateBits
	}
	return result, nil
}

// AsItem is a port of Block::asItem: block-only state (facing, powered, etc.) is discarded,
// block-item state (colour, wood type, etc.) is preserved, and the result is wrapped in an
// ItemBlock via NewItemBlockFunc (see item.go's doc comment on why that's a factory var rather
// than a direct call - block can't import the item package that provides it).
//
// Errors if NewItemBlockFunc hasn't been set, i.e. the running program never imported the item
// package - every real program does (it's how blocks and items get wired together at all), so
// this only bites tests that exercise AsItem() paths without importing item too.
func (b *Block) AsItem() (Item, error) {
	normalized := b.defaultState.Clone()
	itemData, err := b.encodeBlockItemState()
	if err != nil {
		return nil, err
	}
	if err := normalized.DecodeBlockItemState(itemData); err != nil {
		return nil, err
	}
	if NewItemBlockFunc == nil {
		return nil, fmt.Errorf("block: AsItem called but NewItemBlockFunc is nil (the item package was never imported)")
	}
	return NewItemBlockFunc(normalized), nil
}

// GetSide returns the Block on the given side of this one, step tiles away. Panics if this
// block's position has no valid world (mirroring PHP's \LogicException, a programmer error at
// the call site).
func (b *Block) GetSide(side math.Facing, step int) Behavior {
	if !b.position.IsValid() {
		panic("Block does not have a valid world")
	}
	offset, ok := math.FacingOffset[side]
	if !ok {
		offset = [3]int{0, 0, 0}
	}
	world, err := b.position.GetWorld()
	if err != nil {
		panic(err)
	}
	return world.GetBlockAt(
		int(b.position.X)+offset[0]*step,
		int(b.position.Y)+offset[1]*step,
		int(b.position.Z)+offset[2]*step,
	)
}

// GetAllSides returns the six blocks adjacent to this one.
func (b *Block) GetAllSides() []Behavior {
	sides := make([]Behavior, 0, len(math.AllFacing))
	for _, facing := range math.AllFacing {
		sides = append(sides, b.GetSide(facing, 1))
	}
	return sides
}

// GetAdjacentSupportType returns the type of support the block on the given side provides back
// towards this block.
func (b *Block) GetAdjacentSupportType(facing math.Facing) blockutils.SupportType {
	return b.GetSide(facing, 1).GetSupportType(math.Opposite(facing))
}

// blockGeometry captures *Block's GetSide and GetAdjacentSupportType — neither is part of
// Behavior because neither is ever overridden in the original (see their doc comments above) —
// so code that only holds a Behavior value (e.g. a concrete block type's canBeSupportedAt helper,
// the Go equivalent of PHP's StaticSupportTrait/checks scattered across many block types) can
// still reach them via a type assertion. Every concrete block type satisfies this automatically,
// since both methods promote from the embedded Block.
type blockGeometry interface {
	GetSide(side math.Facing, step int) Behavior
	GetAdjacentSupportType(facing math.Facing) blockutils.SupportType
	HasTypeTag(tag string) bool
	HasSameTypeId(other Behavior) bool
	IsFullCube() bool
}

func (b *Block) GetPosition() Position { return b.position }

func (b *Block) ReadStateFromWorld() Behavior { return b.self }

// SetPosition is the Go equivalent of the internal position() method.
func (b *Block) SetPosition(world World, x, y, z int) {
	b.position = NewPosition(float64(x), float64(y), float64(z), world)
	b.collisionBoxes = nil
	b.haveCollisionBoxes = false
}

func (b *Block) String() string {
	full, _ := b.encodeFullState()
	return fmt.Sprintf("Block[%s] (%d:%d)", b.fallbackName, b.GetTypeId(), full)
}

// --- default Behavior implementations ---

func (b *Block) DescribeBlockItemState(w runtime.DataDescriber) {}
func (b *Block) DescribeBlockOnlyState(w runtime.DataDescriber) {}

func (b *Block) CanBePlaced() bool   { return true }
func (b *Block) CanBeReplaced() bool { return false }

func (b *Block) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return blockReplace.CanBeReplaced()
}

func (b *Block) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	tx.AddBlock(blockReplace.GetPosition(), b.self)
	return true
}

func (b *Block) OnPostPlace() {}

func (b *Block) GetBreakInfo() *BlockBreakInfo { return b.typeInfo.GetBreakInfo() }

func (b *Block) GetEnchantmentTags() []string { return b.typeInfo.GetEnchantmentTags() }

// OnBreak is a simplified port of Block::onBreak(): the real version destroys any tile at this
// position and replaces the block with air. Tile is currently just a marker interface (see
// world.go) and there's no ported "air block" registry yet to replace this position's block
// with, so both of those steps are deferred — this is a placeholder that always succeeds.
func (b *Block) OnBreak(item Item, player Player, returnedItems *[]Item) bool {
	return true
}

func (b *Block) OnNearbyBlockChange() {}

func (b *Block) TicksRandomly() bool { return false }
func (b *Block) OnRandomTick()       {}
func (b *Block) OnScheduledUpdate()  {}

func (b *Block) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return false
}

func (b *Block) OnAttack(item Item, face math.Facing, player Player) bool { return false }

func (b *Block) GetFrictionFactor() float64 { return 0.6 }

func (b *Block) GetLightLevel() int { return 0 }

func (b *Block) GetLightFilter() int {
	if b.self.IsTransparent() {
		return 0
	}
	return 15
}

func (b *Block) BlocksDirectSkyLight() bool { return b.self.GetLightFilter() > 0 }

func (b *Block) IsTransparent() bool { return false }

// IsSolid is deliberately confusing/inconsistent in the original too — see the PHP doc comment.
// "Solid" here doesn't mean "has a collision box"; things like signs count as solid, and things
// like skulls don't.
func (b *Block) IsSolid() bool { return true }

func (b *Block) CanBeFlowedInto() bool { return false }
func (b *Block) CanClimb() bool        { return false }

// GetDrops mirrors Block::getDrops(). The PHP original also checks
// `$item->hasEnchantment(VanillaEnchantments::SILK_TOUCH())` before taking the silk-touch branch;
// Item doesn't expose enchantment queries yet (the enchantment package isn't ported), so the
// silk-touch branch is never taken here until that's wired up.
func (b *Block) GetDrops(item Item) []Item {
	if b.self.GetBreakInfo().IsToolCompatible(item) {
		return b.self.GetDropsForCompatibleTool(item)
	}
	return b.self.GetDropsForIncompatibleTool(item)
}

// GetDropsForCompatibleTool/GetSilkTouchDrops are ports of Block's defaults, both matching PHP's
// `return [$this->asItem()];`. AsItem() errors when NewItemBlockFunc is nil (the item package
// isn't linked into the running program - notably, block's own test suite never can: item already
// imports block, so block's internal tests importing item back would be a real Go import cycle);
// on error these just fall back to no drops, exactly like their previous permanent stub did,
// rather than panicking over a condition that's normal for those tests.
func (b *Block) GetDropsForCompatibleTool(item Item) []Item {
	asItem, err := b.AsItem()
	if err != nil {
		return nil
	}
	return []Item{asItem}
}

func (b *Block) GetDropsForIncompatibleTool(item Item) []Item { return nil }

func (b *Block) GetSilkTouchDrops(item Item) []Item {
	asItem, err := b.AsItem()
	if err != nil {
		return nil
	}
	return []Item{asItem}
}

func (b *Block) GetXpDropAmount() int { return 0 }

func (b *Block) IsAffectedBySilkTouch() bool { return false }

// GetPickedItem is a simplified port of Block::getPickedItem(): the addUserData branch (copying a
// tile's cleaned NBT onto the item) is skipped, since the Tile marker interface doesn't expose
// GetCleanedNBT yet - same gap category as everywhere else Tile is too narrow. See
// GetDropsForCompatibleTool's doc comment for why AsItem() failing falls back to nil rather than
// panicking.
func (b *Block) GetPickedItem(addUserData bool) Item {
	asItem, err := b.AsItem()
	if err != nil {
		return nil
	}
	return asItem
}

func (b *Block) GetFuelTime() int           { return 0 }
func (b *Block) GetMaxStackSize() int       { return 64 }
func (b *Block) IsFireProofAsItem() bool    { return false }
func (b *Block) GetFlameEncouragement() int { return 0 }
func (b *Block) GetFlammability() int       { return 0 }
func (b *Block) BurnsForever() bool         { return false }
func (b *Block) IsFlammable() bool          { return b.self.GetFlammability() > 0 }
func (b *Block) OnIncinerate()              {}

func (b *Block) GetModelPositionOffset() (math.Vector3, bool) { return math.Vector3{}, false }

func (b *Block) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB()}
}

func (b *Block) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeFull
}

func (b *Block) GetAffectedBlocks() []Behavior { return []Behavior{b.self} }

func (b *Block) HasEntityCollision() bool { return false }

func (b *Block) OnEntityInside(entity Entity) bool { return true }

func (b *Block) AddVelocityToEntity(entity Entity) (math.Vector3, bool) { return math.Vector3{}, false }

func (b *Block) OnEntityLand(entity Entity) (float64, bool) { return 0, false }

func (b *Block) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {}

// --- collision/geometry, not part of Behavior (not overridden — they build on RecalculateCollisionBoxes/GetModelPositionOffset instead) ---

func (b *Block) CollidesWithBB(bb math.AxisAlignedBB) bool {
	for _, bb2 := range b.GetCollisionBoxes() {
		if bb.IntersectsWith(bb2, 0.00001) {
			return true
		}
	}
	return false
}

func (b *Block) GetCollisionBoxes() []math.AxisAlignedBB {
	if !b.haveCollisionBoxes {
		b.collisionBoxes = b.self.RecalculateCollisionBoxes()
		offset := b.position.Vector3
		if extra, ok := b.self.GetModelPositionOffset(); ok {
			offset = offset.AddVector(extra)
		}
		for i, bb := range b.collisionBoxes {
			b.collisionBoxes[i] = bb.OffsetCopy(offset.X, offset.Y, offset.Z)
		}
		b.haveCollisionBoxes = true
	}
	return b.collisionBoxes
}

func (b *Block) IsFullCube() bool {
	bb := b.GetCollisionBoxes()
	return len(bb) == 1 && bb[0].GetAverageEdgeLength() >= 1 && bb[0].IsCube(0.000001)
}

func (b *Block) CalculateIntercept(pos1, pos2 math.Vector3) (math.RayTraceResult, bool) {
	bbs := b.GetCollisionBoxes()
	if len(bbs) == 0 {
		return math.RayTraceResult{}, false
	}

	found := false
	var currentHit math.RayTraceResult
	currentDistance := 0.0

	for _, bb := range bbs {
		nextHit, ok := bb.CalculateIntercept(pos1, pos2)
		if !ok {
			continue
		}
		nextDistance := nextHit.HitVector.DistanceSquared(pos1)
		if !found || nextDistance < currentDistance {
			currentHit = nextHit
			currentDistance = nextDistance
			found = true
		}
	}

	return currentHit, found
}
