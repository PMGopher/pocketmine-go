package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Behavior is a port of every overridable (public/protected non-final) method on
// pocketmine\block\Block.
//
// PHP's Block base class calls many of these on `$this` from within its OWN default
// implementations — e.g. GetDrops() calls GetBreakInfo(), IsAffectedBySilkTouch(),
// GetSilkTouchDrops(), GetDropsForCompatibleTool()/GetDropsForIncompatibleTool(). If a concrete
// Go block type just embedded *Block, those internal calls would resolve to *Block's own
// (default) methods, not the concrete type's overrides — the same bug class as the
// SimpleLogger/PrefixedLogger embedding pitfall documented in the log package, but here it would
// silently break customization for any of ~15 default methods across every one of the ~270
// concrete block types instead of one logger.
//
// The fix is this interface plus Block.self (set via Block.Init at construction): concrete block
// types embed *Block AND call Init(self) with themselves, and Block's default method bodies that
// need to call another overridable method go through b.self.X() instead of calling their own
// method directly. External callers (World, Player, commands) interact with blocks entirely
// through this interface too, so they always see the right override regardless.
type Behavior interface {
	// Clone returns a deep copy of the concrete block, as a Behavior. Typically a one-liner
	// (`c := *t; return &c`) since Go's struct value copy already does the deep-enough copy for
	// the plain-old-data fields concrete block types have — see Block.Init's doc comment for why
	// this is needed (a defaultState snapshot of the *concrete* type, not just the Block base).
	Clone() Behavior

	GetTypeId() int
	GetStateId() int
	GetPosition() Position

	// State encoding — must always describe the same fields in the same order regardless of
	// current state (see data/runtime's DataDescriber for why these take pointers).
	DescribeBlockItemState(w runtime.DataDescriber)
	DescribeBlockOnlyState(w runtime.DataDescriber)
	// DecodeBlockItemState/DecodeBlockOnlyState are promoted from Block automatically (no
	// per-concrete-type implementation needed) — they're part of Behavior purely so AsItem() can
	// call them on an arbitrary cloned Behavior value.
	DecodeBlockItemState(data int) error
	DecodeBlockOnlyState(data int) error

	CanBePlaced() bool
	CanBeReplaced() bool
	CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool
	Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool
	OnPostPlace()

	GetBreakInfo() *BlockBreakInfo
	GetEnchantmentTags() []string
	OnBreak(item Item, player Player, returnedItems *[]Item) bool
	OnNearbyBlockChange()
	TicksRandomly() bool
	OnRandomTick()
	OnScheduledUpdate()
	OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool
	OnAttack(item Item, face math.Facing, player Player) bool

	GetFrictionFactor() float64
	GetLightLevel() int
	GetLightFilter() int
	BlocksDirectSkyLight() bool
	IsTransparent() bool
	IsSolid() bool
	CanBeFlowedInto() bool
	CanClimb() bool

	GetDrops(item Item) []Item
	GetDropsForCompatibleTool(item Item) []Item
	GetDropsForIncompatibleTool(item Item) []Item
	GetSilkTouchDrops(item Item) []Item
	GetXpDropAmount() int
	IsAffectedBySilkTouch() bool
	GetPickedItem(addUserData bool) Item

	GetFuelTime() int
	GetMaxStackSize() int
	IsFireProofAsItem() bool
	GetFlameEncouragement() int
	GetFlammability() int
	BurnsForever() bool
	IsFlammable() bool
	OnIncinerate()

	GetModelPositionOffset() (math.Vector3, bool)
	RecalculateCollisionBoxes() []math.AxisAlignedBB
	GetSupportType(facing math.Facing) blockutils.SupportType

	GetAffectedBlocks() []Behavior

	HasEntityCollision() bool
	OnEntityInside(entity Entity) bool
	AddVelocityToEntity(entity Entity) (math.Vector3, bool)
	OnEntityLand(entity Entity) (float64, bool)
	OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult)
}
