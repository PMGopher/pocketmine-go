package tile

import "pocketmine-go/pocketmine/nbt"

const NameableTagCustomName = "CustomName"

// Nameable is a port of pocketmine\block\tile\Nameable.
type Nameable interface {
	GetDefaultName() string
	GetName() string
	SetName(name string)
	HasName() bool
}

// nameableShaper lets NameableComponent reach a concrete tile's GetDefaultName - the same
// self-dispatch problem and solution as CopperComponent.OnInteractCopper in the block package
// (PHP traits get copy-pasted per using class and can call an abstract method directly; Go
// embedding can't, so the caller passes self explicitly).
type nameableShaper interface {
	GetDefaultName() string
}

// NameableComponent is a port of pocketmine\block\tile\NameableTrait.
type NameableComponent struct {
	CustomName    string
	HasCustomName bool
}

func (n *NameableComponent) GetName(self nameableShaper) string {
	if n.HasCustomName {
		return n.CustomName
	}
	return self.GetDefaultName()
}

func (n *NameableComponent) SetName(name string) {
	n.HasCustomName = name != ""
	n.CustomName = name
}

func (n *NameableComponent) HasName() bool { return n.HasCustomName }

func (n *NameableComponent) AddAdditionalSpawnData(nbtTag *nbt.CompoundTag) {
	if n.HasCustomName {
		nbtTag.SetString(NameableTagCustomName, nbt.StringTag(n.CustomName))
	}
}

func (n *NameableComponent) LoadName(tag *nbt.CompoundTag) {
	if v, err := tag.GetString(NameableTagCustomName); err == nil {
		n.CustomName = string(v)
		n.HasCustomName = true
	}
}

func (n *NameableComponent) SaveName(tag *nbt.CompoundTag) {
	if n.HasCustomName {
		tag.SetString(NameableTagCustomName, nbt.StringTag(n.CustomName))
	}
}

// ApplyItemCustomName is a port of the tail end of NameableTrait::copyDataFromItem (the part
// after `parent::copyDataFromItem($item)`). Concrete tile types that embed both TileBase and
// NameableComponent must define their own CopyDataFromItem overriding both promoted versions
// (TileBase's and NameableComponent's would otherwise collide as an ambiguous selector), calling
// this after the embedded TileBase.CopyDataFromItem - e.g.:
//
//	func (x *X) CopyDataFromItem(item Item) {
//		x.TileBase.CopyDataFromItem(item)
//		x.ApplyItemCustomName(item)
//	}
func (n *NameableComponent) ApplyItemCustomName(item Item) {
	if item.HasCustomName() {
		n.SetName(item.GetCustomName())
	}
}
