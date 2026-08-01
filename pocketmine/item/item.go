package item

import (
	"encoding/base64"
	"fmt"

	"pocketmine-go/pocketmine/block"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/nbt"
)

// ItemIdentifier is a port of pocketmine\item\ItemIdentifier. The PHP FromBlock() named
// constructor (needing ItemTypeIds::fromBlockTypeId, a registry lookup) isn't ported - nothing
// constructs an ItemIdentifier from a block yet.
type ItemIdentifier struct {
	TypeId int
}

func NewItemIdentifier(typeId int) ItemIdentifier { return ItemIdentifier{TypeId: typeId} }

// ItemUseResult is a port of pocketmine\item\ItemUseResult.
type ItemUseResult int

const (
	ItemUseResultNone ItemUseResult = iota
	ItemUseResultFail
	ItemUseResultSuccess
)

const (
	tagDisplay     = "display"
	tagBlockEntity = "BlockEntityTag"
	tagDisplayName = "Name"
	tagDisplayLore = "Lore"
	tagKeepOnDeath = "minecraft:keep_on_death"
	tagCanPlaceOn  = "CanPlaceOn"
	tagCanDestroy  = "CanDestroy"
)

// Item is a port of pocketmine\item\Item's public surface, using the same self-dispatch pattern
// as block.Behavior/block.Block: concrete leaf types embed ItemBase, call Init(self) once their
// own fields are set to their defaults, and override whichever methods they need.
//
// Not ported (all documented on ItemBase below rather than declared here, since nothing needs
// them yet): enchantment handling (ItemEnchantmentHandlingTrait - needs the unported
// item/enchantment package), NbtSerialize/NbtDeserialize/SafeNbtDeserialize (need
// GlobalItemDataHandlers, a whole item-data-driven serializer/deserializer registry),
// legacyJsonDeserialize (deprecated upgrade path, not worth porting), GetPlacementTransaction
// (needs a concrete world.BlockTransaction and the block registry).
type Item interface {
	Clone() Item

	GetTypeId() int
	GetStateId() int
	GetName() string
	GetVanillaName() string
	HasCustomName() bool
	GetCustomName() string
	SetCustomName(name string)
	ClearCustomName()

	GetLore() []string
	SetLore(lines []string)

	HasCustomBlockData() bool
	GetCustomBlockData() *nbt.CompoundTag
	SetCustomBlockData(compound *nbt.CompoundTag)
	ClearCustomBlockData()

	GetCanPlaceOn() map[string]string
	SetCanPlaceOn(values []string)
	GetCanDestroy() map[string]string
	SetCanDestroy(values []string)

	KeepOnDeath() bool
	SetKeepOnDeath(keep bool)

	HasNamedTag() bool
	GetNamedTag() *nbt.CompoundTag
	SetNamedTag(tag *nbt.CompoundTag)
	ClearNamedTag()

	GetCount() int
	SetCount(count int)
	// Pop reduces this item stack's count by one, discarding the split-off single-item stack -
	// matches block.Item's Pop() exactly (see that interface's doc comment), so any concrete
	// item type here is automatically usable wherever block.Item is expected.
	Pop()
	// PopCount is the full port of Item::pop($count): it splits off and returns a clone holding
	// count items, reducing this stack's own count by the same amount. Pop() above is the
	// count=1 case with the split-off stack discarded.
	PopCount(count int) Item
	IsNull() bool

	GetEnchantability() int

	GetMaxStackSize() int
	GetFuelTime() int
	GetFuelResidue() Item
	IsFireProof() bool
	GetAttackPoints() int
	GetDefensePoints() int

	GetBlockToolType() block.ToolType
	GetBlockToolHarvestLevel() int
	GetMiningEfficiency(isCorrectTool bool) float64

	GetCooldownTicks() int
	GetCooldownTag() (string, bool)

	Equals(other Item, checkCompound bool) bool
	CanStackWith(other Item) bool
	EqualsExact(other Item) bool

	String() string
}

// ItemBase is a port of pocketmine\item\Item's state and default method implementations. See
// describeState's doc comment for how concrete types add runtime state.
type ItemBase struct {
	self Item

	identifier ItemIdentifier
	name       string
	count      int

	customName     string
	lore           []string
	blockEntityTag *nbt.CompoundTag
	canPlaceOn     map[string]string
	canDestroy     map[string]string
	keepOnDeath    bool

	nbtTag *nbt.CompoundTag
}

