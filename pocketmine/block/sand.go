package block

// Sand is a port of pocketmine\block\Sand.
type Sand struct {
	Opaque
	FallableComponent
}

func NewSand(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Sand {
	s := &Sand{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	s.Init(s)
	return s
}

func (s *Sand) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

// OnNearbyBlockChange is FallableTrait::onNearbyBlockChange.
func (s *Sand) OnNearbyBlockChange() { FallableOnNearbyBlockChange(s.self) }
