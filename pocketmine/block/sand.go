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

// OnNearbyBlockChange should replace itself with air and spawn a FallingBlock entity when
// unsupported — needs the block registry and the unported FallingBlock entity type (see
// FallableComponent's doc comment), so this is a no-op for now.
func (s *Sand) OnNearbyBlockChange() {}