// Init finishes constructing b, given self (the concrete item type embedding this ItemBase).
// Must be called exactly once, immediately after the concrete type's own fields are set to their
// initial/default values.
func (b *ItemBase) Init(self Item, identifier ItemIdentifier, name string) {
	b.self = self
	b.identifier = identifier
	b.name = name
	b.count = 1
	b.nbtTag = nbt.NewCompoundTag()
}

// rebind repoints b.self after a concrete type has been copied (e.g. in Clone) - same pattern as
// block.Block.rebind.
func (b *ItemBase) rebind(self Item) { b.self = self }

func (b *ItemBase) HasCustomBlockData() bool { return b.blockEntityTag != nil }

func (b *ItemBase) ClearCustomBlockData() { b.blockEntityTag = nil }

func (b *ItemBase) SetCustomBlockData(compound *nbt.CompoundTag) { b.blockEntityTag = compound.Clone() }

func (b *ItemBase) GetCustomBlockData() *nbt.CompoundTag { return b.blockEntityTag }

func (b *ItemBase) HasCustomName() bool { return b.customName != "" }

func (b *ItemBase) GetCustomName() string { return b.customName }

func (b *ItemBase) SetCustomName(name string) { b.customName = name }

func (b *ItemBase) ClearCustomName() { b.customName = "" }

func (b *ItemBase) GetLore() []string { return b.lore }

func (b *ItemBase) SetLore(lines []string) { b.lore = lines }

func (b *ItemBase) GetCanPlaceOn() map[string]string { return b.canPlaceOn }

func (b *ItemBase) SetCanPlaceOn(values []string) {
	b.canPlaceOn = make(map[string]string, len(values))
	for _, v := range values {
		b.canPlaceOn[v] = v
	}
}

func (b *ItemBase) GetCanDestroy() map[string]string { return b.canDestroy }

func (b *ItemBase) SetCanDestroy(values []string) {
	b.canDestroy = make(map[string]string, len(values))
	for _, v := range values {
		b.canDestroy[v] = v
	}
}

func (b *ItemBase) KeepOnDeath() bool { return b.keepOnDeath }

func (b *ItemBase) SetKeepOnDeath(keep bool) { b.keepOnDeath = keep }

func (b *ItemBase) HasNamedTag() bool { return b.GetNamedTag().Count() > 0 }

// GetNamedTag is a port of Item::getNamedTag.
func (b *ItemBase) GetNamedTag() *nbt.CompoundTag {
	b.serializeCompoundTag(b.nbtTag)
	return b.nbtTag
}

// SetNamedTag is a port of Item::setNamedTag.
func (b *ItemBase) SetNamedTag(tag *nbt.CompoundTag) {
	if tag.Count() == 0 {
		b.ClearNamedTag()
		return
	}
	b.nbtTag = tag.Clone()
	b.deserializeCompoundTag(b.nbtTag)
}

// ClearNamedTag is a port of Item::clearNamedTag.
func (b *ItemBase) ClearNamedTag() {
	b.nbtTag = nbt.NewCompoundTag()
	b.deserializeCompoundTag(b.nbtTag)
}

// deserializeCompoundTag is a port of Item::deserializeCompoundTag, minus the "ench" list round
// trip - EnchantmentInstance/EnchantmentIdMap (item/enchantment package) aren't ported, so any
// "ench" tag on loaded NBT is silently ignored rather than populating enchantments (matching every
// other HasEnchantment-is-always-false assumption already made elsewhere in this port).
func (b *ItemBase) deserializeCompoundTag(tag *nbt.CompoundTag) {
	b.customName = ""
	b.lore = nil

	if display, ok, _ := tag.GetCompoundTag(tagDisplay); ok {
		if name, err := display.GetString(tagDisplayName); err == nil {
			b.customName = string(name)
		}
		if lore, ok, _ := display.GetListTag(tagDisplayLore); ok {
			for _, t := range lore.Values() {
				if s, ok := t.(nbt.StringTag); ok {
					b.lore = append(b.lore, string(s))
				}
			}
		}
	}

	b.blockEntityTag = nil
	if bet, ok, _ := tag.GetCompoundTag(tagBlockEntity); ok {
		b.blockEntityTag = bet
	}

	b.canPlaceOn = nil
	if list, ok, _ := tag.GetListTag(tagCanPlaceOn); ok {
		b.canPlaceOn = make(map[string]string, list.Count())
		for _, t := range list.Values() {
			if s, ok := t.(nbt.StringTag); ok {
				b.canPlaceOn[string(s)] = string(s)
			}
		}
	}

	b.canDestroy = nil
	if list, ok, _ := tag.GetListTag(tagCanDestroy); ok {
		b.canDestroy = make(map[string]string, list.Count())
		for _, t := range list.Values() {
			if s, ok := t.(nbt.StringTag); ok {
				b.canDestroy[string(s)] = string(s)
			}
		}
	}

	b.keepOnDeath = tag.GetByteOr(tagKeepOnDeath, 0) != 0
}

