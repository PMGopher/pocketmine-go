package item

import "pocketmine-go/pocketmine/math"

// renameable is the part of Entity NameTag uses.
type renameable interface {
	CanBeRenamed() bool
	SetNameTag(name string)
}

// NameTag is a port of pocketmine\item\NameTag.
type NameTag struct {
	ItemBase
}

func NewNameTag(identifier ItemIdentifier, name string) *NameTag {
	n := &NameTag{}
	n.Init(n, identifier, name)
	return n
}

func (n *NameTag) Clone() Item {
	c := *n
	c.rebind(&c)
	return &c
}

// OnInteractEntity is a port of NameTag::onInteractEntity: the entity gets the tag's custom name.
func (n *NameTag) OnInteractEntity(player Player, entity Entity, clickVector math.Vector3) bool {
	if r, ok := entity.(renameable); ok && r.CanBeRenamed() && n.HasCustomName() {
		r.SetNameTag(n.GetCustomName())
		n.Pop()
		return true
	}
	return false
}
