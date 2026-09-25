package crafting

import (
	"fmt"

	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/network/mcpe/convert"
)

// RecipeIngredient is a port of pocketmine\crafting\RecipeIngredient.
type RecipeIngredient interface {
	fmt.Stringer
	Accepts(it item.Item) bool
}

// ExactRecipeIngredient is a port of pocketmine\crafting\ExactRecipeIngredient.
type ExactRecipeIngredient struct {
	item item.Item
}

// NewExactRecipeIngredient is a port of ExactRecipeIngredient::__construct.
func NewExactRecipeIngredient(it item.Item) (*ExactRecipeIngredient, error) {
	if it.IsNull() {
		return nil, fmt.Errorf("Recipe ingredients must not be air items")
	}
	if it.GetCount() != 1 {
		return nil, fmt.Errorf("Recipe ingredients cannot require count")
	}
	return &ExactRecipeIngredient{item: it.Clone()}, nil
}

func (r *ExactRecipeIngredient) GetItem() item.Item { return r.item.Clone() }

// Accepts is a port of ExactRecipeIngredient::accepts.
func (r *ExactRecipeIngredient) Accepts(it item.Item) bool {
	//client-side, recipe inputs can't actually require NBT
	//but on the PM side, we currently check for it if the input requires it, so we have to continue to do so for
	//the sake of consistency
	return it.GetCount() >= 1 && r.item.Equals(it, r.item.HasNamedTag())
}

func (r *ExactRecipeIngredient) String() string {
	return "ExactRecipeIngredient(" + fmt.Sprint(r.item) + ")"
}

// MetaWildcardRecipeIngredient is a port of pocketmine\crafting\MetaWildcardRecipeIngredient:
// accepts any item of the given type, whatever its meta.
type MetaWildcardRecipeIngredient struct {
	itemID string
}

func NewMetaWildcardRecipeIngredient(itemID string) *MetaWildcardRecipeIngredient {
	return &MetaWildcardRecipeIngredient{itemID: itemID}
}

func (r *MetaWildcardRecipeIngredient) GetItemId() string { return r.itemID }

// Accepts is a port of MetaWildcardRecipeIngredient::accepts.
func (r *MetaWildcardRecipeIngredient) Accepts(it item.Item) bool {
	if it.GetCount() < 1 {
		return false
	}
	name, ok := convert.ItemTypeName(it)
	return ok && name == r.itemID
}

func (r *MetaWildcardRecipeIngredient) String() string {
	return "MetaWildcardRecipeIngredient(" + r.itemID + ")"
}

// TagWildcardRecipeIngredient is a port of pocketmine\crafting\TagWildcardRecipeIngredient:
// accepts any item whose type is in the given vanilla item tag.
type TagWildcardRecipeIngredient struct {
	tagName string
}

func NewTagWildcardRecipeIngredient(tagName string) *TagWildcardRecipeIngredient {
	return &TagWildcardRecipeIngredient{tagName: tagName}
}

func (r *TagWildcardRecipeIngredient) GetTagName() string { return r.tagName }

// Accepts is a port of TagWildcardRecipeIngredient::accepts.
func (r *TagWildcardRecipeIngredient) Accepts(it item.Item) bool {
	if it.GetCount() < 1 {
		return false
	}
	name, ok := convert.ItemTypeName(it)
	return ok && bedrock.GetItemTagToIdMap().TagContainsId(r.tagName, name)
}

func (r *TagWildcardRecipeIngredient) String() string {
	return "TagWildcardRecipeIngredient(" + r.tagName + ")"
}
