package block

// Magma is a port of pocketmine\block\Magma.
type Magma struct {
	Opaque
}

func NewMagma(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Magma {
	m := &Magma{Opaque{NewBlock(idInfo, name, typeInfo)}}
	m.Init(m)
	return m
}

func (m *Magma) Clone() Behavior {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *Magma) GetLightLevel() int { return 3 }

func (m *Magma) HasEntityCollision() bool { return true }

// OnEntityInside should damage non-sneaking, non-frost-walker Living entities via
// EntityDamageByBlockEvent — needs that unported concrete event subclass, Entity.Attack, and
// frost-walker/sneaking queries not on the minimal Entity/Living interfaces, so this is a no-op
// for now; it still returns true, matching the PHP original's unconditional `return true;`.
func (m *Magma) OnEntityInside(entity Entity) bool { return true }

func (m *Magma) BurnsForever() bool { return true }
