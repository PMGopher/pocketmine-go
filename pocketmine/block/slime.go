package block

// Slime is a port of pocketmine\block\Slime.
type Slime struct {
	Transparent
}

func NewSlime(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Slime {
	s := &Slime{Transparent{NewBlock(idInfo, name, typeInfo)}}
	s.Init(s)
	return s
}

func (s *Slime) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

func (s *Slime) GetFrictionFactor() float64 { return 0.8 } // ???

func (s *Slime) OnEntityLand(entity Entity) (float64, bool) {
	if living, ok := entity.(Living); ok && living.IsSneaking() {
		return 0, false
	}
	entity.ResetFallDistance()
	return -entity.GetMotion().Y, true
}

// TODO (from the PHP original): slime blocks should slow entities walking on them to about 0.4x
// original speed.
