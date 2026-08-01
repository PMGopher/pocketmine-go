package block

// Sculk is a port of pocketmine\block\Sculk.
type Sculk struct {
	Opaque
}

func NewSculk(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Sculk {
	s := &Sculk{Opaque{NewBlock(idInfo, name, typeInfo)}}
	s.Init(s)
	return s
}

func (s *Sculk) Clone() Behavior {
	c := *s
	c.rebind(&c)
	return &c
}

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap).
func (s *Sculk) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (s *Sculk) IsAffectedBySilkTouch() bool { return true }

func (s *Sculk) GetXpDropAmount() int { return 1 }