// serializeCompoundTag is a port of Item::serializeCompoundTag, minus the "ench" list (see
// deserializeCompoundTag's doc comment - enchantments are always empty here, so that branch is
// simply never taken).
func (b *ItemBase) serializeCompoundTag(tag *nbt.CompoundTag) {
	display, hasDisplay, _ := tag.GetCompoundTag(tagDisplay)

	if b.customName != "" {
		if !hasDisplay {
			display = nbt.NewCompoundTag()
			hasDisplay = true
		}
		display.SetString(tagDisplayName, nbt.StringTag(b.customName))
	} else if hasDisplay {
		display.RemoveTag(tagDisplayName)
	}

	if len(b.lore) > 0 {
		values := make([]nbt.Tag, len(b.lore))
		for i, line := range b.lore {
			values[i] = nbt.StringTag(line)
		}
		loreTag, err := nbt.NewListTag(values, nbt.TagString)
		if err != nil {
			panic(err)
		}
		if !hasDisplay {
			display = nbt.NewCompoundTag()
			hasDisplay = true
		}
		display.SetTag(tagDisplayLore, loreTag)
	} else if hasDisplay {
		display.RemoveTag(tagDisplayLore)
	}

	if hasDisplay && display.Count() > 0 {
		tag.SetTag(tagDisplay, display)
	} else {
		tag.RemoveTag(tagDisplay)
	}

	if b.blockEntityTag != nil {
		tag.SetTag(tagBlockEntity, b.blockEntityTag.Clone())
	} else {
		tag.RemoveTag(tagBlockEntity)
	}

	if len(b.canPlaceOn) > 0 {
		values := make([]nbt.Tag, 0, len(b.canPlaceOn))
		for _, v := range b.canPlaceOn {
			values = append(values, nbt.StringTag(v))
		}
		listTag, err := nbt.NewListTag(values, nbt.TagString)
		if err != nil {
			panic(err)
		}
		tag.SetTag(tagCanPlaceOn, listTag)
	} else {
		tag.RemoveTag(tagCanPlaceOn)
	}

	if len(b.canDestroy) > 0 {
		values := make([]nbt.Tag, 0, len(b.canDestroy))
		for _, v := range b.canDestroy {
			values = append(values, nbt.StringTag(v))
		}
		listTag, err := nbt.NewListTag(values, nbt.TagString)
		if err != nil {
			panic(err)
		}
		tag.SetTag(tagCanDestroy, listTag)
	} else {
		tag.RemoveTag(tagCanDestroy)
	}

	if b.keepOnDeath {
		tag.SetByte(tagKeepOnDeath, 1)
	} else {
		tag.RemoveTag(tagKeepOnDeath)
	}
}

func (b *ItemBase) GetCount() int { return b.count }

func (b *ItemBase) SetCount(count int) { b.count = count }

// PopCount is a port of Item::pop. It panics (matching the PHP original's
// InvalidArgumentException) if count exceeds the current stack.
func (b *ItemBase) PopCount(count int) Item {
	if count > b.count {
		panic("cannot pop more items than are on the stack")
	}
	popped := b.self.Clone()
	popped.SetCount(count)
	b.count -= count
	return popped
}

func (b *ItemBase) Pop() { b.self.PopCount(1) }

func (b *ItemBase) IsNull() bool { return b.count <= 0 }

// GetName is a port of Item::getName (final in PHP - concrete types shouldn't override this;
// they should override GetVanillaName instead).
func (b *ItemBase) GetName() string {
	if b.self.HasCustomName() {
		return b.self.GetCustomName()
	}
	return b.self.GetVanillaName()
}

