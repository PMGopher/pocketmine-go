package item

import (
	"pocketmine-go/pocketmine/block"
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// ItemBlock is a port of pocketmine\item\ItemBlock: items that directly represent a block (stone,
// dirt, wood, etc.), as opposed to items merely associated with one (seeds aren't wheat crops -
// they place them).
//
// ItemIdentifier.FromBlock (needing ItemTypeIds::fromBlockTypeId, a registry mapping block type
// IDs to item type IDs, not ported) isn't used here - NewItemBlock takes an explicit
// ItemIdentifier instead of deriving one from the block the way the PHP constructor does.
type ItemBlock struct {
	ItemBase

	Block block.Behavior
}

func NewItemBlock(identifier ItemIdentifier, blk block.Behavior) *ItemBlock {
	ib := &ItemBlock{Block: blk}
	ib.Init(ib, identifier, blk.GetName())
	return ib
}

// init wires block.NewItemBlockFunc to a real ItemBlock constructor - see that var's doc comment
// in block/item.go for why this indirection exists (block can't import item directly). The item
// type ID is -blk.GetTypeId(): ItemTypeIds::fromBlockTypeId in the PHP original is exactly this
// negation, not a lookup table, so no block-type-ID-to-item-type-ID registry is needed here.
func init() {
	block.NewItemBlockFunc = func(blk block.Behavior) block.Item {
		return NewItemBlock(NewItemIdentifier(-blk.GetTypeId()), blk)
	}
}

// Clone deep-copies the wrapped block too, not just the ItemBlock's own fields.
func (i *ItemBlock) Clone() Item {
	c := *i
	c.Block = i.Block.Clone()
	c.rebind(&c)
	return &c
}

func (i *ItemBlock) describeState(w runtime.DataDescriber) { i.Block.DescribeBlockItemState(w) }

// GetBlock is a port of ItemBlock::getBlock. The PHP signature's optional $clickedFace parameter
// is dropped: it isn't read by the base implementation (only some Block::place overrides use
// face information, which callers already have independently), and nothing here calls GetBlock
// with a face yet.
func (i *ItemBlock) GetBlock() block.Behavior { return i.Block.Clone() }

func (i *ItemBlock) GetFuelTime() int { return i.Block.GetFuelTime() }

func (i *ItemBlock) IsFireProof() bool { return i.Block.IsFireProofAsItem() }

func (i *ItemBlock) GetMaxStackSize() int { return i.Block.GetMaxStackSize() }

// IsNull is a port of ItemBlock::isNull - also null if the wrapped block is air, matching the PHP
// original's air-as-null-slot special case.
func (i *ItemBlock) IsNull() bool { return i.ItemBase.IsNull() || i.Block.GetTypeId() == block.AIR }