func (b *ItemBase) GetVanillaName() string { return b.name }

func (b *ItemBase) GetEnchantability() int { return 1 }

// describeState is a port of Item::describeState's default (NOOP) implementation. Concrete item
// types with runtime state (e.g. Dye) override this by shadowing the method on their own type -
// since b.self is an Item interface value, the type assertion in GetStateId below resolves to
// whichever describeState is actually in the concrete type's method set (its own, if it defines
// one, otherwise this promoted default), the same "narrow-scope self-dispatch" pattern used
// throughout block/ for PHP abstract-base-with-internal-self-calls situations.
func (b *ItemBase) describeState(w runtime.DataDescriber) {}

type stateDescriber interface {
	describeState(w runtime.DataDescriber)
}

// GetStateId is a port of Item::getStateId/computeStateData, using a bit-interleave in the same
// spirit as the PHP original's morton2d_encode (see morton2DEncode below) - it's internal to this
// port rather than a byte-for-byte match of the C extension's bit layout, since nothing here
// persists a state ID across processes yet (NbtSerialize/NbtDeserialize aren't ported).
func (b *ItemBase) GetStateId() int {
	writer := runtime.NewWriter(16)
	b.self.(stateDescriber).describeState(writer)
	return morton2DEncode(b.identifier.TypeId, writer.GetValue())
}

func (b *ItemBase) GetTypeId() int { return b.identifier.TypeId }

func (b *ItemBase) GetMaxStackSize() int { return 64 }

func (b *ItemBase) GetFuelTime() int { return 0 }

// GetFuelResidue is a port of Item::getFuelResidue: clones self, then pops one item off the
// clone (discarding the popped-off stack), leaving the clone's own count reduced by one.
func (b *ItemBase) GetFuelResidue() Item {
	residue := b.self.Clone()
	residue.Pop()
	return residue
}

func (b *ItemBase) IsFireProof() bool { return false }

func (b *ItemBase) GetAttackPoints() int { return 1 }

func (b *ItemBase) GetDefensePoints() int { return 0 }

func (b *ItemBase) GetBlockToolType() block.ToolType { return block.ToolTypeNone }

func (b *ItemBase) GetBlockToolHarvestLevel() int { return 0 }

func (b *ItemBase) GetMiningEfficiency(isCorrectTool bool) float64 { return 1 }

func (b *ItemBase) GetCooldownTicks() int { return 0 }

func (b *ItemBase) GetCooldownTag() (string, bool) { return "", false }

// Equals is a port of Item::equals (PHP's $checkDamage parameter is deprecated/unused there too,
// so it's dropped here).
func (b *ItemBase) Equals(other Item, checkCompound bool) bool {
	return b.self.GetStateId() == other.GetStateId() &&
		(!checkCompound || b.self.GetNamedTag().Equals(other.GetNamedTag()))
}

func (b *ItemBase) CanStackWith(other Item) bool { return b.self.Equals(other, true) }

func (b *ItemBase) EqualsExact(other Item) bool {
	return b.self.CanStackWith(other) && b.count == other.GetCount()
}

// String is a port of Item::__toString.
func (b *ItemBase) String() string {
	s := fmt.Sprintf("Item %s (%d:%d)x%d", b.name, b.self.GetTypeId(), b.self.GetStateId(), b.count)
	if b.self.HasNamedTag() {
		root, err := nbt.NewTreeRoot(b.self.GetNamedTag(), "")
		if err != nil {
			panic(err)
		}
		encoded, err := nbt.NewLittleEndianSerializer().Write(root)
		if err != nil {
			panic(err)
		}
		s += " tags:0x" + base64.StdEncoding.EncodeToString(encoded)
	}
	return s
}

// morton2DEncode interleaves the bits of x and y (x in the low bit of each pair, y in the high
// bit) into a single value, in the same spirit as the PHP original's morton2d_encode - see
// GetStateId's doc comment for why an exact bit-layout match with the C extension isn't needed
// here.
func morton2DEncode(x, y int) int {
	result := 0
	for i := 0; i < 32; i++ {
		result |= ((x >> i) & 1) << (2 * i)
		result |= ((y >> i) & 1) << (2*i + 1)
	}
	return result
}
